package temporalcli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/temporalio/ui-server/v2/server/version"
	"go.temporal.io/server/common/headers"
	"gopkg.in/yaml.v3"
)

// Version is the value put as the default command version. This is often
// replaced at build time via ldflags.
var Version = "0.0.0-DEV"

type CommandContext struct {
	// This context is closed on interrupt
	context.Context
	Options  CommandOptions
	EnvFlags map[string]map[string]string
	EnvViper *viper.Viper

	// These values may not be available until after pre-run of main command
	Printer    Printer
	JSONOutput bool
	Logger     *slog.Logger
}

type CommandOptions struct {
	// If empty, assumed to be os.Args[1:]
	Args []string
	// If unset, defaulted to $HOME/.config/temporalio/temporal.yaml
	EnvFile string
	// If unset, attempts to extract --env from Args
	Env string
	// If true, does not do any env reading
	DisableEnv bool

	// These two default to OS values
	Stdout io.Writer
	Stderr io.Writer

	// Defaults to logging error then os.Exit(1)
	Fail func(error)
}

func NewCommandContext(ctx context.Context, options CommandOptions) (*CommandContext, context.CancelFunc, error) {
	cctx := &CommandContext{Context: ctx, Options: options}
	if err := cctx.preprocessOptions(); err != nil {
		return nil, nil, err
	}

	// Use viper for flag binding
	v := viper.New()
	v.Set(cctx.Options.Env, cctx.EnvFlags[cctx.Options.Env])
	cctx.EnvViper = v.Sub(cctx.Options.Env)

	// Setup interrupt handler
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	cctx.Context = ctx
	return cctx, stop, nil
}

func (c *CommandContext) preprocessOptions() error {
	if len(c.Options.Args) == 0 {
		c.Options.Args = os.Args[1:]
	}

	if c.Options.Stdout == nil {
		c.Options.Stdout = os.Stdout
	}
	if c.Options.Stderr == nil {
		c.Options.Stderr = os.Stderr
	}

	if !c.Options.DisableEnv {
		if c.Options.EnvFile == "" {
			// Default to --env-file, prefetched from CLI args
			for i, arg := range c.Options.Args {
				if arg == "--env-file" && i+1 < len(c.Options.Args) {
					c.Options.EnvFile = c.Options.Args[i+1]
				}
			}
			// Default to inside home dir
			if c.Options.EnvFile == "" {
				c.Options.EnvFile = defaultEnvFile("temporalio", "temporal")
			}
		}

		if c.Options.Env == "" {
			c.Options.Env = "default"
			// Default to --env, prefetched from CLI args
			for i, arg := range c.Options.Args {
				if arg == "--env" && i+1 < len(c.Options.Args) {
					c.Options.Env = c.Options.Args[i+1]
				}
			}
		}

		// Load env flags
		if c.Options.EnvFile != "" {
			var err error
			if c.EnvFlags, err = readEnvFile(c.Options.EnvFile); err != nil {
				return err
			}
		}
	}

	// Setup default fail callback
	if c.Options.Fail == nil {
		c.Options.Fail = func(err error) {
			if c.Logger != nil {
				c.Logger.Error(err.Error())
			} else {
				fmt.Fprintln(os.Stderr, err)
			}
			os.Exit(1)
		}
	}
	return nil
}

func (c *CommandContext) BindConfigFlags(flags *pflag.FlagSet) {
	if err := c.EnvViper.BindPFlags(flags); err != nil {
		panic(err)
	}
}

func (c *CommandContext) WriteEnvToFile() error {
	if c.Options.EnvFile == "" {
		return fmt.Errorf("unable to find place for env file (unknown HOME dir)")
	}
	c.Logger.Info("Writing env file", "file", c.Options.EnvFile)
	return writeEnvFile(c.Options.EnvFile, c.EnvFlags)
}

// TODO(cretz): Make it clear this logs error
func Execute(ctx context.Context, options CommandOptions) {
	// Create context and run
	cctx, cancel, err := NewCommandContext(ctx, options)
	if err == nil {
		defer cancel()
		cmd := NewTemporalCommand(cctx)
		cmd.Command.SetArgs(cctx.Options.Args)
		err = cmd.Command.ExecuteContext(cctx)
	}

	// Use failure handler, but can still return
	if err != nil {
		cctx.Options.Fail(err)
	}
}

func (c *TemporalCommand) initCommand(cctx *CommandContext) {
	c.Command.Version = fmt.Sprintf("%s (server %s) (ui %s)", Version, headers.ServerVersion, version.UIVersion)
	// Unfortunately color is a global option, so we can set in pre-run but we
	// must unset in post-run
	origNoColor := color.NoColor
	c.Command.PersistentPreRunE = func(*cobra.Command, []string) error {
		// Only override if never or always, let auto keep the value
		if c.Color.Value == "never" || c.Color.Value == "always" {
			color.NoColor = c.Color.Value == "never"
		}
		return c.preRun(cctx)
	}
	c.Command.PersistentPostRun = func(*cobra.Command, []string) {
		color.NoColor = origNoColor
	}
}

func (c *TemporalCommand) preRun(cctx *CommandContext) error {
	// Configure logger if not already on context
	if cctx.Logger == nil {
		// If level is off, make noop logger
		if c.LogLevel.Value == "off" {
			cctx.Logger = newNopLogger()
		} else {
			var level slog.Level
			if err := level.UnmarshalText([]byte(c.LogLevel.Value)); err != nil {
				return fmt.Errorf("invalid log level %q: %w", c.LogLevel.Value, err)
			}
			var handler slog.Handler
			switch c.LogFormat.Value {
			case "text":
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
				return fmt.Errorf("invalid log format %q", c.LogFormat.Value)
			}
			cctx.Logger = slog.New(handler)
		}
	}

	// Configure printer if not already on context
	if cctx.Printer == nil {
		switch c.Output.Value {
		case "text":
			opts := TextPrinterOptions{Output: cctx.Options.Stdout}
			switch c.TimeFormat.Value {
			case "iso":
				opts.FormatTime = func(t time.Time) string { return t.Format(time.RFC3339) }
			case "raw":
				opts.FormatTime = func(t time.Time) string { return fmt.Sprintf("%v", t) }
			case "relative":
				opts.FormatTime = humanize.Time
			default:
				panic("unknown time format")
			}
			cctx.Printer = NewTextPrinter(opts)
		case "json":
			cctx.Printer = NewJSONPrinter(JSONPrinterOptions{Output: cctx.Options.Stdout})
			cctx.JSONOutput = true
		default:
			return fmt.Errorf("invalid output format %q", c.Output.Value)
		}
	}
	return nil
}

// May be empty result if can't get user home dir
func defaultEnvFile(appName, configName string) string {
	// No env file if no $HOME
	if dir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(dir, ".config", appName, configName+".yaml")
	}
	return ""
}

func readEnvFile(file string) (env map[string]map[string]string, err error) {
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

func writeEnvFile(file string, env map[string]map[string]string) error {
	b, err := yaml.Marshal(map[string]any{"env": env})
	if err != nil {
		return fmt.Errorf("failed marshaling YAML: %w", err)
	}
	// Make parent directories as needed
	if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
		return fmt.Errorf("failed making env file parent dirs: %w", err)
	} else if err := os.WriteFile(file, b, 0644); err != nil {
		return fmt.Errorf("failed writing env file: %w", err)
	}
	return nil
}

func newNopLogger() *slog.Logger { return slog.New(discardLogHandler{}) }

type discardLogHandler struct{}

func (discardLogHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (discardLogHandler) Handle(context.Context, slog.Record) error { return nil }
func (d discardLogHandler) WithAttrs([]slog.Attr) slog.Handler      { return d }
func (d discardLogHandler) WithGroup(string) slog.Handler           { return d }
