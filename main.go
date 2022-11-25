package main

import (
	"log"

	_ "prom/app/_docs"
	"prom/app/config"

	"prom/app/db"

	"github.com/gofiber/fiber/v2"
)

var conf = config.GetConfig()


func main() {
	fiberApp := fiber.New()

	userRepo, err := db.New(conf.DBConnectionString)

	if err != nil {
		log.Fatal("Error setting up the mysql user repository %w: ", err)
	}

	a := &application{
		httpAdapter: fiberApp,
		userRepo:    userRepo,
    tracerProvider: initTracer,
    metricsProvider: initMetricsProvider,
	}

	a.start()
}
