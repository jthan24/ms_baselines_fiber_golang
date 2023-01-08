package main

import (
	"log"

	"prom/app/config"

	"prom/app"
	"prom/app/db"
	"prom/app/otel"

	"github.com/gofiber/fiber/v2"
)

var conf = config.GetConfig()

func main() {
	fiberApp := fiber.New()
	userRepo, err := db.New(conf.DBConnectionString)
	if err != nil {
		log.Fatal("Error setting up the mysql user repository %w: ", err)
	}

	a := &app.Application{
		HttpAdapter:     fiberApp,
		UserRepo:        userRepo,
		TracerProvider:  otel.InitTracer,
		MetricsProvider: otel.InitMetricsProvider,
	}

	a.Start()
	a.Shutdown()
}
