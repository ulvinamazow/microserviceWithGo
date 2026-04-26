package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/trace"

	"go.uber.org/zap"
)

func LoggerWithTrace(ctx context.Context) *zap.Logger {
	span := trace.SpanFromContext(ctx)

	if span.SpanContext().HasTraceID() {
		return zap.L().With(
			zap.String("trace_id", span.SpanContext().TraceID().String()),
			zap.String("span_id", span.SpanContext().SpanID().String()),
		)
	}

	return zap.L()
}
