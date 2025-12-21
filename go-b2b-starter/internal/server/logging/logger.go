package logging

import (
	"go.uber.org/zap"
)

type Logger struct {
	*zap.SugaredLogger
}

func InitLogger() (*Logger, error) {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	sugar := zapLogger.Sugar()
	return &Logger{sugar}, nil
}

func (l *Logger) Error(msg string, err error) {
	l.SugaredLogger.Errorw(msg, "error", err)
}

func (l *Logger) Fatal(msg string, err error) {
	l.SugaredLogger.Fatalw(msg, "error", err)
}

// Helper function to create a zap.Field for errors
func Error(err error) zap.Field {
	return zap.Error(err)
}
