package cmd

import (
	"context"
	"fmt"

	"go.uber.org/dig"

	"github.com/moasq/go-b2b-starter/app/example_cognitive"
	"github.com/moasq/go-b2b-starter/app/example_cognitive/app/services"
	docEvents "github.com/moasq/go-b2b-starter/app/example_documents/domain/events"
	"github.com/moasq/go-b2b-starter/pkg/eventbus"
)

func Init(container *dig.Container) error {
	module := cognitive.NewModule(container)
	if err := module.RegisterDependencies(); err != nil {
		return fmt.Errorf("failed to register cognitive dependencies: %w", err)
	}

	// Wire up event listener for document uploads
	if err := container.Invoke(func(
		bus eventbus.EventBus,
		listener services.DocumentListener,
	) error {
		// Subscribe to DocumentUploaded events
		return bus.Subscribe(docEvents.DocumentUploadedEventType, func(ctx context.Context, event eventbus.Event) error {
			// Type assert to get the specific event
			docEvent, ok := event.(*docEvents.DocumentUploaded)
			if !ok {
				return fmt.Errorf("unexpected event type: %T", event)
			}

			// Handle the event
			return listener.HandleDocumentUploaded(ctx, docEvent.DocumentID, docEvent.OrganizationID, docEvent.ExtractedText)
		})
	}); err != nil {
		return fmt.Errorf("failed to wire document event listener: %w", err)
	}

	return nil
}
