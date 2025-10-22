package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.org/ulvinamazow/microservice_with_go/app/healthcheck"
	"github.org/ulvinamazow/microservice_with_go/pkg/config"
	_ "github.org/ulvinamazow/microservice_with_go/pkg/log"
)

type Request any
type Response any

type HandlerInterface[R Request, Res Response] interface {
	Handle(ctx context.Context, req *R) (*Res, error)
}

func handle[R Request, Res Response](handler HandlerInterface[R, Res]) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req R
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		res, err := handler.Handle(c.Context(), &req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.JSON(res)
	}
}

func main() {

	appConfig := config.Read()
	defer zap.L().Sync()

	zap.L().Info("app starting...")

	healthchechkHandler := healthcheck.NewHealthCheckHandler()

	app := fiber.New(fiber.Config{
		IdleTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 5 * time.Second,
		Concurrency:  1024,
	})

	app.Get("/healthcheck", handle(healthchechkHandler))

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

func Http() {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, er := http.NewRequestWithContext(ctx, http.MethodGet, "https://google.com", nil)
	if er != nil {
		zap.L().Error("failed to create HTTP request", zap.Error(er))
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		zap.L().Error("failed to make HTTP request", zap.Error(err))
	}

	zap.L().Info("HTTP request successful", zap.Int("status_code", resp.StatusCode))
}
