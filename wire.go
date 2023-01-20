//go:build wireinject
// +build wireinject

package main

import (
	"prom/app"
  "github.com/google/wire"
	logadapter "prom/app/otel/zapadapter"
	"prom/core/domain/logger"
  "prom/core/domain/repository"
	"prom/app/db"
	"github.com/gofiber/fiber/v2"
	"prom/app/otel"
)

func ProvideZapLogger() (logger.Logger, error) {
	return logadapter.NewZapLogger()
}

func ProvideMysqlUserRepo() (repository.Connection, error)  {
  return db.New(conf.DBConnectionString)
}

func ProvideFiberHttpAdapter() *fiber.App  {
  return fiber.New()
}


func ProvideOtelAWSProvider() *app.OtelProviderImpl {
		return &app.OtelProviderImpl{
			TracerProvider:  otel.InitTracer,
			MetricsProvider: otel.InitMetricsProvider,
		}
}


var Set = wire.NewSet(
    ProvideZapLogger,
    ProvideMysqlUserRepo,
    ProvideFiberHttpAdapter,
    ProvideOtelAWSProvider,
    wire.Struct(new(app.Application), "Logger", "UserRepo", "HttpAdapter", "OtelProvider"))


func initializeApplication() (*app.Application, error) {
    wire.Build(Set)
    return &app.Application{}, nil
}
