package resilience

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  100 * time.Microsecond,
		MaxDelay:   2 * time.Second,
	}
}

func Retry[T any](
	ctx context.Context,
	cfg RetryConfig,
	operation func(context.Context) (T, error),
) (T, error) {
	var lastErr error
	delay := cfg.BaseDelay

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		result, err := operation(ctx)
		if err == nil {
			return result, nil
		}

		lastErr = err
		zap.L().Warn("Rtryable error, retrying",
			zap.Int("attempt", attempt),
			zap.Error(err),
		)

		if attempt == cfg.MaxRetries {
			break
		}

		select {
		case <-ctx.Done():
			var zero T
			return zero, fmt.Errorf("Retry aborted: %w", ctx.Err())
		case <-time.After(delay):

		}

		delay *= 2
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}
	}
	var zero T
	return zero, fmt.Errorf("operation failed after %d attempts: %w", cfg.MaxRetries+1, lastErr)
}
