package printer_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"unicode"

	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/internal/printer"
)

type writerFunc func([]byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) {
	return f(p)
}

type failingJSONMarshaler struct {
	err error
}

func (m failingJSONMarshaler) MarshalJSON() ([]byte, error) {
	return nil, m.err
}

// TODO(cretz): Test:
// * Text printer specific fields
// * Text printer specific and non-specific fields and all sorts of table options
// * JSON printer

func TestPrinter_Text(t *testing.T) {
	type MyStruct struct {
		Foo              string
		Bar              bool
		unexportedBaz    string
		ReallyLongField  any
		Omitted          string `cli:",omit"`
		OmittedCardEmpty string `cli:",cardOmitEmpty"`
	}
	var buf bytes.Buffer
	p := printer.Printer{Output: &buf}
	// Simple struct non-table no fields set
	require.NoError(t, p.PrintStructured([]*MyStruct{
		{
			Foo:           "1",
			unexportedBaz: "2",
			ReallyLongField: struct {
				Key any `json:"key"`
			}{Key: 123},
			Omitted:          "value",
			OmittedCardEmpty: "value",
		},
		{
			Foo:             "not-a-number",
			Bar:             true,
			ReallyLongField: map[string]int{"": 0},
		},
	}, printer.StructuredOptions{}))
	// Check
	require.Equal(t, normalizeMultiline(`
  Foo               1
  Bar               false
  ReallyLongField   {"key":123}
  OmittedCardEmpty  value

  Foo              not-a-number
  Bar              true
  ReallyLongField  map[:0]`), normalizeMultiline(buf.String()))

	// TODO(cretz): Tables and more options
}

func normalizeMultiline(s string) string {
	// Split lines, trim trailing space on each (also removes \r), remove empty
	// lines, re-join
	var ret string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimRightFunc(line, unicode.IsSpace)
		// Only non-empty lines
		if line != "" {
			if ret != "" {
				ret += "\n"
			}
			ret += line
		}
	}
	return ret
}

func TestPrinter_JSON(t *testing.T) {
	var buf bytes.Buffer

	// With indentation
	p := printer.Printer{Output: &buf, JSON: true, JSONIndent: "  "}
	p.Println("should not print")
	require.NoError(t, p.PrintStructured(map[string]string{"foo": "bar"}, printer.StructuredOptions{}))
	require.Equal(t, `{
  "foo": "bar"
}
`, buf.String())

	// Without indentation
	buf.Reset()
	p = printer.Printer{Output: &buf, JSON: true}
	p.Println("should not print")
	require.NoError(t, p.PrintStructured(map[string]string{"foo": "bar"}, printer.StructuredOptions{}))
	require.Equal(t, "{\"foo\":\"bar\"}\n", buf.String())
}

func TestPrinter_PrintlnStrictErrReturnsAcknowledgementWriteFailures(t *testing.T) {
	const helperEnv = "TEMPORAL_CLI_PRINTER_STRICT_ACKNOWLEDGEMENT_EPIPE"
	if os.Getenv(helperEnv) != "" {
		p := printer.Printer{Output: writerFunc(func([]byte) (int, error) {
			return 0, syscall.EPIPE
		})}
		if err := p.PrintlnStrictErr("not printed"); err != syscall.EPIPE {
			os.Exit(10)
		}
		_, _ = os.Stderr.WriteString("strict acknowledgement returned EPIPE\n")
		return
	}

	t.Run("returns the original EPIPE without exiting", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run=^TestPrinter_PrintlnStrictErrReturnsAcknowledgementWriteFailures$")
		cmd.Env = append(os.Environ(), helperEnv+"=1")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err)
		require.Contains(t, string(output), "strict acknowledgement returned EPIPE")
	})

	t.Run("returns short write", func(t *testing.T) {
		p := printer.Printer{Output: writerFunc(func(p []byte) (int, error) {
			return len(p) - 1, nil
		})}
		require.ErrorIs(t, p.PrintlnStrictErr("not printed"), io.ErrShortWrite)
	})

	t.Run("is silent in JSON", func(t *testing.T) {
		var output bytes.Buffer
		p := printer.Printer{Output: &output, JSON: true}
		require.NoError(t, p.PrintlnStrictErr("not printed"))
		require.Empty(t, output.String())
	})
}

func TestPrinter_PrintlnStrictErrReturnsEPIPEFromRealStdoutPipe(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows pipes do not generate SIGPIPE")
	}
	const (
		helperEnv      = "TEMPORAL_CLI_PRINTER_STRICT_REAL_STDOUT_PIPE"
		helperExitCode = 23
		helperMarker   = "strict real stdout pipe returned EPIPE\n"
	)
	if os.Getenv(helperEnv) != "" {
		p := printer.Printer{Output: os.Stdout}
		if err := p.PrintlnStrictErr("not printed"); !errors.Is(err, syscall.EPIPE) {
			_, _ = os.Stderr.WriteString("strict real stdout pipe did not return EPIPE\n")
			os.Exit(24)
		}
		_, _ = os.Stderr.WriteString(helperMarker)
		os.Exit(helperExitCode)
	}

	pipeReader, pipeWriter, err := os.Pipe()
	require.NoError(t, err)
	require.NoError(t, pipeReader.Close())
	t.Cleanup(func() { _ = pipeWriter.Close() })

	cmd := exec.Command(os.Args[0], "-test.run=^TestPrinter_PrintlnStrictErrReturnsEPIPEFromRealStdoutPipe$")
	cmd.Env = append(os.Environ(), helperEnv+"=1")
	cmd.Stdout = pipeWriter
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	require.NoError(t, pipeWriter.Close())

	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, helperExitCode, exitErr.ExitCode(), "stderr: %s", stderr.String())
	require.Contains(t, stderr.String(), helperMarker)
}

func TestPrinter_StartListErrReturnsWriteFailure(t *testing.T) {
	wantErr := errors.New("write failed")
	p := printer.Printer{
		Output: writerFunc(func([]byte) (int, error) {
			return 0, wantErr
		}),
		JSON:       true,
		JSONIndent: "  ",
	}

	require.ErrorIs(t, p.StartListErr(), wantErr)
}

func TestPrinter_StartListErrCanStartNewListAfterWriteFailure(t *testing.T) {
	wantErr := errors.New("write failed")
	p := printer.Printer{
		Output: writerFunc(func([]byte) (int, error) {
			return 0, wantErr
		}),
		JSON:       true,
		JSONIndent: "  ",
	}
	require.ErrorIs(t, p.StartListErr(), wantErr)

	var output bytes.Buffer
	p.Output = &output
	require.NoError(t, p.StartListErr())
	require.NoError(t, p.PrintStructured(map[string]string{"key": "value"}, printer.StructuredOptions{}))
	require.NoError(t, p.EndListErr())
	require.Equal(t, `[
{
  "key": "value"
}
]
`, output.String())
}

func TestPrinter_EndListErrReturnsWriteFailure(t *testing.T) {
	wantErr := errors.New("write failed")
	writes := 0
	p := printer.Printer{
		Output: writerFunc(func(p []byte) (int, error) {
			writes++
			if writes == 1 {
				return len(p), nil
			}
			return 0, wantErr
		}),
		JSON:       true,
		JSONIndent: "  ",
	}
	require.NoError(t, p.StartListErr())

	require.ErrorIs(t, p.EndListErr(), wantErr)
}

func TestPrinter_EndListErrResetsListStateAfterWriteFailure(t *testing.T) {
	wantErr := errors.New("write failed")
	p := printer.Printer{
		Output:     &bytes.Buffer{},
		JSON:       true,
		JSONIndent: "  ",
	}
	require.NoError(t, p.StartListErr())
	p.Output = writerFunc(func([]byte) (int, error) {
		return 0, wantErr
	})
	require.ErrorIs(t, p.EndListErr(), wantErr)
	p.Output = &bytes.Buffer{}

	require.NoError(t, p.StartListErr())
}

func TestPrinter_VoidListBoundariesRetainWriteFailureBehavior(t *testing.T) {
	wantErr := errors.New("write failed")
	failingOutput := writerFunc(func([]byte) (int, error) {
		return 0, wantErr
	})

	t.Run("start", func(t *testing.T) {
		p := printer.Printer{Output: failingOutput, JSON: true, JSONIndent: "  "}
		require.NotPanics(t, p.StartList)
	})

	t.Run("end", func(t *testing.T) {
		p := printer.Printer{Output: &bytes.Buffer{}, JSON: true, JSONIndent: "  "}
		p.StartList()
		p.Output = failingOutput
		require.NotPanics(t, p.EndList)
	})

	t.Run("println short write", func(t *testing.T) {
		p := printer.Printer{Output: writerFunc(func(p []byte) (int, error) {
			return len(p) - 1, nil
		})}
		require.NotPanics(t, func() { p.Println("partially printed") })
	})
}

func TestPrinter_PrintStructuredTextRetainsWriteFailureBehavior(t *testing.T) {
	wantErr := errors.New("write failed")
	p := printer.Printer{Output: writerFunc(func([]byte) (int, error) {
		return 0, wantErr
	})}

	require.PanicsWithValue(t, wantErr, func() {
		_ = p.PrintStructured(struct{ Value string }{Value: "not printed"}, printer.StructuredOptions{})
	})
}

func TestPrinter_PrintStructuredTextRetainsSerializationFailureBehavior(t *testing.T) {
	wantErr := errors.New("serialization failed")
	var buf bytes.Buffer
	p := printer.Printer{Output: &buf}

	require.NoError(t, p.PrintStructured(struct {
		Value failingJSONMarshaler
	}{Value: failingJSONMarshaler{err: wantErr}}, printer.StructuredOptions{}))
	require.Contains(t, buf.String(), "<failed converting to string:")
	require.Contains(t, buf.String(), "serialization failed>")
}

func TestPrinter_PrintStructuredTextRetainsDataErrorBehavior(t *testing.T) {
	p := printer.Printer{Output: &bytes.Buffer{}}

	err := p.PrintStructured(map[string]string{"key": "value"}, printer.StructuredOptions{})

	require.ErrorContains(t, err, "cannot derive fields from map")
}

func TestPrinter_PrintStructuredTableIterRetainsWriteFailureBehavior(t *testing.T) {
	wantErr := errors.New("write failed")
	p := printer.Printer{Output: writerFunc(func([]byte) (int, error) {
		return 0, wantErr
	})}

	require.PanicsWithValue(t, wantErr, func() {
		_ = p.PrintStructuredTableIter(
			reflect.TypeOf(struct{ Value string }{}),
			nil,
			printer.StructuredOptions{Table: &printer.TableOptions{}},
		)
	})
}

func TestPrinter_PrintStructuredErrTextReturnsWriteFailure(t *testing.T) {
	wantErr := errors.New("write failed")
	p := printer.Printer{Output: writerFunc(func([]byte) (int, error) {
		return 0, wantErr
	})}

	err := p.PrintStructuredErr(
		struct{ Value string }{Value: "not printed"},
		printer.StructuredOptions{},
	)

	require.ErrorIs(t, err, wantErr)
}

func TestPrinter_PrintStructuredErrTextReturnsSerializationFailure(t *testing.T) {
	wantErr := errors.New("serialization failed")
	p := printer.Printer{Output: &bytes.Buffer{}}

	err := p.PrintStructuredErr(struct {
		Value failingJSONMarshaler
	}{Value: failingJSONMarshaler{err: wantErr}}, printer.StructuredOptions{})

	require.ErrorIs(t, err, wantErr)
}

func TestPrinter_PrintStructuredJSONRetainsShortWriteBehavior(t *testing.T) {
	p := printer.Printer{
		Output: writerFunc(func(p []byte) (int, error) {
			return len(p) - 1, nil
		}),
		JSON: true,
	}

	require.NoError(t, p.PrintStructured(map[string]string{"key": "value"}, printer.StructuredOptions{}))
}

func TestPrinter_PrintStructuredErrReturnsShortWrite(t *testing.T) {
	p := printer.Printer{
		Output: writerFunc(func(p []byte) (int, error) {
			return len(p) - 1, nil
		}),
		JSON: true,
	}

	require.ErrorIs(t, p.PrintStructuredErr(map[string]string{"key": "value"}, printer.StructuredOptions{}), io.ErrShortWrite)
}

func TestPrinter_ErrorReturningMethodsExitSuccessfullyOnBrokenPipe(t *testing.T) {
	const helperEnv = "TEMPORAL_CLI_PRINTER_BROKEN_PIPE_METHOD"
	if method := os.Getenv(helperEnv); method != "" {
		brokenPipeOutput := writerFunc(func([]byte) (int, error) {
			return 0, syscall.EPIPE
		})
		var err error
		switch method {
		case "start-list":
			p := printer.Printer{Output: brokenPipeOutput, JSON: true, JSONIndent: "  "}
			err = p.StartListErr()
		case "print-structured":
			p := printer.Printer{Output: brokenPipeOutput, JSON: true}
			err = p.PrintStructuredErr(map[string]string{"key": "value"}, printer.StructuredOptions{})
		case "end-list":
			p := printer.Printer{Output: &bytes.Buffer{}, JSON: true, JSONIndent: "  "}
			require.NoError(t, p.StartListErr())
			p.Output = brokenPipeOutput
			err = p.EndListErr()
		default:
			os.Exit(12)
		}
		if err != nil {
			os.Exit(10)
		}
		os.Exit(11)
	}

	for _, method := range []string{"start-list", "print-structured", "end-list"} {
		t.Run(method, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestPrinter_ErrorReturningMethodsExitSuccessfullyOnBrokenPipe$")
			cmd.Env = append(os.Environ(), helperEnv+"="+method)
			require.NoError(t, cmd.Run())
		})
	}
}

func TestPrinter_JSONList(t *testing.T) {
	var buf bytes.Buffer

	// With indentation
	p := printer.Printer{Output: &buf, JSON: true, JSONIndent: "  "}
	p.StartList()
	p.Println("should not print")
	require.NoError(t, p.PrintStructured(map[string]string{"foo": "bar"}, printer.StructuredOptions{}))
	require.NoError(t, p.PrintStructured(map[string]string{"baz": "qux"}, printer.StructuredOptions{}))
	p.EndList()
	require.Equal(t, `[
{
  "foo": "bar"
},
{
  "baz": "qux"
}
]
`, buf.String())

	// Without indentation
	buf.Reset()
	p = printer.Printer{Output: &buf, JSON: true}
	p.StartList()
	p.Println("should not print")
	require.NoError(t, p.PrintStructured(map[string]string{"foo": "bar"}, printer.StructuredOptions{}))
	require.NoError(t, p.PrintStructured(map[string]string{"baz": "qux"}, printer.StructuredOptions{}))
	p.EndList()
	require.Equal(t, "{\"foo\":\"bar\"}\n{\"baz\":\"qux\"}\n", buf.String())

	// Empty with indentation
	buf.Reset()
	p = printer.Printer{Output: &buf, JSON: true, JSONIndent: "  "}
	p.StartList()
	p.Println("should not print")
	p.EndList()
	require.Equal(t, "[\n]\n", buf.String())

	// Empty without indentation
	buf.Reset()
	p = printer.Printer{Output: &buf, JSON: true}
	p.StartList()
	p.Println("should not print")
	p.EndList()
	require.Equal(t, "", buf.String())
}

// Asserts the printer package don't panic if the CLI is run without a STDOUT.
// This is a tricky thing to validate, as it must be done in a subprocess and as
// `go test` has its own internal fix for improper STDOUT. This was fixed in
// Go 1.22, but keeping this here as a regression test.
// See https://github.com/temporalio/cli/issues/544.
func TestPrinter_NoPanicIfNoStdout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipped on Windows")
		return
	}

	goPath, err := exec.LookPath("go")
	if err != nil {
		t.Fatalf("Error finding go executable: %v", err)
	}
	// Don't use exec.Command here, as it silently replaces nil file descriptors
	// with /dev/null on the parent side. We specifically want to test what
	// happens when stdout is nil.
	p, err := os.StartProcess(
		goPath,
		[]string{"go", "run", "./test/main.go"},
		&os.ProcAttr{
			Files: []*os.File{os.Stdin, nil, os.Stderr},
		},
	)
	if err != nil {
		t.Fatalf("Error running command: %v", err)
	}
	state, err := p.Wait()
	if err != nil {
		t.Fatalf("Error running command: %v", err)
	}
	if state.ExitCode() != 0 {
		t.Fatalf("Error running command; exit code = %d", state.ExitCode())
	}
}
