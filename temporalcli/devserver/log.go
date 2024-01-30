package devserver

import (
	"context"
	"log/slog"

	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
)

type slogLogger struct{ log *slog.Logger }

var _ log.Logger

func (s slogLogger) Debug(msg string, tags ...tag.Tag) { s.Log(slog.LevelDebug, msg, tags) }
func (s slogLogger) Info(msg string, tags ...tag.Tag)  { s.Log(slog.LevelInfo, msg, tags) }
func (s slogLogger) Warn(msg string, tags ...tag.Tag)  { s.Log(slog.LevelWarn, msg, tags) }
func (s slogLogger) Error(msg string, tags ...tag.Tag) { s.Log(slog.LevelError, msg, tags) }

// Panics and fatals are just errors
func (s slogLogger) DPanic(msg string, tags ...tag.Tag) { s.Log(slog.LevelError, msg, tags) }
func (s slogLogger) Panic(msg string, tags ...tag.Tag)  { s.Log(slog.LevelError, msg, tags) }
func (s slogLogger) Fatal(msg string, tags ...tag.Tag)  { s.Log(slog.LevelError, msg, tags) }

func (s slogLogger) Log(level slog.Level, msg string, tags []tag.Tag) {
	if s.log.Enabled(context.Background(), level) {
		s.log.LogAttrs(context.Background(), level, msg, logTagsToAttrs(tags)...)
	}
}

func logTagsToAttrs(tags []tag.Tag) []slog.Attr {
	attrs := make([]slog.Attr, len(tags))
	for i, tag := range tags {
		attrs[i] = slog.Any(tag.Key(), tag.Value())
	}
	return attrs
}
