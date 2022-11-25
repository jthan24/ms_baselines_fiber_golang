package otel

import (
	"prom/app/config"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var tracerSingleLock = &sync.Mutex{}

var tracerInstance trace.Tracer

func GetTracerInstance() trace.Tracer {
	if tracerInstance == nil {
		tracerSingleLock.Lock()

		defer tracerSingleLock.Unlock()
		if tracerInstance == nil {
			tracerInstance = otel.Tracer(config.GetConfig().ServiceName)
		}
	}
	return tracerInstance
}
