package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"prom/app/config"
	"prom/app/fbr"
	"prom/core/domain/repository"
	"sync"
	"syscall"
	"time"

	"prom/core/domain/logger"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
)

type ProviderCancelFunc = func(context.Context) error

type OtelProviderImpl struct {
	TracerProvider  func(context.Context) ProviderCancelFunc
	MetricsProvider func(context.Context) ProviderCancelFunc
}

type Application struct {
	Logger          logger.Logger
	HttpAdapter     *fiber.App
	UserRepo        repository.Connection
  OtelProvider    *OtelProviderImpl
	// TODO cleaner shutdown func
	tracerProviderShutdownFunc  ProviderCancelFunc
	metricsProviderShutdownFunc ProviderCancelFunc
}

var conf = config.GetConfig()

func (a *Application) Start() {
	a.tracerProviderShutdownFunc = a.OtelProvider.TracerProvider(context.Background())
	a.metricsProviderShutdownFunc = a.OtelProvider.MetricsProvider(context.Background())
	a.HttpAdapter.Use(otelfiber.Middleware(conf.ServiceName,
		otelfiber.WithPropagators(xray.Propagator{}),
	))
	fbr.InitHttpAdapter(a.HttpAdapter, a.UserRepo, a.Logger)
}

func shutdownHelper(
	ctx context.Context,
	timeout time.Duration,
	ops map[string]ProviderCancelFunc,
) <-chan struct{} {
	wait := make(chan struct{})
	go func() {
		// Create a buffered channel with capacity 1 to avoid blocking
		s := make(chan os.Signal, 1)

		// SIGHUP could be used to recreate application without restarting the whole process
		signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		<-s

		log.Println("Shutting down")

		// set timeout for forcing shutdown incase of hanging
		timeoutFunc := time.AfterFunc(timeout, func() {
			log.Printf("%d milliseconds has passed, forcing shutdown", timeout.Milliseconds())
			os.Exit(0)
		})

		defer timeoutFunc.Stop()

		var wg sync.WaitGroup
		wg.Add(len(ops))

		for key, op := range ops {
			go func(key string, op ProviderCancelFunc) {
				defer wg.Done()
				log.Printf("Cleaning up: %s", key)
				if err := op(ctx); err != nil {
					log.Println(fmt.Sprintf("%s: clean up failed: %v", key, err))
					return
				}

				log.Printf("%s was shutdown gracefully", key)
			}(key, op)
		}

		wg.Wait()

		close(wait)
	}()

	return wait
}

func (a *Application) Shutdown() {
	wait := shutdownHelper(context.Background(), 2*time.Second, map[string]ProviderCancelFunc{
		"httpAdapter": func(ctx context.Context) error {
			return a.HttpAdapter.Shutdown()
		},
		"metricsProvider": func(ctx context.Context) error {
			return a.metricsProviderShutdownFunc(ctx)
		},
		"tracerProvider": func(ctx context.Context) error {
			return a.tracerProviderShutdownFunc(ctx)
		},
	})
	<-wait
}
