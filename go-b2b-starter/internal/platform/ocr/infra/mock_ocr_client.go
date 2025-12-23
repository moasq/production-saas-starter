package infra

import (
	"context"
	"strings"
	"time"

	"github.com/moasq/go-b2b-starter/internal/platform/ocr/domain"
	loggerDomain "github.com/moasq/go-b2b-starter/internal/platform/logger/domain"
)

// MockOCRClient is a mock implementation for development/testing
// In production, this would be replaced with actual Google Vision client
type MockOCRClient struct {
	config Config
	logger loggerDomain.Logger
}

func NewMockOCRClient(config Config, logger loggerDomain.Logger) (domain.OCRService, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &MockOCRClient{
		config: config,
		logger: logger,
	}, nil
}

func (m *MockOCRClient) ExtractText(ctx context.Context, base64File string, mimeType string) (*domain.OCRResponse, error) {
	m.logger.Info("Mock OCR extraction starting", map[string]any{
		"mime_type": mimeType,
	})

	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	// Mock extracted text based on file type
	var mockText string
	var pages int = 1

	if mimeType == "application/pdf" {
		pages = 2
		mockText = `INVOICE
Invoice Number: INV-2024-001
Date: January 15, 2024

Bill To:
ABC Company
123 Main Street
City, State 12345

Description                    Qty    Unit Price    Total
Professional Services          10     $150.00      $1,500.00
Consulting                      5     $200.00      $1,000.00

                              Subtotal: $2,500.00
                                   Tax: $250.00
                                 Total: $2,750.00

Payment Terms: Net 30
Due Date: February 15, 2024`
	} else if strings.HasPrefix(mimeType, "image/") {
		mockText = `RECEIPT
Store: Tech Solutions Inc.
Date: 2024-01-10
Receipt #: R-789456

Items:
- Software License    $299.99
- Support Package     $99.99

Subtotal: $399.98
Tax: $32.00
Total: $431.98

Thank you for your business!`
	} else {
		return nil, domain.ErrUnsupportedFile
	}

	response := &domain.OCRResponse{
		Text:       mockText,
		Pages:      pages,
		Confidence: 0.95,
	}

	m.logger.Info("Mock OCR extraction completed", map[string]any{
		"pages":       pages,
		"text_length": len(mockText),
		"confidence":  response.Confidence,
	})

	return response, nil
}

