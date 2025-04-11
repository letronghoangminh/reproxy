package logger

import (
	"sync"

	"github.com/letronghoangminh/reproxy/pkg/config"
	"github.com/letronghoangminh/reproxy/pkg/interfaces"
)

var (
	defaultLogger interfaces.Logger
	mu            sync.Mutex
)

func NewLogger(cfg config.Config) interfaces.Logger {
	logger := NewZapLogger(cfg)

	mu.Lock()
	defer mu.Unlock()
	defaultLogger = logger

	return logger
}

func GetLogger() interfaces.Logger {
	mu.Lock()
	defer mu.Unlock()

	if defaultLogger == nil {
		return &noopLogger{}
	}

	return defaultLogger
}

type noopLogger struct{}

func (l *noopLogger) Debug(msg string, fields ...interface{})      {}
func (l *noopLogger) Info(msg string, fields ...interface{})       {}
func (l *noopLogger) Warn(msg string, fields ...interface{})       {}
func (l *noopLogger) Error(msg string, fields ...interface{})      {}
func (l *noopLogger) Fatal(msg string, fields ...interface{})      {}
func (l *noopLogger) With(fields ...interface{}) interfaces.Logger { return l }
func (l *noopLogger) Sync() error                                  { return nil }
