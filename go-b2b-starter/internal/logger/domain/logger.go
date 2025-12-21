package domain

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

type OutputType int

const (
	ConsoleOutput OutputType = iota
	FileOutput
	BothOutput
)

type Fields map[string]interface{}

type Logger interface {
	Debug(msg string, fields ...Fields)
	Info(msg string, fields ...Fields)
	Warn(msg string, fields ...Fields)
	Error(msg string, fields ...Fields)
	Fatal(msg string, fields ...Fields)
	WithFields(fields Fields) Logger
}
