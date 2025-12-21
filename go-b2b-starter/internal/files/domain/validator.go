package domain

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

// ValidateFileContent verifies that the actual file content matches the declared file type
// based on magic bytes inspection. This prevents file extension spoofing attacks.
func ValidateFileContent(reader io.Reader, filename string) error {
	// Detect actual MIME type from file content (magic bytes)
	mtype, err := mimetype.DetectReader(reader)
	if err != nil {
		return fmt.Errorf("failed to detect file MIME type: %w", err)
	}

	// Get file extension
	ext := strings.ToLower(filepath.Ext(filename))

	// Get allowed MIME types for this extension
	allowedMIMEs, ok := getAllowedMIMETypes(ext)
	if !ok {
		return fmt.Errorf("unsupported file extension: %s", ext)
	}

	// Check if detected MIME type matches expected types
	detectedMIME := mtype.String()
	for _, allowed := range allowedMIMEs {
		if detectedMIME == allowed {
			return nil
		}
	}

	return fmt.Errorf("file content type (%s) does not match extension (%s)", detectedMIME, ext)
}

// getAllowedMIMETypes returns the list of allowed MIME types for a given file extension
func getAllowedMIMETypes(ext string) ([]string, bool) {
	// Invoice-specific allowed MIME types
	mimeMap := map[string][]string{
		".pdf": {
			"application/pdf",
		},
		".png": {
			"image/png",
		},
		".jpg": {
			"image/jpeg",
		},
		".jpeg": {
			"image/jpeg",
		},
	}

	mimes, ok := mimeMap[ext]
	return mimes, ok
}

// SanitizeFilename removes dangerous characters and path traversal attempts from filenames
// to prevent security vulnerabilities.
func SanitizeFilename(filename string) string {
	// Step 1: Remove any path components (prevents path traversal)
	filename = filepath.Base(filename)

	// Step 2: Get extension before sanitization
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	// Step 3: Remove dangerous characters, keep only safe ones
	// Allow: alphanumeric, dash, underscore, space
	safePattern := regexp.MustCompile(`[^a-zA-Z0-9\-_ ]`)
	nameWithoutExt = safePattern.ReplaceAllString(nameWithoutExt, "_")

	// Step 4: Remove multiple consecutive underscores or spaces
	multipleUnderscores := regexp.MustCompile(`_+`)
	nameWithoutExt = multipleUnderscores.ReplaceAllString(nameWithoutExt, "_")

	multipleSpaces := regexp.MustCompile(`\s+`)
	nameWithoutExt = multipleSpaces.ReplaceAllString(nameWithoutExt, " ")

	// Step 5: Trim leading/trailing whitespace and underscores
	nameWithoutExt = strings.Trim(nameWithoutExt, " _")

	// Step 6: If filename is empty after sanitization, use default
	if nameWithoutExt == "" {
		nameWithoutExt = "file"
	}

	// Step 7: Reconstruct filename with extension
	sanitized := nameWithoutExt + ext

	// Step 8: Limit total filename length to 255 characters (filesystem limit)
	maxLength := 255
	if len(sanitized) > maxLength {
		// Keep the extension, truncate the name
		extLen := len(ext)
		maxNameLen := maxLength - extLen
		if maxNameLen > 0 {
			sanitized = sanitized[:maxNameLen] + ext
		} else {
			// If extension itself is too long (unlikely), just truncate
			sanitized = sanitized[:maxLength]
		}
	}

	return sanitized
}

// IsInvoiceFileType checks if the file extension is allowed for invoice uploads
func IsInvoiceFileType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExtensions := []string{".pdf", ".png", ".jpg", ".jpeg"}

	for _, allowed := range allowedExtensions {
		if ext == allowed {
			return true
		}
	}

	return false
}
