package log

import (
	"errors"

	"go.uber.org/zap"
)

type (
	Logger struct {
		logger    *zap.Logger
		isVerbose bool
	}

	loggerNoFields Logger
)

func New(l *zap.Logger, isVerbose bool) (*Logger, error) {
	if l == nil {
		return nil, errors.New("logger cannot be nil")
	}

	return &Logger{
		logger:    l,
		isVerbose: isVerbose,
	}, nil
}

func (l loggerNoFields) Info(msg string) {
	if l.isVerbose {
		l.logger.Info(msg)
	}
}

func (l Logger) Info(msg string, fields ...zap.Field) {
	if l.isVerbose {
		l.logger.Info(msg, fields...)
	}
}

func (l loggerNoFields) Warn(msg string) {
	if l.isVerbose {
		l.logger.Warn(msg)
	}
}

func (l Logger) Warning(msg string, fields ...zap.Field) {
	if l.isVerbose {
		l.logger.Warn(msg, fields...)
	}
}

func (l Logger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

func (l Logger) WithoutFields() loggerNoFields {
	return loggerNoFields(l)
}
