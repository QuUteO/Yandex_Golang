package logger

import (
	"context"
	"go.uber.org/zap"
)

const (
	Key = "logger"

	RequestID = "request_id"
)

type Logger struct {
	l *zap.Logger
}

// New конструктор принимает контекст, создаем логер и кладем в контекст
func New(ctx context.Context) (context.Context, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	// кладем значение в контекст
	ctx = context.WithValue(ctx, Key, &Logger{logger})

	return ctx, nil
}

// методы логера:

// GetLoggerFromCtx получение контекста
func GetLoggerFromCtx(ctx context.Context) *Logger {
	return ctx.Value(Key).(*Logger)
}

// Info метод info
func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String("request_id", ctx.Value(RequestID).(string)))
	}
	l.l.Info(msg, fields...)
}

func (l *Logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String("request_id", ctx.Value(RequestID).(string)))
	}
	l.l.Info(msg, fields...)
}
