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

	grpcserver "github.org/ulvinamazow/microservice_with_go/api/grpc"
	"github.org/ulvinamazow/microservice_with_go/app/healthcheck"
	"github.org/ulvinamazow/microservice_with_go/app/product"
	"github.org/ulvinamazow/microservice_with_go/app/user"
	"github.org/ulvinamazow/microservice_with_go/pkg/config"
	"github.org/ulvinamazow/microservice_with_go/pkg/database"
	_ "github.org/ulvinamazow/microservice_with_go/pkg/log"
	"github.org/ulvinamazow/microservice_with_go/pkg/messaging"
	"github.org/ulvinamazow/microservice_with_go/pkg/repository"
	"github.org/ulvinamazow/microservice_with_go/pkg/resilience"
	telemetry "github.org/ulvinamazow/microservice_with_go/pkg/telemetry"
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
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		if err := c.ParamsParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		if err := c.QueryParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		if err := c.ReqHeaderParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		res, err := handler.Handle(c.UserContext(), &req)
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

	// Mongo connection

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	otelEndpoint := "localhost:4318"
	tracerProvider, err := telemetry.InitTracer("go-microservice", otelEndpoint)
	if err != nil {
		zap.L().Fatal("OpenTelmetry cannot started", zap.Error(err))
	}
	defer telemetry.ShutDownTracer(tracerProvider)

	mongoClient, err := database.Connect(ctx, appConfig.MongoDB)
	if err != nil {
		zap.L().Fatal("Error establishing MongoDB connection", zap.Error(err))
	}
	defer func() {
		shutdownCtx, sCancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer sCancel()
		if err := mongoClient.Close(shutdownCtx); err != nil {
			zap.L().Error("Error closing MongoDB connection", zap.Error(err))
		}
	}()

	// Kafka
	kafkaProducer := messaging.NewProducer(appConfig.Kafka.Brokers, appConfig.Kafka.Topic)
	defer func() {
		if err := kafkaProducer.Close(); err != nil {
			zap.L().Error("Error while closing Kafka producer", zap.Error(err))
		}
	}()
	customerCfg := messaging.CustomerConfig{
		Brokers: appConfig.Kafka.Brokers,
		Topic:   appConfig.Kafka.Topic,
		GroupID: "product-customer-group",
	}
	kafkaCustomer := messaging.NewCustomer(customerCfg)
	customerCtx, customerCancel := context.WithCancel(context.Background())
	defer customerCancel()
	go kafkaCustomer.Start(customerCtx)

	baseProductRepo := repository.NewMongoProductRepository(mongoClient.DB, appConfig.MongoDB.Collection)

	// Sonra onu retry və circuit breaker ilə əhatə edirik
	productRepo := repository.NewResilientProductRepository(
		baseProductRepo,
		resilience.DefaultRetryConfig(),
		resilience.DefaultCircuitBreakerConfig(),
	)

	// gRPC service
	grpcProductService := grpcserver.NewProductService(productRepo)
	// strting gRPC with goroutine
	grpcSrv := grpcserver.StartGRPCServer(":"+appConfig.GRPC.Port, grpcProductService)
	defer grpcSrv.GracefulStop()

	userRepo := repository.NewMongoUserRepository(mongoClient.DB, appConfig.Auth.UserCollection)

	// ---------------- Handlerlərin yaradılması ----------------
	healthcheckHandler := healthcheck.NewHealthCheckHandler()
	getProductHandler := product.NewGetProductHandler(productRepo)
	createProductHandler := product.NewCreateProductHandler(productRepo, kafkaProducer)
	deleteProductHandler := product.NewDeleteProducHandler(productRepo) // yazı səhvi düzəldildi
	listProductsHandler := product.NewListProductsHandler(productRepo)
	signupHandler := user.NewSignupHandler(userRepo)
	loginHandler := user.NewLoginHandler(userRepo, appConfig.Auth.JWTSecret, appConfig.Auth.JWTExpirationMin)

	app := fiber.New(fiber.Config{
		IdleTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 5 * time.Second,
		Concurrency:  1024,
	})

	app.Use(telemetry.NewFiberTracingMiddleware("go-microservice"))
	app.Get("/healthcheck", handle(healthcheckHandler))

	// Product routes
	app.Post("/auth/signup", handle(signupHandler))
	app.Post("/auth/login", handle(loginHandler))
	app.Get("/products", handle(listProductsHandler))
	app.Get("/products/:id", handle(getProductHandler))
	app.Post("/products", handle(createProductHandler))
	app.Delete("/products/:id", handle(deleteProductHandler))

	// metrics
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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	zap.L().Info("Stop signal received, stopping server...")

	customerCancel()

	if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
		zap.L().Error("Error during stop", zap.Error(err))
	}
	zap.L().Info("Server stopped successfully")

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
