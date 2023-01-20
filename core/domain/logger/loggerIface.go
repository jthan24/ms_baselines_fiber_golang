package logger

import (
	"context"

	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(ctx context.Context, msg string, fields ...zapcore.Field)
	Info(ctx context.Context, msg string, fields ...zapcore.Field)
	Warn(ctx context.Context, msg string, fields ...zapcore.Field)
	Error( ctx context.Context, msg string, fields ...zapcore.Field)
	Sync() error
}
