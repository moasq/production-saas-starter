package logger

import (
	"os"

	"github.com/moasq/go-b2b-starter/internal/platform/logger/domain"
	"github.com/rs/zerolog"
)

type Logger = domain.Logger
type Fields = domain.Fields

type structuredLogger struct{ logger zerolog.Logger }

// New writes structured logs to stdout; the container platform owns retention.
func New() Logger {
	return &structuredLogger{logger: zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Timestamp().Logger()}
}

func (l *structuredLogger) Debug(msg string, fields ...Fields) { write(l.logger.Debug(), msg, fields) }
func (l *structuredLogger) Info(msg string, fields ...Fields)  { write(l.logger.Info(), msg, fields) }
func (l *structuredLogger) Warn(msg string, fields ...Fields)  { write(l.logger.Warn(), msg, fields) }
func (l *structuredLogger) Error(msg string, fields ...Fields) { write(l.logger.Error(), msg, fields) }
func (l *structuredLogger) Fatal(msg string, fields ...Fields) { write(l.logger.Fatal(), msg, fields) }
func (l *structuredLogger) WithFields(fields Fields) Logger {
	return &structuredLogger{logger: l.logger.With().Fields(fields).Logger()}
}
func write(event *zerolog.Event, msg string, fields []Fields) {
	for _, values := range fields {
		event = event.Fields(values)
	}
	event.Msg(msg)
}
