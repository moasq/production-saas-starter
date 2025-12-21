package logger

import (
	"github.com/moasq/go-b2b-starter/internal/logger/domain"
	zerolog "github.com/moasq/go-b2b-starter/internal/logger/internal/zerologger"
)

func New(opts ...domain.Option) domain.Logger {
	options := &domain.Options{
		Level:  domain.InfoLevel,
		Output: domain.ConsoleOutput,
		FileOptions: domain.FileOptions{
			Filename:   "app.log",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		},
	}
	for _, opt := range opts {
		opt(options)
	}
	return zerolog.NewLogger(options)
}

// Re-export types and constants for ease of use
type (
	Logger = domain.Logger
	Fields = domain.Fields
	Level  = domain.Level
	Option = domain.Option
)

var (
	DebugLevel = domain.DebugLevel
	InfoLevel  = domain.InfoLevel
	WarnLevel  = domain.WarnLevel
	ErrorLevel = domain.ErrorLevel
	FatalLevel = domain.FatalLevel

	ConsoleOutput = domain.ConsoleOutput
	FileOutput    = domain.FileOutput
	BothOutput    = domain.BothOutput

	WithLevel       = domain.WithLevel
	WithOutput      = domain.WithOutput
	WithFileOptions = domain.WithFileOptions
)
