package resilience

import (
	"context"
	"time"

	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

type CircuitBreakerConfig struct {
	MaxRequests      uint32
	Interval         time.Duration
	Timeout          time.Duration
	FailureThreshold uint32
}

func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxRequests:      3,
		Interval:         10 * time.Second,
		Timeout:          5 * time.Second,
		FailureThreshold: 3,
	}
}

func NewCircoutBreaker[T any](
	name string,
	cfg CircuitBreakerConfig,
	operation func(context.Context) (T, error),
) func(context.Context) (T, error) {
	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: cfg.MaxRequests,
		Interval:    cfg.Interval,
		Timeout:     cfg.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= cfg.FailureThreshold
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			zap.L().Warn("Circuit breaker state changed",
				zap.String("name", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()),
			)
		},
	}

	cb := gobreaker.NewCircuitBreaker(settings)

	return func(ctx context.Context) (T, error) {
		result, err := cb.Execute(func() (interface{}, error) {
			return operation(ctx)
		})
		if err != nil {
			var zero T
			return zero, err
		}
		return result.(T), nil
	}
}
