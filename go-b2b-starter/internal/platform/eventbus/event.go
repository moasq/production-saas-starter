package eventbus

import (
	"context"
	"time"
)

// Event represents a domain event that can be published and subscribed to
type Event interface {
	// EventName returns the unique name/type of the event
	EventName() string
	// EventID returns a unique identifier for this specific event instance
	EventID() string
	// Timestamp returns when the event was created
	Timestamp() time.Time
	// Metadata returns additional event metadata
	Metadata() map[string]interface{}
}

// BaseEvent provides common event functionality
type BaseEvent struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	CreatedAt time.Time              `json:"created_at"`
	Meta      map[string]interface{} `json:"metadata,omitempty"`
}

func (e BaseEvent) EventName() string {
	return e.Name
}

func (e BaseEvent) EventID() string {
	return e.ID
}

func (e BaseEvent) Timestamp() time.Time {
	return e.CreatedAt
}

func (e BaseEvent) Metadata() map[string]interface{} {
	return e.Meta
}

// EventHandler represents a function that handles an event
type EventHandler[T Event] func(ctx context.Context, event T) error

// EventMiddleware can be used to add cross-cutting concerns like logging, metrics, etc.
type EventMiddleware func(next EventHandler[Event]) EventHandler[Event]