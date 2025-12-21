# OCR Module Guide

Simple guide for extracting text from documents using OCR (Optical Character Recognition).

## Setup

Add to your `.env`:

```bash
MISTRAL_API_KEY=your-mistral-api-key-here
```

Optional:
```bash
MISTRAL_OCR_ENDPOINT=https://api.mistral.ai/v1/ocr  # Default
OCR_TIMEOUT_SEC=120                                 # Default
```

## Usage in Your Module

### 1. Inject the OCR Service

```go
import "github.com/moasq/go-b2b-starter/pkg/ocr/domain"

type InvoiceService struct {
    ocrService domain.OCRService
}

func NewInvoiceService(ocrService domain.OCRService) *InvoiceService {
    return &InvoiceService{ocrService: ocrService}
}
```

### 2. Extract Text from Document

```go
func (s *InvoiceService) ProcessDocument(ctx context.Context, base64File string, mimeType string) (string, error) {
    // Extract text from the document
    response, err := s.ocrService.ExtractText(ctx, base64File, mimeType)
    if err != nil {
        return "", fmt.Errorf("OCR failed: %w", err)
    }

    // Use the extracted text
    s.logger.Info("OCR completed", map[string]any{
        "pages":      response.Pages,
        "confidence": response.Confidence,
        "text_length": len(response.Text),
    })

    return response.Text, nil
}
```

### 3. Real-World Example: Invoice Processing

```go
func (s *InvoiceService) ExtractInvoiceData(ctx context.Context, fileData []byte) (*Invoice, error) {
    // 1. Convert file to base64
    base64File := base64.StdEncoding.EncodeToString(fileData)

    // 2. Extract text using OCR
    ocrResponse, err := s.ocrService.ExtractText(ctx, base64File, "application/pdf")
    if err != nil {
        return nil, err
    }

    // 3. Check confidence score
    if ocrResponse.Confidence < 0.7 {
        return nil, fmt.Errorf("OCR confidence too low: %.2f", ocrResponse.Confidence)
    }

    // 4. Parse the extracted text
    invoice := s.parseInvoiceText(ocrResponse.Text)

    return invoice, nil
}
```

## Response Structure

The `OCRResponse` includes:

```go
type OCRResponse struct {
    Text       string  // Extracted text from the document
    Pages      int     // Number of pages processed
    Confidence float32 // OCR confidence (0.0 to 1.0)
}
```

## Supported File Types

- **PDF**: `application/pdf`
- **Images**: `image/jpeg`, `image/png`
- **Other formats** supported by Mistral OCR API

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `MISTRAL_API_KEY` | *required* | Your Mistral API key |
| `MISTRAL_OCR_ENDPOINT` | `https://api.mistral.ai/v1/ocr` | OCR API endpoint |
| `OCR_TIMEOUT_SEC` | `120` | Request timeout in seconds |

## Best Practices

**1. Validate confidence scores:**
```go
if response.Confidence < 0.8 {
    // Low confidence - may need manual review
}
```

**2. Handle timeouts:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()

response, err := s.ocrService.ExtractText(ctx, base64File, mimeType)
```

**3. Process large files in chunks:**
For multi-page PDFs, OCR processes all pages automatically. Monitor the `Pages` field in the response.

That's it! Just inject `OCRService` and extract text from any document.
