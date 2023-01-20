package zap

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZapLogger() (*ZapLogger, error) {
	cfg := zap.NewProductionConfig()
	// This will log all rows even when CPU is throtling see https://pkg.go.dev/go.uber.org/zap#SamplingConfig
	cfg.Sampling = nil

	adapter, err := cfg.Build()

	if err != nil {
		return nil, err
	}

	return &ZapLogger{
		adapter: adapter,
	}, nil
}

type ZapLogger struct {
	adapter *zap.Logger
}

func getTracingInfo(ctx context.Context) (*zap.Field, *zap.Field) {
	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	if spanCtx.HasSpanID() && spanCtx.HasTraceID() {
    trid := zap.String("trace-id", spanCtx.TraceID().String())
    spid :=  zap.String("span-id", spanCtx.SpanID().String())
		return &trid, &spid
	}
	return nil, nil
}

func (l *ZapLogger) Info(ctx context.Context, msg string,  args ...zapcore.Field) {
	traceId, spanId := getTracingInfo(ctx)
	l.adapter.Info(msg, append(args, *traceId, *spanId)...)
}

func (l *ZapLogger) Warn(ctx context.Context, msg string, args ...zapcore.Field) {
	traceId, spanId := getTracingInfo(ctx)
	l.adapter.Warn(msg, append(args, *traceId, *spanId)...)
}

func (l *ZapLogger) Error(ctx context.Context, msg string, args ...zapcore.Field) {
	traceId, spanId := getTracingInfo(ctx)
	l.adapter.Error(msg, append(args, *traceId, *spanId)...)
}

func (l *ZapLogger) Debug(ctx context.Context, msg string, args ...zapcore.Field) {
	traceId, spanId := getTracingInfo(ctx)
	l.adapter.Debug(msg, append(args, *traceId, *spanId)...)
}

func (l *ZapLogger) Sync() error {
	return l.adapter.Sync()
}
