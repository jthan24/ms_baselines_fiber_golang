package main

import (
	"context"
	"fmt"
	"log"
	"prom/core/domain/repository"

	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.uber.org/zap"
)

type providerCancelFunc = func(context.Context) error

type application struct {
	httpAdapter *fiber.App
	userRepo    repository.Connection
  tracerProvider func(context.Context) providerCancelFunc
  metricsProvider func(context.Context) providerCancelFunc
}

func (a *application) start() {
	logger := otelzap.New(zap.NewExample())
	defer logger.Sync()

	// TODO avoid global zap loggers an pass them via dependency injection
	undo := otelzap.ReplaceGlobals(logger)
	defer undo()

  initTracerCancelFunc := a.tracerProvider(context.Background())
	defer func() {
		err := initTracerCancelFunc(context.Background())
		log.Fatal(err)
	}()

  initMetricsCancelFunc := a.metricsProvider(context.Background())
	defer func() {
		err := initMetricsCancelFunc(context.Background())
		log.Fatal(err)
	}()

	a.httpAdapter.Use(otelfiber.Middleware(conf.ServiceName,
		otelfiber.WithPropagators(xray.Propagator{}),
	))

	// TODO swagger docs should be only available when developing localy
	a.httpAdapter.Get("/swagger/*", swagger.New(
		swagger.Config{
			// URL:         fmt.Sprintf("http://localhost:3000/swagger/doc.json"),
			URL:         fmt.Sprintf("/swagger/doc.json"),
			DeepLinking: true,
		},
	))
	a.httpAdapter.Get("/v1/user", func(c *fiber.Ctx) error {
		return ListUsers(c, a.userRepo)
	})
	a.httpAdapter.Get("/v1/user/:id", func(c *fiber.Ctx) error {
		return GetUser(c, a.userRepo)
	})
	a.httpAdapter.Put("/v1/user", func(c *fiber.Ctx) error {
		return CreateUser(c, a.userRepo)
	})
	a.httpAdapter.Post("/v1/user/:id", func(c *fiber.Ctx) error {
		return UpdateUser(c, a.userRepo)
	})
	a.httpAdapter.Delete("/v1/user/:id", func(c *fiber.Ctx) error {
		return DeleteUser(c, a.userRepo)
	})

	log.Fatal("Shutting down app", zap.Error(a.httpAdapter.Listen(fmt.Sprintf(":%s", conf.Port))))
}
