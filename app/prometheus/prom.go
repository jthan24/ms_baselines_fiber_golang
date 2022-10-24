package prometheus

import (
	"math/rand"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func Initialize(app *fiber.App) {
	factory := promauto.With(prometheus.DefaultRegisterer)
	histogram := factory.NewHistogram(prometheus.HistogramOpts{
		Name:    "random_numbers",
		Help:    "A histogram of normally distributed random numbers.",
		Buckets: prometheus.LinearBuckets(-3, .1, 61),
	})

	go func() {
		for {
			histogram.Observe(rand.NormFloat64())
		}
	}()

	prometheus := fiberprometheus.NewWithRegistry(
		prometheus.DefaultRegisterer,
		"ms-baselines-golang-fiber",
		"http",
		"",
		nil,
	)
	prometheus.RegisterAt(app, "/metrics")

	app.Use(prometheus.Middleware)
}
