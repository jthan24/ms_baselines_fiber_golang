package fbr

import (
	"fmt"
	"prom/app/config"
	"prom/core/domain/logger"
	"prom/core/domain/repository"

	_ "prom/app/fbr/_docs"

	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
)

var conf = config.GetConfig()

func InitHttpAdapter(app *fiber.App, userRepo repository.Connection, log logger.Logger) {
	app.Use(recover.New(recover.Config{
    Next: nil,
    EnableStackTrace: true,
    StackTraceHandler: recover.ConfigDefault.StackTraceHandler,
  }))

	app.Use(otelfiber.Middleware(conf.ServiceName,
		otelfiber.WithPropagators(xray.Propagator{}),
	))

	app.Get("/swagger/*", swagger.New(
		swagger.Config{
			URL:         fmt.Sprintf("/swagger/doc.json"),
			DeepLinking: true,
		},
	))
	app.Get("/v1/user", func(c *fiber.Ctx) error {
		return ListUsers(c, userRepo, log)
	})
	app.Get("/v1/user/:id", func(c *fiber.Ctx) error {
		return GetUser(c, userRepo, log)
	})
	app.Post("/v1/user", func(c *fiber.Ctx) error {
		return CreateUser(c, userRepo, log)
	})
	app.Put("/v1/user/:id", func(c *fiber.Ctx) error {
		return UpdateUser(c, userRepo, log)
	})
	app.Delete("/v1/user/:id", func(c *fiber.Ctx) error {
		return DeleteUser(c, userRepo, log)
	})

	go func() {
		app.Listen(fmt.Sprintf(":%s", conf.Port))
	}()
}
