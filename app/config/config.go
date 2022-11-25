package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

var (
	c *appConfig
)

type appConfig struct {
  Port               string `env:"PORT" env-default:"3000"`
	ServiceName        string `env:"SERVICE_NAME"         env-required:"true"`
	OTELCollectorURL   string `env:"OTEL_COLLECTOR_URL"                       env-default:"localhost:4317"`
	DBConnectionString string `env:"DB_CONNECTION_STRING" env-required:"true"`
	EnableOtelTraces   bool   `env:"ENABLE_OTEL_TRACES"                       env-default:"true"`
}

func GetConfig() *appConfig {
	if c == nil {
		c = new(appConfig)
		err := cleanenv.ReadEnv(c)
		if err != nil {
			log.Fatal(err)
		}
	}
	return c
}
