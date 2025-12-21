package domain

import "errors"

// Domain errors for documents
var (
	// Validation errors
	ErrDocumentOrganizationRequired = errors.New("document organization ID is required")
	ErrDocumentTitleRequired        = errors.New("document title is required")
	ErrDocumentFileNameRequired     = errors.New("document file name is required")
	ErrDocumentFileAssetRequired    = errors.New("document file asset ID is required")

	// Not found errors
	ErrDocumentNotFound = errors.New("document not found")

	// Processing errors
	ErrDocumentAlreadyProcessed = errors.New("document has already been processed")
	ErrDocumentProcessingFailed = errors.New("document processing failed")
	ErrTextExtractionFailed     = errors.New("text extraction from document failed")

	// File errors
	ErrInvalidFileType     = errors.New("invalid file type: only PDF files are allowed")
	ErrFileTooLarge        = errors.New("file size exceeds maximum allowed limit")
	ErrFileUploadFailed    = errors.New("failed to upload file")
	ErrFileDownloadFailed  = errors.New("failed to download file")
)
