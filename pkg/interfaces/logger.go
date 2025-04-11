package interfaces

type Logger interface {
	Debug(msg string, fields ...interface{})

	Info(msg string, fields ...interface{})

	Warn(msg string, fields ...interface{})

	Error(msg string, fields ...interface{})

	Fatal(msg string, fields ...interface{})

	With(fields ...interface{}) Logger

	Sync() error
}
