package domain

import "context"

// OCRService provides text extraction from files
type OCRService interface {
	ExtractText(ctx context.Context, base64File string, mimeType string) (*OCRResponse, error)
}