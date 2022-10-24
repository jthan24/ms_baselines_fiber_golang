package main

import (
	"context"
	"fmt"
	"log"

	_ "prom/app/_docs"

	"prom/app/mysql"

	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var userRepo mysql.UserRepo
var dsn string = "user:password@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=True&loc=Local"

var (
	serviceName  = "ExampleService"
	collectorURL = "localhost:4317"
)

func initTracer() *sdktrace.TracerProvider {
	ctx := context.Background()
	conn, err := grpc.DialContext(
		ctx,
		collectorURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatal(err, fmt.Errorf("failed to create gRPC connection to collector: %w", err))
	}

	// xray id generator
	idg := xray.NewIDGenerator()

	// Set up a trace exporter
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		log.Fatal(err, fmt.Errorf("failed to create trace exporter: %w", err))
	}
	// exporter, err := stdout.New(stdout.WithPrettyPrint())
	if err != nil {
		log.Fatal(err)
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
				semconv.ServiceNameKey.String("my-server"),
			)),
	)
	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(xray.Propagator{})
	// otel.SetTextMapPropagator(
	// 	propagation.NewCompositeTextMapPropagator(
	// 		propagation.TraceContext{},
	// 		propagation.Baggage{},
	// 	),
	// )
	return tp
}

func main() {
	app := fiber.New()
	// prometheus.Initialize(app)

	tp := initTracer()
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	config := swagger.Config{
		URL:         fmt.Sprintf("http://localhost:3000/swagger/doc.json"),
		DeepLinking: true,
	}

	var err error

	userRepo, err = mysql.New(dsn)

	if err != nil {
		log.Fatal(err)
	}

	userRepo.Initialize()

	app.Get("/swagger/*", swagger.New(config))
	app.Use(otelfiber.Middleware("my-server"))

	app.Get("/v1/user", ListUsers)
	app.Get("/v1/user/:id", GetUser)
	app.Put("/v1/user", CreateUser)
	app.Post("/v1/user/:id", UpdateUser)
	app.Delete("/v1/user/:id", DeleteUser)

	log.Fatal(app.Listen(":3000"))
}
