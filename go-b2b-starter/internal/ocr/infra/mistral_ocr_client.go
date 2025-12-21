package infra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/moasq/go-b2b-starter/internal/ocr/domain"
	loggerDomain "github.com/moasq/go-b2b-starter/internal/logger/domain"
)

type MistralOCRClient struct {
	config Config
	client *http.Client
	logger loggerDomain.Logger
}

// Mistral API request/response structures
type MistralOCRRequest struct {
	Model              string              `json:"model"`
	Document           MistralDocument     `json:"document"`
	IncludeImageBase64 bool                `json:"include_image_base64"`
}

type MistralDocument struct {
	Type        string `json:"type"`         // "document_url" or "image_url"
	DocumentURL string `json:"document_url,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
}

type MistralOCRResponse struct {
	Pages []MistralPage `json:"pages"`
}

type MistralPage struct {
	Index    int                    `json:"index"`
	Markdown string                 `json:"markdown"`
	Images   []MistralImage        `json:"images,omitempty"`
	Bboxes   []MistralBoundingBox  `json:"bboxes,omitempty"`
}

type MistralImage struct {
	Base64 string `json:"base64,omitempty"`
}

type MistralBoundingBox struct {
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
	Width  float32 `json:"width"`
	Height float32 `json:"height"`
	Text   string  `json:"text,omitempty"`
}

func NewMistralOCRClient(config Config, logger loggerDomain.Logger) (domain.OCRService, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	client := &http.Client{
		Timeout: time.Duration(config.TimeoutSec) * time.Second,
	}

	return &MistralOCRClient{
		config: config,
		client: client,
		logger: logger,
	}, nil
}


func (m *MistralOCRClient) ExtractText(ctx context.Context, base64File string, mimeType string) (*domain.OCRResponse, error) {
	m.logger.Info("Starting Mistral OCR extraction", map[string]any{
		"mime_type": mimeType,
	})

	// Validate file constraints
	if err := m.validateInput(base64File, mimeType); err != nil {
		return nil, err
	}

	// Build Mistral API request
	mistralRequest := m.buildMistralRequest(base64File, mimeType)

	// Make API call with retries
	mistralResponse, err := m.callMistralAPI(ctx, mistralRequest)
	if err != nil {
		return nil, err
	}

	// Convert response to domain format
	response := m.convertResponse(mistralResponse)

	m.logger.Info("Mistral OCR extraction completed", map[string]any{
		"pages":       response.Pages,
		"text_length": len(response.Text),
		"confidence":  response.Confidence,
	})

	return response, nil
}


func (m *MistralOCRClient) validateInput(base64File string, mimeType string) error {
	// Validate base64 file is not empty
	if base64File == "" {
		return domain.ErrInvalidInput
	}

	// Validate supported MIME types
	if !m.isSupportedMimeType(mimeType) {
		return domain.ErrUnsupportedFile
	}

	return nil
}

func (m *MistralOCRClient) isSupportedMimeType(mimeType string) bool {
	supportedTypes := []string{
		"application/pdf",
		"image/jpeg",
		"image/jpg", 
		"image/png",
		"image/tiff",
		"image/avif",
		"image/webp",
	}

	for _, supported := range supportedTypes {
		if mimeType == supported {
			return true
		}
	}

	return false
}

func (m *MistralOCRClient) buildMistralRequest(base64File string, mimeType string) MistralOCRRequest {
	mistralRequest := MistralOCRRequest{
		Model:              "mistral-ocr-latest",
		IncludeImageBase64: false, // Simplified - no layout extraction
	}

	// Determine document type based on MIME type and format as data URI
	if mimeType == "application/pdf" {
		dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64File)
		mistralRequest.Document = MistralDocument{
			Type:        "document_url",
			DocumentURL: dataURI,
		}
	} else if strings.HasPrefix(mimeType, "image/") {
		dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64File)
		mistralRequest.Document = MistralDocument{
			Type:     "image_url",
			ImageURL: dataURI,
		}
	}

	return mistralRequest
}

func (m *MistralOCRClient) callMistralAPI(ctx context.Context, mistralRequest MistralOCRRequest) (*MistralOCRResponse, error) {
	requestBody, err := json.Marshal(mistralRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", m.config.APIEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.config.MistralAPIKey)

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, domain.ErrAuthFailed
	}
	if resp.StatusCode == http.StatusBadRequest {
		return nil, domain.ErrInvalidInput
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, resp.Status)
	}

	var response MistralOCRResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}


func (m *MistralOCRClient) convertResponse(mistralResponse *MistralOCRResponse) *domain.OCRResponse {
	// Concatenate all page markdown with form feed separators
	var fullText strings.Builder
	for i, page := range mistralResponse.Pages {
		if i > 0 {
			fullText.WriteString("\f") // Page separator
		}
		fullText.WriteString(page.Markdown)
	}

	// Calculate confidence based on content quality
	confidence := m.calculateConfidence(fullText.String(), len(mistralResponse.Pages))

	return &domain.OCRResponse{
		Text:       fullText.String(),
		Pages:      len(mistralResponse.Pages),
		Confidence: confidence,
	}
}

func (m *MistralOCRClient) calculateConfidence(text string, pages int) float32 {
	if len(text) == 0 {
		return 0.0
	}

	// Base confidence for Mistral OCR
	confidence := float32(0.90)

	// Adjust based on text length (more text usually means better OCR)
	textLength := len(text)
	if textLength > 1000 {
		confidence += 0.05
	}
	if textLength > 5000 {
		confidence += 0.03
	}

	// Multi-page documents might have slightly lower confidence
	if pages > 5 {
		confidence -= 0.02
	}

	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}