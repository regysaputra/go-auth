package logger

import (
	"fmt"
	"log/slog"
	"os"
)

// SlogAsynqLogger is an adapter to make slog.Logger compatible with asynq.Logger
type SlogAsynqLogger struct {
	logger *slog.Logger
}

// NewSlogAsynqLogger creates a new SlogAsynqLogger
func NewSlogAsynqLogger(logger *slog.Logger) *SlogAsynqLogger {
	return &SlogAsynqLogger{logger: logger}
}

// Debug These methods implement the asynq.Logger interface
func (l *SlogAsynqLogger) Debug(args ...interface{}) {
	l.logger.Debug(fmt.Sprint(args...))
}

// Info These methods implement the asynq.Logger interface
func (l *SlogAsynqLogger) Info(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

// Warn These methods implement the asynq.Logger interface
func (l *SlogAsynqLogger) Warn(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

// Error These methods implement the asynq.Logger interface
func (l *SlogAsynqLogger) Error(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}

// Fatal These methods implement the asynq.Logger interface
func (l *SlogAsynqLogger) Fatal(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...)) // the slog doesn't have Fatal, so we use Error
	os.Exit(1)
}
