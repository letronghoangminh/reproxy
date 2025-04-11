package logger

import (
	"os"

	"github.com/letronghoangminh/reproxy/pkg/config"
	"github.com/letronghoangminh/reproxy/pkg/interfaces"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	logger *zap.Logger
}

func (l *ZapLogger) Debug(msg string, fields ...interface{}) {
	l.logger.Debug(msg, toZapFields(fields)...)
}

func (l *ZapLogger) Info(msg string, fields ...interface{}) {
	l.logger.Info(msg, toZapFields(fields)...)
}

func (l *ZapLogger) Warn(msg string, fields ...interface{}) {
	l.logger.Warn(msg, toZapFields(fields)...)
}

func (l *ZapLogger) Error(msg string, fields ...interface{}) {
	l.logger.Error(msg, toZapFields(fields)...)
}

func (l *ZapLogger) Fatal(msg string, fields ...interface{}) {
	l.logger.Fatal(msg, toZapFields(fields)...)
}

func (l *ZapLogger) With(fields ...interface{}) interfaces.Logger {
	return &ZapLogger{
		logger: l.logger.With(toZapFields(fields)...),
	}
}

func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}

func NewZapLogger(cfg config.Config) interfaces.Logger {
	var level zapcore.Level
	switch cfg.Global.LogLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "fatal":
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)),
		level,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))

	return &ZapLogger{
		logger: logger,
	}
}

// toZapFields converts a slice of interface{} to zap.Field
// It assumes fields are provided in pairs: key1, value1, key2, value2, ...
func toZapFields(fields []interface{}) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields)/2)

	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			if key, ok := fields[i].(string); ok {
				zapFields = append(zapFields, zap.Any(key, fields[i+1]))
			}
		}
	}

	return zapFields
}
