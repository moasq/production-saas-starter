package zerolog

import (
	"io"
	"os"
	"time"

	logger "github.com/moasq/go-b2b-starter/internal/logger/domain"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

type zerologLogger struct {
	zl zerolog.Logger
}

func newZerologLogger(opts *logger.Options) logger.Logger {
	var output io.Writer

	switch opts.Output {
	case logger.ConsoleOutput:
		output = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	case logger.FileOutput:
		output = &lumberjack.Logger{
			Filename:   opts.FileOptions.Filename,
			MaxSize:    opts.FileOptions.MaxSize,
			MaxBackups: opts.FileOptions.MaxBackups,
			MaxAge:     opts.FileOptions.MaxAge,
			Compress:   opts.FileOptions.Compress,
		}
	case logger.BothOutput:
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		fileWriter := &lumberjack.Logger{
			Filename:   opts.FileOptions.Filename,
			MaxSize:    opts.FileOptions.MaxSize,
			MaxBackups: opts.FileOptions.MaxBackups,
			MaxAge:     opts.FileOptions.MaxAge,
			Compress:   opts.FileOptions.Compress,
		}
		output = zerolog.MultiLevelWriter(consoleWriter, fileWriter)
	default:
		output = os.Stdout
	}

	zerolog.TimeFieldFormat = time.RFC3339

	zl := zerolog.New(output).With().Timestamp().Logger()

	// Set the log level
	zl = zl.Level(convertLogLevel(opts.Level))

	return &zerologLogger{zl: zl}
}

func (l *zerologLogger) Debug(msg string, fields ...logger.Fields) {
	l.log(l.zl.Debug(), msg, fields...)
}

func (l *zerologLogger) Info(msg string, fields ...logger.Fields) {
	l.log(l.zl.Info(), msg, fields...)
}

func (l *zerologLogger) Warn(msg string, fields ...logger.Fields) {
	l.log(l.zl.Warn(), msg, fields...)
}

func (l *zerologLogger) Error(msg string, fields ...logger.Fields) {
	l.log(l.zl.Error(), msg, fields...)
}

func (l *zerologLogger) Fatal(msg string, fields ...logger.Fields) {
	l.log(l.zl.Fatal(), msg, fields...)
}

func (l *zerologLogger) WithFields(fields logger.Fields) logger.Logger {
	return &zerologLogger{zl: l.zl.With().Fields(fields).Logger()}
}

func (l *zerologLogger) log(event *zerolog.Event, msg string, fields ...logger.Fields) {
	if len(fields) > 0 {
		event.Fields(fields[0])
	}
	event.Msg(msg)
}

func convertLogLevel(level logger.Level) zerolog.Level {
	switch level {
	case logger.DebugLevel:
		return zerolog.DebugLevel
	case logger.InfoLevel:
		return zerolog.InfoLevel
	case logger.WarnLevel:
		return zerolog.WarnLevel
	case logger.ErrorLevel:
		return zerolog.ErrorLevel
	case logger.FatalLevel:
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}
