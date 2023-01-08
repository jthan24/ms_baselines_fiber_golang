package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

var tracer = otel.Tracer("client-app")

func main() {
	initProvider()

	r := mux.NewRouter()
	r.Use(otelmux.Middleware("my-server"))

	// labels represent additional key-value descriptors that can be bound to a
	// metric observer or recorder.
	commonLabels := []attribute.KeyValue{
		attribute.String("labelA", "chocolate"),
		attribute.String("labelB", "raspberry"),
		attribute.String("labelC", "vanilla"),
	}

	r.HandleFunc(
		"/outgoing-http-call",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			w.Header().Set("Content-Type", "application/json")

			client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
			ctx := r.Context()

			xrayTraceID, _ := func(ctx context.Context) (string, error) {

				req, _ := http.NewRequestWithContext(
					ctx,
					"GET",
					"http://localhost:3000/v1/user",
					nil,
				)

				res, err := client.Do(req)
				if err != nil {
					handleErr(err, "HTTP call to aws.amazon.com failed")
				}

				_, _ = ioutil.ReadAll(res.Body)
				_ = res.Body.Close()

				return getXrayTraceID(trace.SpanFromContext(ctx)), err

			}(ctx)

			ctx, span := tracer.Start(
				ctx,
				"CollectorExporter-Example",
				trace.WithAttributes(commonLabels...))
			defer span.End()

			json := simplejson.New()
			json.Set("traceId", xrayTraceID)
			payload, _ := json.MarshalJSON()

			_, _ = w.Write(payload)

		}),
	)

	http.Handle("/", r)

  fmt.Println("Starting test client, curl localhost:8080/outgoing-http-call")
	// Start server
	_ = http.ListenAndServe("localhost:8080", nil)
}

func initProvider() {
	ctx := context.Background()
	endpoint := "0.0.0.0:4317" // setting default endpoint for exporter

	// Create and start new OTLP trace exporter
	traceExporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
	)
	handleErr(err, "failed to create new OTLP trace exporter")

	idg := xray.NewIDGenerator()

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		// the service name used to display traces in backends
		semconv.ServiceNameKey.String("test-client"),
	)
	handleErr(err, "failed to create resource")

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithIDGenerator(idg),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(xray.Propagator{})
}

func getXrayTraceID(span trace.Span) string {

	xrayTraceID := span.SpanContext().TraceID().String()
	result := fmt.Sprintf("1-%s-%s", xrayTraceID[0:8], xrayTraceID[8:])
	return result
}

func handleErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}
