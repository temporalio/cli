package client

import (
	"fmt"

	sdklog "go.temporal.io/sdk/log"

	log "go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
)

const extraSkipForSdkLogger = 1

type SdkLogger struct {
	logger log.Logger
}

var _ sdklog.Logger = (*SdkLogger)(nil)

func NewSdkLogger(logger log.Logger) *SdkLogger {
	if sl, ok := logger.(log.SkipLogger); ok {
		logger = sl.Skip(extraSkipForSdkLogger)
	}

	return &SdkLogger{
		logger: logger,
	}
}

func (l *SdkLogger) tags(keyvals []interface{}) []tag.Tag {
	if len(keyvals)%2 != 0 {
		return []tag.Tag{tag.Error(fmt.Errorf("odd number of keyvals pairs: %v", keyvals))}
	}

	var tags []tag.Tag
	for i := 0; i < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			key = fmt.Sprintf("%v", keyvals[i])
		}
		tags = append(tags, tag.NewAnyTag(key, keyvals[i+1]))
	}

	return tags
}

func (l *SdkLogger) Debug(msg string, keyvals ...interface{}) {
	l.logger.Debug(msg, l.tags(keyvals)...)
}

func (l *SdkLogger) Info(msg string, keyvals ...interface{}) {
	l.logger.Info(msg, l.tags(keyvals)...)
}

func (l *SdkLogger) Warn(msg string, keyvals ...interface{}) {
	l.logger.Warn(msg, l.tags(keyvals)...)
}

func (l *SdkLogger) Error(msg string, keyvals ...interface{}) {
	l.logger.Error(msg, l.tags(keyvals)...)
}

func (l *SdkLogger) With(keyvals ...interface{}) sdklog.Logger {
	return NewSdkLogger(
		log.With(l.logger, l.tags(keyvals)...))
}
