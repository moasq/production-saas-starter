package eventbus

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/moasq/go-b2b-starter/internal/platform/logger/domain"
)

// LoggingMiddleware adds logging to event handling
func LoggingMiddleware(logger domain.Logger) EventMiddleware {
	return func(next EventHandler[Event]) EventHandler[Event] {
		return func(ctx context.Context, event Event) error {
			start := time.Now()
			logger.Info("Processing event", map[string]interface{}{
				"event_name": event.EventName(),
				"event_id":   event.EventID(),
				"timestamp":  event.Timestamp(),
			})

			err := next(ctx, event)
			duration := time.Since(start)

			if err != nil {
				logger.Error("Event processing failed", map[string]interface{}{
					"event_name": event.EventName(),
					"event_id":   event.EventID(),
					"error":      err.Error(),
					"duration":   duration,
				})
			} else {
				logger.Info("Event processed successfully", map[string]interface{}{
					"event_name": event.EventName(),
					"event_id":   event.EventID(),
					"duration":   duration,
				})
			}

			return err
		}
	}
}

// RecoveryMiddleware recovers from panics in event handlers
func RecoveryMiddleware(logger domain.Logger) EventMiddleware {
	return func(next EventHandler[Event]) EventHandler[Event] {
		return func(ctx context.Context, event Event) (err error) {
			defer func() {
				if r := recover(); r != nil {
					stack := debug.Stack()

					// Get event metadata size for debugging
					metadata := event.Metadata()
					metadataSize := len(fmt.Sprintf("%+v", metadata))

					logger.Error("Event handler panicked", map[string]interface{}{
						"event_name":       event.EventName(),
						"event_id":         event.EventID(),
						"event_timestamp":  event.Timestamp(),
						"panic":            r,
						"stack_trace":      string(stack),
						"metadata_size":    metadataSize,
						"metadata_keys":    getMapKeys(metadata),
						"recovery_context": "eventbus_middleware",
					})
					err = fmt.Errorf("event handler panicked: %v", r)
				}
			}()

			return next(ctx, event)
		}
	}
}

// Helper function to safely extract map keys
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// MetricsMiddleware adds metrics collection to event handling
func MetricsMiddleware() EventMiddleware {
	return func(next EventHandler[Event]) EventHandler[Event] {
		return func(ctx context.Context, event Event) error {
			start := time.Now()

			err := next(ctx, event)
			duration := time.Since(start)

			// Here you could send metrics to Prometheus, StatsD, etc.
			// For now, we'll just log the metrics
			_ = duration // Placeholder for actual metrics implementation

			return err
		}
	}
}
