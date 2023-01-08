package fbr

import (
	"fmt"
	"prom/app/config"
	"prom/core/domain/repository"

	_ "prom/app/fbr/_docs"
	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
)

var conf = config.GetConfig()

func InitHttpAdapter(app *fiber.App, userRepo repository.Connection) {

	app.Use(otelfiber.Middleware(conf.ServiceName,
		otelfiber.WithPropagators(xray.Propagator{}),
	))

	// TODO swagger docs should be only available when developing localy
	// TODO move fiber init to fiber package
	app.Get("/swagger/*", swagger.New(
		swagger.Config{
			// URL:         fmt.Sprintf("http://localhost:3000/swagger/doc.json"),
			URL:         fmt.Sprintf("/swagger/doc.json"),
			DeepLinking: true,
		},
	))
	app.Get("/v1/user", func(c *fiber.Ctx) error {
		return ListUsers(c, userRepo)
	})
	app.Get("/v1/user/:id", func(c *fiber.Ctx) error {
		return GetUser(c, userRepo)
	})
	app.Put("/v1/user", func(c *fiber.Ctx) error {
		return CreateUser(c, userRepo)
	})
	app.Post("/v1/user/:id", func(c *fiber.Ctx) error {
		return UpdateUser(c, userRepo)
	})
	app.Delete("/v1/user/:id", func(c *fiber.Ctx) error {
		return DeleteUser(c, userRepo)
	})

	go func() {
		app.Listen(fmt.Sprintf(":%s", conf.Port))
	}()
}
