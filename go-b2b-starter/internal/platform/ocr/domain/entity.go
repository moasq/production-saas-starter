package domain

// OCRResponse represents the result of OCR text extraction
type OCRResponse struct {
	Text       string  `json:"text"`       // Extracted text
	Pages      int     `json:"pages"`      // Number of pages processed
	Confidence float32 `json:"confidence"` // OCR confidence score (0.0 to 1.0)
}