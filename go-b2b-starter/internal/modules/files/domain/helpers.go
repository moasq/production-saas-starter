package domain

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
)

// ConvertFileToBase64 reads a file from storage and converts it to a base64 data URI
// Returns a data URI in the format: data:{mimeType};base64,{encodedContent}
func ConvertFileToBase64(ctx context.Context, repo FileRepository, fileID int32) (string, error) {
	// Download the file content and metadata
	content, fileAsset, err := repo.Download(ctx, fileID)
	if err != nil {
		return "", fmt.Errorf("failed to download file %d: %w", fileID, err)
	}
	defer content.Close()

	// Read all content into memory
	data, err := io.ReadAll(content)
	if err != nil {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	// Convert to base64 and format as data URI
	return formatAsDataURI(data, fileAsset.ContentType), nil
}

// ConvertReaderToBase64 converts an io.Reader to a base64 data URI
// Useful for converting files that haven't been stored yet
func ConvertReaderToBase64(content io.Reader, contentType string) (string, error) {
	// Read all content into memory
	data, err := io.ReadAll(content)
	if err != nil {
		return "", fmt.Errorf("failed to read content: %w", err)
	}

	// Convert to base64 and format as data URI
	return formatAsDataURI(data, contentType), nil
}

// formatAsDataURI creates a properly formatted data URI from binary data
func formatAsDataURI(data []byte, contentType string) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", contentType, encoded)
}