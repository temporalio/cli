package server

import (
	uiLogger "github.com/temporalio/ui-server/v2/server/log"
	uiTag "github.com/temporalio/ui-server/v2/server/log/tag"
	serverLogger "go.temporal.io/server/common/log"
	serverTag "go.temporal.io/server/common/log/tag"
)

type UILoggerAdapter struct {
	logger serverLogger.Logger
}

func NewUILoggerAdapter(logger serverLogger.Logger) uiLogger.Logger {
	return &UILoggerAdapter{logger: logger}
}

func (l *UILoggerAdapter) Debug(msg string, tags ...uiTag.Tag) {
	l.logger.Debug(msg, convertTags(tags)...)
}

func (l *UILoggerAdapter) Info(msg string, tags ...uiTag.Tag) {
	l.logger.Info(msg, convertTags(tags)...)
}

func (l *UILoggerAdapter) Warn(msg string, tags ...uiTag.Tag) {
	l.logger.Warn(msg, convertTags(tags)...)
}

func (l *UILoggerAdapter) Error(msg string, tags ...uiTag.Tag) {
	l.logger.Error(msg, convertTags(tags)...)
}

func (l *UILoggerAdapter) DPanic(msg string, tags ...uiTag.Tag) {
	l.logger.DPanic(msg, convertTags(tags)...)
}

func (l *UILoggerAdapter) Panic(msg string, tags ...uiTag.Tag) {
	l.logger.Panic(msg, convertTags(tags)...)
}

func (l *UILoggerAdapter) Fatal(msg string, tags ...uiTag.Tag) {
	l.logger.Fatal(msg, convertTags(tags)...)
}

func convertTags(uiTags []uiTag.Tag) []serverTag.Tag {
	commonTags := make([]serverTag.Tag, len(uiTags))
	for i, t := range uiTags {
		commonTags[i] = serverTag.Tag(t)
	}
	return commonTags
}
