package eventbus

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// EventBus handles publishing and subscribing to events
type EventBus interface {
	// Publish publishes an event to all subscribers
	Publish(ctx context.Context, event Event) error
	// Subscribe registers a handler for a specific event type
	Subscribe(eventName string, handler EventHandler[Event]) error
	// Unsubscribe removes a handler for a specific event type
	Unsubscribe(eventName string, handler EventHandler[Event]) error
	// Close gracefully shuts down the event bus
	Close() error
}

// InMemoryEventBus is an in-memory implementation of EventBus
type InMemoryEventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]EventHandler[Event]
	middleware  []EventMiddleware
	closed      bool
}

func NewInMemoryEventBus(middleware ...EventMiddleware) EventBus {
	return &InMemoryEventBus{
		subscribers: make(map[string][]EventHandler[Event]),
		middleware:  middleware,
		closed:      false,
	}
}

// Publish publishes an event to all registered handlers
func (bus *InMemoryEventBus) Publish(ctx context.Context, event Event) error {
	bus.mu.RLock()
	if bus.closed {
		bus.mu.RUnlock()
		return fmt.Errorf("event bus is closed")
	}

	handlers := make([]EventHandler[Event], len(bus.subscribers[event.EventName()]))
	copy(handlers, bus.subscribers[event.EventName()])
	bus.mu.RUnlock()

	if len(handlers) == 0 {
		return nil
	}

	// Execute handlers concurrently
	var wg sync.WaitGroup
	errCh := make(chan error, len(handlers))

	for i, handler := range handlers {
		wg.Add(1)
		go func(handlerIndex int, h EventHandler[Event]) {
			defer wg.Done()

			// Apply middleware chain
			finalHandler := h
			for i := len(bus.middleware) - 1; i >= 0; i-- {
				finalHandler = bus.middleware[i](finalHandler)
			}

			if err := finalHandler(ctx, event); err != nil {
				errCh <- fmt.Errorf("handler error for event %s: %w", event.EventName(), err)
			}
		}(i, handler)
	}

	// Wait for all handlers to complete
	wg.Wait()
	close(errCh)

	// Collect any errors
	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("event handling errors: %v", errors)
	}

	return nil
}

// Subscribe registers a handler for a specific event type
func (bus *InMemoryEventBus) Subscribe(eventName string, handler EventHandler[Event]) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	if bus.closed {
		return fmt.Errorf("event bus is closed")
	}

	bus.subscribers[eventName] = append(bus.subscribers[eventName], handler)

	return nil
}

// Unsubscribe removes a handler for a specific event type
func (bus *InMemoryEventBus) Unsubscribe(eventName string, handler EventHandler[Event]) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	handlers := bus.subscribers[eventName]
	for i, h := range handlers {
		// Compare function pointers
		if reflect.ValueOf(h).Pointer() == reflect.ValueOf(handler).Pointer() {
			bus.subscribers[eventName] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}

	return nil
}

// Close gracefully shuts down the event bus
func (bus *InMemoryEventBus) Close() error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	bus.closed = true
	bus.subscribers = make(map[string][]EventHandler[Event])
	return nil
}

// GetSubscriberCount returns the number of subscribers for an event (for testing/debugging)
func (bus *InMemoryEventBus) GetSubscriberCount(eventName string) int {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	return len(bus.subscribers[eventName])
}