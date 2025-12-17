package services

import (
	"context"
	"fmt"
)

type documentListener struct {
	embeddingService EmbeddingService
}

func NewDocumentListener(
	embeddingService EmbeddingService,
) DocumentListener {
	return &documentListener{
		embeddingService: embeddingService,
	}
}

func (l *documentListener) HandleDocumentUploaded(ctx context.Context, documentID, orgID int32, text string) error {
	// Skip if no text to embed
	if text == "" {
		return nil
	}

	// Create embedding for the document
	_, err := l.embeddingService.EmbedDocument(ctx, orgID, documentID, text)
	if err != nil {
		return fmt.Errorf("failed to embed document: %w", err)
	}

	return nil
}
