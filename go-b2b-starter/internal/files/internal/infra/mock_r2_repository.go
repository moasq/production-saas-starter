package infra

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/moasq/go-b2b-starter/internal/files/domain"
	"github.com/moasq/go-b2b-starter/internal/logger"
)

type mockR2Repository struct {
	logger logger.Logger
}

// NewMockR2Repository creates a mock R2 repository for development mode
// This repository simulates R2 operations without actual cloud storage
func NewMockR2Repository(log logger.Logger) domain.R2Repository {
	return &mockR2Repository{
		logger: log,
	}
}

func (m *mockR2Repository) UploadObject(ctx context.Context, objectKey string, content io.Reader, size int64, contentType string) error {
	m.logger.Warn("Mock R2: Simulating file upload (no actual storage)", map[string]any{
		"object_key":   objectKey,
		"size":         size,
		"content_type": contentType,
	})

	// Drain the reader to simulate upload
	io.Copy(io.Discard, content)

	return nil
}

func (m *mockR2Repository) DownloadObject(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	m.logger.Warn("Mock R2: Simulating file download (returning empty content)", map[string]any{
		"object_key": objectKey,
	})

	// Return empty reader
	return io.NopCloser(strings.NewReader("")), nil
}

func (m *mockR2Repository) DeleteObject(ctx context.Context, objectKey string) error {
	m.logger.Warn("Mock R2: Simulating file deletion (no actual storage)", map[string]any{
		"object_key": objectKey,
	})

	return nil
}

func (m *mockR2Repository) GetPresignedURL(ctx context.Context, objectKey string, expiryHours int) (string, error) {
	m.logger.Warn("Mock R2: Generating mock presigned URL", map[string]any{
		"object_key":   objectKey,
		"expiry_hours": expiryHours,
	})

	// Return a mock URL
	return fmt.Sprintf("https://mock-r2-storage.example.com/%s?expires=%dh", objectKey, expiryHours), nil
}

func (m *mockR2Repository) ObjectExists(ctx context.Context, objectKey string) (bool, error) {
	m.logger.Warn("Mock R2: Checking object existence (always returns true)", map[string]any{
		"object_key": objectKey,
	})

	// Always return true for mock
	return true, nil
}
