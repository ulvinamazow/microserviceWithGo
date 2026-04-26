package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	otelhttp "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.uber.org/zap"
)

func InitTracer(serviceName, endpoint string) (*sdktrace.TracerProvider, error) {
	exporter, err := otelhttp.New(
		context.Background(),
		otelhttp.WithEndpoint(endpoint),
		otelhttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	zap.L().Info("OpenTelemetry monitoring started",
		zap.String("service", serviceName),
		zap.String("endpoint", endpoint),
	)

	return tp, nil
}

func ShutDownTracer(tp *sdktrace.TracerProvider) {
	if tp != nil {
		ctx := context.Background()
		if err := tp.Shutdown(ctx); err != nil {
			zap.L().Error("Error stopping TracerProvider")
		}
	}
}
