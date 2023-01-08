package otel

import (
	"context"
	"fmt"
	"prom/app/config"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var conf = config.GetConfig()

func InitTracer(ctx context.Context) func(context.Context) error {
	conn, err := grpc.DialContext(
		ctx,
		conf.OTELCollectorURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		otelzap.L().Fatal("Failed to create gRPC connection to collector:", zap.Error(err))
	}

	// xray id generator
	idg := xray.NewIDGenerator()

	// Set up a trace exporter
	var exporter sdktrace.SpanExporter

	if conf.EnableOtelTraces {
		exporter, err = otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	} else {
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
	}

	if err != nil {
		otelzap.L().Fatal("Failed setting up the trace exporter:", zap.Error(err))
	}

	bsp := sdktrace.NewBatchSpanProcessor(
		exporter,
	)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithIDGenerator(idg),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(conf.ServiceName),
			)),
	)
	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(xray.Propagator{})

	return func(ctx context.Context) error {
		if err := tp.Shutdown(ctx); err != nil {
			return fmt.Errorf("Error shutting down tracer provider: %w", err)
		}
		return nil
	}
}

func InitMetricsProvider(ctx context.Context) func(context.Context) error {
	exp, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(conf.OTELCollectorURL),
	)

	if err != nil {
		otelzap.L().Fatal("Error configuring metrics provider", zap.Error(err))
	}
	meterProvider := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(exp)),
		metric.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(conf.ServiceName),
			semconv.CloudPlatformAWSEKS,
			semconv.DBSystemMySQL,
		)))
	global.SetMeterProvider(meterProvider)

	return func(context.Context) error {
		if err := meterProvider.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("Error shutting down metrics provider: %w", err)
		}
		return nil
	}
}
