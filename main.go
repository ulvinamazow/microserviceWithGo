package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.org/ulvinamazow/microservice_with_go/pkg/config"
	_ "github.org/ulvinamazow/microservice_with_go/pkg/log"
)

func main() {

	appConfig := config.Read()
	defer zap.L().Sync()

	zap.L().Info("app starting...")

	app := fiber.New()

	app.Get("/healthcheck", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// start the server in a goroutine so we can listen for shutdown signals
	go func() {
		if err := app.Listen(fmt.Sprintf(":%s", appConfig.Port)); err != nil {
			// Listen returns an error when the server is closed; log fatal only if it's unexpected
			zap.L().Fatal("Failed to start server", zap.Error(err))
		}
	}()

	gracefulShutDown(app)

}

func gracefulShutDown(app *fiber.App) {
	// listen for OS signals for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	zap.L().Info("shutdown signal received, shutting down server...")

	if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
		zap.L().Error("error during shutdown", zap.Error(err))
	}

	zap.L().Info("server stopped gracefully")
}
