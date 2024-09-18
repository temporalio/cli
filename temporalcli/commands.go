package temporalcli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/temporalio/cli/temporalcli/internal/printer"
	"github.com/temporalio/ui-server/v2/server/version"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/failure/v1"
	"go.temporal.io/api/temporalproto"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/server/common/headers"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"
)

// Version is the value put as the default command version. This is often
// replaced at build time via ldflags.
var Version = "0.0.0-DEV"

type CommandContext struct {
	// This context is closed on interrupt
	context.Context
	Options          CommandOptions
	EnvConfigValues  map[string]map[string]string
	FlagsWithEnvVars []*pflag.Flag

	// These values may not be available until after pre-run of main command
	Printer               *printer.Printer
	Logger                *slog.Logger
	JSONOutput            bool
	JSONShorthandPayloads bool

	// Is set to true if any command actually started running. This is a hack to workaround the fact
	// that cobra does not properly exit nonzero if an unknown command/subcommand is given.
	ActuallyRanCommand bool
}

type CommandOptions struct {
	// If empty, assumed to be os.Args[1:]
	Args []string
	// If unset, defaulted to $HOME/.config/temporalio/temporal.yaml
	EnvConfigFile string
	// If unset, attempts to extract --env from Args (which defaults to "default")
	EnvConfigName string
	// If true, does not do any env config reading
	DisableEnvConfig bool
	// If nil, os.LookupEnv is used. This is for environment variables and not
	// related to env config stuff above.
	LookupEnv func(string) (string, bool)

	// These three fields below default to OS values
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	// Defaults to logging error then os.Exit(1)
	Fail func(error)

	AdditionalClientGRPCDialOptions []grpc.DialOption
}

// NewCommandContext creates a CommandContext for use by the rest of the CLI.
// Among other things, this parses the env config file and modifies
// options/flags according to the parameters set there.
//
// A CommandContext and CancelFunc are always returned, even in the event of an
// error; this is so the CommandContext can be used to print an appropriate
// error message.
func NewCommandContext(ctx context.Context, options CommandOptions) (*CommandContext, context.CancelFunc, error) {
	cctx := &CommandContext{Context: ctx, Options: options}
	if err := cctx.preprocessOptions(); err != nil {
		return cctx, func() {}, err
	}

	// Setup interrupt handler
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	cctx.Context = ctx
	return cctx, stop, nil
}

const temporalEnv = "TEMPORAL_ENV"

func (c *CommandContext) preprocessOptions() error {
	if len(c.Options.Args) == 0 {
		c.Options.Args = os.Args[1:]
	}
	if c.Options.LookupEnv == nil {
		c.Options.LookupEnv = os.LookupEnv
	}

	if c.Options.Stdin == nil {
		c.Options.Stdin = os.Stdin
	}
	if c.Options.Stdout == nil {
		c.Options.Stdout = os.Stdout
	}
	if c.Options.Stderr == nil {
		c.Options.Stderr = os.Stderr
	}

	// Setup default fail callback
	if c.Options.Fail == nil {
		c.Options.Fail = func(err error) {
			// If context is closed, say that the program was interrupted and ignore
			// the actual error
			if c.Err() != nil {
				err = fmt.Errorf("program interrupted")
			}
			if c.Logger != nil {
				c.Logger.Error(err.Error())
			} else {
				fmt.Fprintln(os.Stderr, err)
			}
			os.Exit(1)
		}
	}

	// Update options according to the env file. MUST BE DONE LAST.
	//
	// Why last?  Callers need the CommandContext to be usable no matter what,
	// because they rely on it to print errors even if env parsing fails.  In
	// that situation, we will return both the CommandContext AND an error.
	if !c.Options.DisableEnvConfig {
		if c.Options.EnvConfigFile == "" {
			// Default to --env-file, prefetched from CLI args
			for i, arg := range c.Options.Args {
				if arg == "--env-file" && i+1 < len(c.Options.Args) {
					c.Options.EnvConfigFile = c.Options.Args[i+1]
				}
			}
			// Default to inside home dir
			if c.Options.EnvConfigFile == "" {
				c.Options.EnvConfigFile = defaultEnvConfigFile("temporalio", "temporal")
			}
		}

		if c.Options.EnvConfigName == "" {
			c.Options.EnvConfigName = "default"
			if envVal, ok := c.Options.LookupEnv(temporalEnv); ok {
				c.Options.EnvConfigName = envVal
			}
			// Default to --env, prefetched from CLI args
			for i, arg := range c.Options.Args {
				if arg == "--env" && i+1 < len(c.Options.Args) {
					c.Options.EnvConfigName = c.Options.Args[i+1]
				}
			}
		}

		// Load env flags
		if c.Options.EnvConfigFile != "" {
			var err error
			if c.EnvConfigValues, err = readEnvConfigFile(c.Options.EnvConfigFile); err != nil {
				return err
			}
		}
	}

	return nil
}

const flagEnvVarAnnotation = "__temporal_env_var"

func (c *CommandContext) BindFlagEnvVar(flag *pflag.Flag, envVar string) {
	if flag.Annotations == nil {
		flag.Annotations = map[string][]string{}
	}
	flag.Annotations[flagEnvVarAnnotation] = []string{envVar}
	c.FlagsWithEnvVars = append(c.FlagsWithEnvVars, flag)
}

func (c *CommandContext) WriteEnvConfigToFile() error {
	if c.Options.EnvConfigFile == "" {
		return fmt.Errorf("unable to find place for env file (unknown HOME dir)")
	}
	c.Logger.Info("Writing env file", "file", c.Options.EnvConfigFile)
	return writeEnvConfigFile(c.Options.EnvConfigFile, c.EnvConfigValues)
}

func (c *CommandContext) MarshalFriendlyJSONPayloads(m *common.Payloads) (json.RawMessage, error) {
	if m == nil {
		return []byte("null"), nil
	}
	// Use one if there's one, otherwise just serialize whole thing
	if p := m.GetPayloads(); len(p) == 1 {
		return c.MarshalProtoJSON(p[0])
	}
	return c.MarshalProtoJSON(m)
}

// Starts with newline
func (c *CommandContext) MarshalFriendlyFailureBodyText(f *failure.Failure, indent string) (s string) {
	for f != nil {
		s += "\n" + indent + "Message: " + f.Message
		if f.StackTrace != "" {
			s += "\n" + indent + "StackTrace:\n" + indent + "    " +
				strings.Join(strings.Split(f.StackTrace, "\n"), "\n"+indent+"    ")
		}
		if f = f.Cause; f != nil {
			s += "\n" + indent + "Cause:"
			indent += "    "
		}
	}
	return
}

// Takes payload shorthand into account, can use
// MarshalProtoJSONNoPayloadShorthand if needed
func (c *CommandContext) MarshalProtoJSON(m proto.Message) ([]byte, error) {
	return c.MarshalProtoJSONWithOptions(m, c.JSONShorthandPayloads)
}

func (c *CommandContext) MarshalProtoJSONWithOptions(m proto.Message, jsonShorthandPayloads bool) ([]byte, error) {
	opts := temporalproto.CustomJSONMarshalOptions{Indent: c.Printer.JSONIndent}
	if jsonShorthandPayloads {
		opts.Metadata = map[string]any{common.EnablePayloadShorthandMetadataKey: true}
	}
	return opts.Marshal(m)
}

func (c *CommandContext) UnmarshalProtoJSON(b []byte, m proto.Message) error {
	return UnmarshalProtoJSONWithOptions(b, m, c.JSONShorthandPayloads)
}

func UnmarshalProtoJSONWithOptions(b []byte, m proto.Message, jsonShorthandPayloads bool) error {
	opts := temporalproto.CustomJSONUnmarshalOptions{DiscardUnknown: true}
	if jsonShorthandPayloads {
		opts.Metadata = map[string]any{common.EnablePayloadShorthandMetadataKey: true}
	}
	return opts.Unmarshal(b, m)
}

// Set flag values from environment file & variables. Returns a callback to log anything interesting
// since logging will not yet be initialized when this runs.
func (c *CommandContext) populateFlagsFromEnv(flags *pflag.FlagSet) (func(*slog.Logger), error) {
	if flags == nil {
		return func(logger *slog.Logger) {}, nil
	}
	var logCalls []func(*slog.Logger)
	var flagErr error
	flags.VisitAll(func(flag *pflag.Flag) {
		// If the flag was already changed by the user, we don't overwrite
		if flagErr != nil || flag.Changed {
			return
		}
		// Env config first, then environ
		if v, ok := c.EnvConfigValues[c.Options.EnvConfigName][flag.Name]; ok {
			if err := flag.Value.Set(v); err != nil {
				flagErr = fmt.Errorf("failed setting flag %v from config with value %v: %w", flag.Name, v, err)
				return
			}
			flag.Changed = true
		}
		if anns := flag.Annotations[flagEnvVarAnnotation]; len(anns) == 1 {
			if envVal, ok := c.Options.LookupEnv(anns[0]); ok {
				if err := flag.Value.Set(envVal); err != nil {
					flagErr = fmt.Errorf("failed setting flag %v with env name %v and value %v: %w",
						flag.Name, anns[0], envVal, err)
					return
				}
				if flag.Changed {
					logCalls = append(logCalls, func(l *slog.Logger) {
						l.Info("Env var overrode --env setting", "env_var", anns[0], "flag", flag.Name)
					})
				}
				flag.Changed = true
			}
		}
	})
	logFn := func(logger *slog.Logger) {
		for _, call := range logCalls {
			call(logger)
		}
	}
	return logFn, flagErr
}

// Returns error if JSON output enabled
func (c *CommandContext) promptYes(message string, autoConfirm bool) (bool, error) {
	if c.JSONOutput && !autoConfirm {
		return false, fmt.Errorf("must bypass prompts when using JSON output")
	}
	c.Printer.Print(message, " ")
	if autoConfirm {
		c.Printer.Println("yes")
		return true, nil
	}
	line, _ := bufio.NewReader(c.Options.Stdin).ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes", nil
}

// Returns error if JSON output enabled
func (c *CommandContext) promptString(message string, expected string, autoConfirm bool) (bool, error) {
	if c.JSONOutput && !autoConfirm {
		return false, fmt.Errorf("must bypass prompts when using JSON output")
	}
	c.Printer.Print(message, " ")
	if autoConfirm {
		c.Printer.Println(expected)
		return true, nil
	}
	line, _ := bufio.NewReader(c.Options.Stdin).ReadString('\n')
	line = strings.TrimSpace(line)
	return line == expected, nil
}

// Execute runs the Temporal CLI with the given context and options. This
// intentionally does not return an error but rather invokes Fail on the
// options.
func Execute(ctx context.Context, options CommandOptions) {
	// Create context and run. We always get a context and cancel func back even
	// if an error was returned. This is so we can use the context to print an
	// error message using the appropriate Fail() method, regardless of why the
	// failure occurred.
	//
	// (In most cases, an error here likely means a problem with the user's env
	// config file, or some other issue in their environment.)
	cctx, cancel, err := NewCommandContext(ctx, options)
	defer cancel()

	if err == nil {
		// We have a context; let's actually run the command.
		cmd := NewTemporalCommand(cctx)
		cmd.Command.SetArgs(cctx.Options.Args)
		err = cmd.Command.ExecuteContext(cctx)
	}

	if err != nil {
		// Either we failed to create the context, OR the command itself failed.
		// Either way, we need to print an error message.
		cctx.Options.Fail(err)
	}

	// If no command ever actually got run, exit nonzero with an error.  This is
	// an ugly hack to make sure that iff the user explicitly asked for help, we
	// exit with a zero error code.  (The other situation in which help is
	// printed is when the user invokes an unknown command--we still want a
	// non-zero exit in that case.)  We should revisit this if/when the
	// following Cobra issues get fixed:
	//
	// - https://github.com/spf13/cobra/issues/1156
	// - https://github.com/spf13/cobra/issues/706
	if !cctx.ActuallyRanCommand {
		zeroExitArgs := []string{"--help", "-h", "--version", "-v", "help"}
		if slices.ContainsFunc(cctx.Options.Args, func(a string) bool {
			return slices.Contains(zeroExitArgs, a)
		}) {
			return
		}
		cctx.Options.Fail(fmt.Errorf("unknown command"))
	}
}

func (c *TemporalCommand) initCommand(cctx *CommandContext) {
	c.Command.Version = VersionString()
	// Unfortunately color is a global option, so we can set in pre-run but we
	// must unset in post-run
	origNoColor := color.NoColor
	c.Command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Populate environ. We will make the error return here which will cause
		// usage to be printed.
		logCalls, err := cctx.populateFlagsFromEnv(cmd.Flags())
		if err != nil {
			return err
		}

		// Default color.NoColor global is equivalent to "auto" so only override if
		// never or always
		if c.Color.Value == "never" || c.Color.Value == "always" {
			color.NoColor = c.Color.Value == "never"
		}

		res := c.preRun(cctx)

		logCalls(cctx.Logger)

		// Always disable color if JSON output is on (must be run after preRun so JSONOutput is set)
		if cctx.JSONOutput {
			color.NoColor = true
		}
		cctx.ActuallyRanCommand = true

		if cctx.Options.EnvConfigName != "default" {
			if _, ok := cctx.EnvConfigValues[cctx.Options.EnvConfigName]; !ok {
				if _, ok := cmd.Annotations["ignoresMissingEnv"]; !ok {
					// stfu about help output
					cmd.SilenceErrors = true
					cmd.SilenceUsage = true
					return fmt.Errorf("environment %q not found", cctx.Options.EnvConfigName)
				}
			}
		}
		return res
	}
	c.Command.PersistentPostRun = func(*cobra.Command, []string) {
		color.NoColor = origNoColor
	}
}

var buildInfo string

func VersionString() string {
	// To add build-time information to the version string, use
	// go build -ldflags "-X github.com/temporalio/cli/temporalcli.buildInfo=<MyString>"
	var bi = buildInfo
	if bi != "" {
		bi = fmt.Sprintf(", %s", bi)
	}
	return fmt.Sprintf("%s (Server %s, UI %s%s)", Version, headers.ServerVersion, version.UIVersion, bi)
}

func (c *TemporalCommand) preRun(cctx *CommandContext) error {
	// Configure logger if not already on context
	if cctx.Logger == nil {
		// If level is never, make noop logger
		if c.LogLevel.Value == "never" {
			cctx.Logger = newNopLogger()
		} else {
			var level slog.Level
			if err := level.UnmarshalText([]byte(c.LogLevel.Value)); err != nil {
				return fmt.Errorf("invalid log level %q: %w", c.LogLevel.Value, err)
			}
			var handler slog.Handler
			switch c.LogFormat.Value {
			// We have a "pretty" alias for compatibility
			case "text", "pretty":
				handler = slog.NewTextHandler(cctx.Options.Stderr, &slog.HandlerOptions{
					Level: level,
					// Remove the TZ
					ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
						if a.Key == slog.TimeKey && a.Value.Kind() == slog.KindTime {
							a.Value = slog.StringValue(a.Value.Time().Format("2006-01-02T15:04:05.000"))
						}
						return a
					},
				})
			case "json":
				handler = slog.NewJSONHandler(cctx.Options.Stderr, &slog.HandlerOptions{Level: level})
			default:
				return fmt.Errorf("unreachable: invalid log format %q", c.LogFormat)
			}
			cctx.Logger = slog.New(handler)
		}
	}

	// Configure printer if not already on context
	cctx.JSONOutput = c.Output.Value == "json" || c.Output.Value == "jsonl"
	// Only indent JSON if not jsonl
	var jsonIndent string
	if c.Output.Value == "json" {
		jsonIndent = "  "
	}
	if cctx.Printer == nil {
		printerOutput := cctx.Options.Stdout
		// Disable printer by making writer noop if "none" chosen
		if c.Output.Value == "none" {
			printerOutput = nopWriter{}
		}
		cctx.Printer = &printer.Printer{
			Output:               printerOutput,
			JSON:                 cctx.JSONOutput,
			JSONIndent:           jsonIndent,
			JSONPayloadShorthand: !c.NoJsonShorthandPayloads,
		}
		switch c.TimeFormat.Value {
		case "iso":
			cctx.Printer.FormatTime = func(t time.Time) string { return t.Format(time.RFC3339) }
		case "raw":
			cctx.Printer.FormatTime = func(t time.Time) string { return fmt.Sprintf("%v", t) }
		case "relative":
			cctx.Printer.FormatTime = humanize.Time
		default:
			return fmt.Errorf("invalid time format %q", c.TimeFormat.Value)
		}
	}
	cctx.JSONShorthandPayloads = !c.NoJsonShorthandPayloads
	return nil
}

// May be empty result if can't get user home dir
func defaultEnvConfigFile(appName, configName string) string {
	// No env file if no $HOME
	if dir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(dir, ".config", appName, configName+".yaml")
	}
	return ""
}

func readEnvConfigFile(file string) (env map[string]map[string]string, err error) {
	b, err := os.ReadFile(file)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed reading env file: %w", err)
	}
	var m map[string]map[string]map[string]string
	if err := yaml.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("failed unmarshalling env YAML: %w", err)
	}
	return m["env"], nil
}

func writeEnvConfigFile(file string, env map[string]map[string]string) error {
	b, err := yaml.Marshal(map[string]any{"env": env})
	if err != nil {
		return fmt.Errorf("failed marshaling YAML: %w", err)
	}
	// Make parent directories as needed
	if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
		return fmt.Errorf("failed making env file parent dirs: %w", err)
	} else if err := os.WriteFile(file, b, 0600); err != nil {
		return fmt.Errorf("failed writing env file: %w", err)
	}
	return nil
}

func aliasNormalizer(aliases map[string]string) func(f *pflag.FlagSet, name string) pflag.NormalizedName {
	return func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		if actual := aliases[name]; actual != "" {
			name = actual
		}
		return pflag.NormalizedName(name)
	}
}

func newNopLogger() *slog.Logger { return slog.New(discardLogHandler{}) }

type discardLogHandler struct{}

func (discardLogHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (discardLogHandler) Handle(context.Context, slog.Record) error { return nil }
func (d discardLogHandler) WithAttrs([]slog.Attr) slog.Handler      { return d }
func (d discardLogHandler) WithGroup(string) slog.Handler           { return d }

func timestampToTime(t *timestamppb.Timestamp) time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.AsTime()
}

type nopWriter struct{}

func (nopWriter) Write(b []byte) (int, error) { return len(b), nil }

type structuredError struct {
	Message string `json:"message"`
	Type    string `json:"type,omitempty"`
	Details any    `json:"details,omitempty"`
}

func fromApplicationError(err *temporal.ApplicationError) (*structuredError, error) {
	var deets any
	if err := err.Details(&deets); err != nil && !errors.Is(err, temporal.ErrNoData) {
		return nil, err
	}
	return &structuredError{
		Message: err.Error(),
		Type:    err.Type(),
		Details: deets,
	}, nil
}
