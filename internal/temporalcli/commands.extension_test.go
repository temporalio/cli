package temporalcli_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/imports"
)

// Go code snippets for cross-platform test extensions.
var (
	codeEchoArgs   = `fmt.Println("Args:", strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe"), strings.Join(os.Args[1:], " "))`
	codeEchoStderr = func(msg string) string {
		return fmt.Sprintf(`fmt.Fprintln(os.Stderr, %q)`, msg)
	}
	codeExit = func(code int) string {
		return fmt.Sprintf(`os.Exit(%d)`, code)
	}
	codeSleep = func(d time.Duration) string {
		return fmt.Sprintf(`time.Sleep(%d)`, d)
	}
	codeCat     = `io.Copy(os.Stdout, os.Stdin)`
	codeEchoEnv = func(name string) string {
		return fmt.Sprintf(`fmt.Println(os.Getenv(%q))`, name)
	}
)

func TestExtension_InvokesRootExtension(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo", codeEchoArgs)

	res := h.Execute("foo")

	assert.Equal(t, "Args: temporal-foo \n", res.Stdout.String())
}

func TestExtension_InvokesSubcommandExtension(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo-bar", codeEchoArgs)

	res := h.Execute("foo", "bar")

	assert.Equal(t, "Args: temporal-foo-bar \n", res.Stdout.String())
}

func TestExtension_PrefersMostSpecificExtension(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo", codeEchoArgs)
	h.createExtension("temporal-foo-bar", codeEchoArgs)

	res := h.Execute("foo", "bar")

	assert.Equal(t, "Args: temporal-foo-bar \n", res.Stdout.String())
}

func TestExtension_ConvertsDashToUnderscoreInLookup(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo-bar_baz", codeEchoArgs)

	// Dash in arg is converted to underscore when looking up extension.
	res := h.Execute("foo", "bar-baz")
	assert.Equal(t, "Args: temporal-foo-bar_baz \n", res.Stdout.String())

	// Underscore in arg stays as underscore.
	res = h.Execute("foo", "bar_baz")
	assert.Equal(t, "Args: temporal-foo-bar_baz \n", res.Stdout.String())
}

func TestExtension_DoesNotOverrideBuiltinCommand(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-workflow", codeEchoArgs)
	h.createExtension("temporal-workflow-list", codeEchoArgs)

	t.Run("root command", func(t *testing.T) {
		res := h.Execute("workflow", "--help")
		assert.Contains(t, res.Stdout.String(), "Workflow commands perform operations on Workflow Executions")
	})

	t.Run("subcommand", func(t *testing.T) {
		res := h.Execute("workflow", "list", "--help")
		assert.Contains(t, res.Stdout.String(), "List Workflow Executions")
	})
}

func TestExtension_Flags(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo", codeEchoArgs)
	h.createExtension("temporal-foo-bar", codeEchoArgs)  // should never be called
	h.createExtension("temporal-foo-json", codeEchoArgs) // should never be called
	h.createExtension("temporal-workflow-diagram", codeEchoArgs)
	h.createExtension("temporal-workflow-diagram-foo", codeEchoArgs)  // should never be called
	h.createExtension("temporal-workflow-diagram-json", codeEchoArgs) // should never be called

	cases := []struct {
		args string
		want string
		err  string
	}{
		// Root extension

		{args: "--no-json-shorthand-payloads foo", want: "temporal-foo --no-json-shorthand-payloads"}, // boolean flag
		{args: "--output json foo", want: "temporal-foo --output json"},
		{args: "--output=json foo", want: "temporal-foo --output=json"},
		{args: "-o json foo", want: "temporal-foo -o json"}, // shorthand
		{args: "-o=json foo", want: "temporal-foo -o=json"},
		{args: "--unknown-flag value foo", err: "unknown flag"}, // unknown flag before extension name
		{args: "--output invalid foo", err: "invalid argument"}, // invalid value for known flag

		{args: "foo --output json", want: "temporal-foo --output json"}, // not temporal-foo-json
		{args: "foo --output=json", want: "temporal-foo --output=json"},
		{args: "foo -o json", want: "temporal-foo -o json"},
		{args: "foo -o=json", want: "temporal-foo -o=json"},
		{args: "foo -x bar", want: "temporal-foo -x bar"},                         // not temporal-foo-x
		{args: "foo --output invalid", err: "invalid argument"},                   // invalid value for known flag
		{args: "foo arg1 -x value arg2", want: "temporal-foo arg1 -x value arg2"}, // order preserved

		// Subcommand extension

		{args: "--output json workflow diagram", want: "temporal-workflow-diagram --output json"},
		{args: "--output=json workflow diagram", want: "temporal-workflow-diagram --output=json"},
		{args: "-o json workflow diagram", want: "temporal-workflow-diagram -o json"}, // shorthand
		{args: "-o=json workflow diagram", want: "temporal-workflow-diagram -o=json"},
		{args: "--unknown-flag value workflow diagram", err: "unknown flag"},

		{args: "workflow --tls diagram", want: "temporal-workflow-diagram --tls"}, // boolean flag
		{args: "workflow --namespace my-ns diagram", want: "temporal-workflow-diagram --namespace my-ns"},
		{args: "workflow --namespace=my-ns diagram", want: "temporal-workflow-diagram --namespace=my-ns"},
		{args: "workflow -n my-ns diagram", want: "temporal-workflow-diagram -n my-ns"}, // shorthand
		{args: "workflow -n=my-ns diagram", want: "temporal-workflow-diagram -n=my-ns"},
		{args: "workflow --unknown-flag diagram", err: "unknown flag"},       // unknown flag before extension name
		{args: "workflow --output invalid diagram", err: "invalid argument"}, // invalid value for known flag

		{args: "workflow diagram --output json", want: "temporal-workflow-diagram --output json"}, // not temporal-workflow-diagram-json
		{args: "workflow diagram --output=json", want: "temporal-workflow-diagram --output=json"},
		{args: "workflow diagram -o json", want: "temporal-workflow-diagram -o json"}, // shorthand
		{args: "workflow diagram -o=json", want: "temporal-workflow-diagram -o=json"},
		{args: "workflow diagram -x foo", want: "temporal-workflow-diagram -x foo"},                         // not temporal-workflow-diagram-foo
		{args: "workflow diagram arg1 -x value arg2", want: "temporal-workflow-diagram arg1 -x value arg2"}, // order preserved
		{args: "workflow diagram foo --flag value", want: "temporal-workflow-diagram-foo --flag value"},     // nested commands
		{args: "workflow diagram --output invalid", err: "invalid argument"},                                // invalid value for known flag

		// Note: Flag aliases are already implicitly tested via other command-specific tests.
	}

	for _, c := range cases {
		res := h.Execute(strings.Split(c.args, " ")...)
		if c.err != "" {
			assert.ErrorContains(t, res.Err, c.err)
		} else {
			assert.Equal(t, "Args: "+c.want+"\n", res.Stdout.String())
			assert.NoError(t, res.Err)
		}
	}
}

func TestExtension_PassesStdin(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo", codeCat)
	h.Stdin.WriteString("hello from stdin")

	res := h.Execute("foo")

	assert.Equal(t, "hello from stdin", res.Stdout.String())
}

func TestExtension_InheritsEnvironmentVariables(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo", codeEchoEnv("TEST_EXT_VAR"))
	os.Setenv("TEST_EXT_VAR", "test_value_123")
	t.Cleanup(func() { os.Unsetenv("TEST_EXT_VAR") })

	res := h.Execute("foo")

	assert.Equal(t, "test_value_123\n", res.Stdout.String())
}

func TestExtension_RelaysStderr(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo", codeEchoStderr("stderr output"))

	res := h.Execute("foo")

	assert.Empty(t, res.Stdout.String())
	assert.Equal(t, "stderr output\n", res.Stderr.String())
}

func TestExtension_RelaysStdoutAndStderr(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo", `
		fmt.Fprintln(os.Stdout, "stdout line")
		fmt.Fprintln(os.Stderr, "stderr line")
	`)

	res := h.Execute("foo")

	assert.Equal(t, "stdout line\n", res.Stdout.String())
	assert.Equal(t, "stderr line\n", res.Stderr.String())
}

func TestExtension_FailsOnNonExecutableCommand(t *testing.T) {
	h := newExtensionHarness(t)
	// Create file without execute permission.
	path := filepath.Join(h.binDir, "temporal-foo")
	err := os.WriteFile(path, []byte("a text file"), 0644)
	require.NoError(t, err)

	res := h.Execute("foo")

	assert.Contains(t, res.Stdout.String(), "Usage:") // help text is shown
	assert.EqualError(t, res.Err, "unknown command")
}

func TestExtension_PassesThroughNonZeroExit(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo", codeEchoArgs, codeExit(42))

	res := h.Execute("foo")

	assert.Equal(t, "Args: temporal-foo \n", res.Stdout.String())
	assert.NoError(t, res.Err)
}

func TestExtension_FailsOnCommandTimeout(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo", codeSleep(10*time.Second))

	res := h.Execute("foo", "--command-timeout", "100ms")
	assert.EqualError(t, res.Err, "program interrupted")

	res = h.Execute("foo", "--command-timeout", "invalid")
	assert.ErrorContains(t, res.Err, "invalid argument \"invalid\"")
}

func TestExtension_FailsOnCommandCancellation(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo", codeSleep(10*time.Second))
	go func() {
		time.Sleep(100 * time.Millisecond)
		h.CancelContext()
	}()

	res := h.Execute("foo")

	assert.EqualError(t, res.Err, "program interrupted")
}

type extensionHarness struct {
	*CommandHarness
	binDir string
}

func newExtensionHarness(t *testing.T) *extensionHarness {
	t.Helper()

	binDir := t.TempDir()
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath)
	t.Cleanup(func() { os.Setenv("PATH", oldPath) })

	return &extensionHarness{
		CommandHarness: NewCommandHarness(t),
		binDir:         binDir,
	}
}

func (h *extensionHarness) createExtension(name string, code ...string) string {
	h.t.Helper()

	// Wrap code in main function.
	source := fmt.Sprintf("package main\n\nfunc main() {\n%s\n}\n", strings.Join(code, "\n"))

	// Run goimports to resolve imports.
	formatted, err := imports.Process("main.go", []byte(source), nil)
	require.NoError(h.t, err, "Failed to process imports for %s:\n%s", name, source)

	// Write source file.
	srcPath := filepath.Join(h.binDir, name+".go")
	require.NoError(h.t, os.WriteFile(srcPath, formatted, 0644))

	// Build executable.
	binPath := filepath.Join(h.binDir, name)
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", binPath, srcPath)
	output, err := cmd.CombinedOutput()
	require.NoError(h.t, err, "Failed to compile %s: %s\nSource:\n%s", name, output, formatted)

	return binPath
}
