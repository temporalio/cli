package cliext

import (
	"context"
	"fmt"
	"io"
	"log/slog"
)

// NewLogger creates a new slog.Logger from CommonOptions.
// The output is written to the provided writer (typically stderr).
// Returns a nop logger if LogLevel is "never".
func NewLogger(opts CommonOptions, w io.Writer) (*slog.Logger, error) {
	// If level is never, make noop logger
	if opts.LogLevel.Value == "never" {
		return newNopLogger(), nil
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(opts.LogLevel.Value)); err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", opts.LogLevel.Value, err)
	}

	var handler slog.Handler
	switch opts.LogFormat.Value {
	// We have a "pretty" alias for compatibility
	case "", "text", "pretty":
		handler = slog.NewTextHandler(w, &slog.HandlerOptions{
			Level: level,
			// Remove the TZ from timestamps
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey && a.Value.Kind() == slog.KindTime {
					a.Value = slog.StringValue(a.Value.Time().Format("2006-01-02T15:04:05.000"))
				}
				return a
			},
		})
	case "json":
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})
	default:
		return nil, fmt.Errorf("invalid log format %q", opts.LogFormat.Value)
	}

	return slog.New(handler), nil
}

func newNopLogger() *slog.Logger {
	return slog.New(discardHandler{})
}

type discardHandler struct{}

func (discardHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (discardHandler) Handle(context.Context, slog.Record) error { return nil }
func (d discardHandler) WithAttrs([]slog.Attr) slog.Handler      { return d }
func (d discardHandler) WithGroup(string) slog.Handler           { return d }
