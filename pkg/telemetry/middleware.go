package telemetry

import (
	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
)

func NewFiberTracingMiddleware(serviceName string) fiber.Handler {
	return otelfiber.Middleware()
}
