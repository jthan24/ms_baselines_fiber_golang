package main

import (
	"context"
	"fmt"
	stdlog "log"
	"net/url"
	"os"
	"os/signal"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-logr/stdr"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	xraySampler "go.opentelemetry.io/contrib/samplers/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"

	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	serviceName  = "ExampleService"
	collectorURL = "localhost:4317"
)

func initMetricsProvider(ctx context.Context) func(context.Context) error {
	exp, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(collectorURL),
	)

	if err != nil {
		panic(err)
	}
	meterProvider := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(exp)))
	global.SetMeterProvider(meterProvider)

	return meterProvider.Shutdown
}

func initProvider(parentContext context.Context) (func(context.Context) error, error) {

	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	conn, err := grpc.DialContext(
		ctx,
		collectorURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	idg := xray.NewIDGenerator()

	// remoteSamplerCtx := context.Background()

	endpoint, err := url.Parse("http://127.0.0.1:2000")
	if err != nil {
		return nil, fmt.Errorf("failed to parse url for remote sampler: %w", err)
	}

	stdr.SetVerbosity(5)
	log := stdr.NewWithOptions(
		stdlog.New(os.Stderr, "", stdlog.LstdFlags),
		stdr.Options{LogCaller: stdr.All},
	)

	var rs sdktrace.Sampler
	// instantiate remote sampler with options
	rs, err = xraySampler.NewRemoteSampler(
		parentContext,
		serviceName,
		"eks",
		xraySampler.WithEndpoint(*endpoint),
		xraySampler.WithLogger(log),
	)
	if err != nil {
		return nil, fmt.Errorf("failed creating remote sampler: %w", err)
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(
		traceExporter,
	)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(rs),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithIDGenerator(idg),
	)
	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(xray.Propagator{})

	// Shutdown will flush any remaining spans and shut down the exporter.
	return tracerProvider.Shutdown, nil
}

func main() {
	stdlog.Printf("Waiting for connection...")

	// This must be always be background, checkout https://aws-otel.github.io/docs/getting-started/go-sdk/trace-manual-instr#sampling-using-aws-x-ray-remote-sampler for more information
	parentCtx := context.Background()

	ctx, cancel := signal.NotifyContext(parentCtx, os.Interrupt)
	defer cancel()

	shutdown, err := initProvider(parentCtx)
	if err != nil {
		stdlog.Fatal(err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			stdlog.Fatal("failed to shutdown TracerProvider: %w", err)
		}
	}()

	tracer := otel.Tracer("test-tracer")

	meter := global.Meter("test-meter")

	requestCount, err := meter.SyncInt64().Counter(
		fmt.Sprintf("%s/request_counter", serviceName),
		instrument.WithDescription("The number of requests received"),
	)

	if err != nil {
		stdlog.Fatal(err)
	}

	// Attributes represent additional key-value descriptors that can be bound
	// to a metric observer or recorder.
	commonAttrs := []attribute.KeyValue{
		attribute.String("attrA", "chocolate"),
		attribute.String("attrB", "raspberry"),
		attribute.String("attrC", "vanilla"),
	}

	// work begins
	ctx, span := tracer.Start(
		parentCtx,
		"OperationX",
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(commonAttrs...))
	defer span.End()

	cfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	otelaws.AppendMiddlewares(&cfg.APIOptions)

	// Metrics
	shutdownMetrics := initMetricsProvider(ctx)
	defer func() {
		if err := shutdownMetrics(ctx); err != nil {
			stdlog.Fatal("failed to shutdown MetricsProvider: %w", err)
		}
	}()

	s3Client := s3.NewFromConfig(cfg)

	for i := 0; i < 10; i++ {
		requestCount.Add(ctx, 10, commonAttrs...)

		<-time.After(time.Second)
	}

	for i := 0; i < 10; i++ {
		spanCtx, iSpan := tracer.Start(
			ctx,
			fmt.Sprintf("sub-Do-work-%d", i),
			trace.WithSpanKind(trace.SpanKindServer),
		)
		requestCount.Add(spanCtx, 1, commonAttrs...)
		input := &s3.ListBucketsInput{}
		time.Sleep(200 * time.Millisecond)
		_, err := s3Client.ListBuckets(spanCtx, input)
		if err != nil {
			fmt.Printf("Got an error retrieving buckets, %v", err)
			return
		}

		<-time.After(time.Second)
		iSpan.End()
	}

	stdlog.Printf("Done!")
}
