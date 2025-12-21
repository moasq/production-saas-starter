package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/moasq/go-b2b-starter/internal/documents/domain"
	"github.com/moasq/go-b2b-starter/internal/documents/domain/events"
	"github.com/moasq/go-b2b-starter/internal/eventbus"
	filemanager "github.com/moasq/go-b2b-starter/internal/files"
	filedomain "github.com/moasq/go-b2b-starter/internal/files/domain"
	"github.com/moasq/go-b2b-starter/internal/logger"
	loggerdomain "github.com/moasq/go-b2b-starter/internal/logger/domain"
	ocrdomain "github.com/moasq/go-b2b-starter/internal/ocr/domain"
)

type documentService struct {
	docRepo     domain.DocumentRepository
	fileService filedomain.FileService
	ocrService  ocrdomain.OCRService
	eventBus    eventbus.EventBus
	logger      logger.Logger
}

func NewDocumentService(
	docRepo domain.DocumentRepository,
	fileService filedomain.FileService,
	ocrService ocrdomain.OCRService,
	eventBus eventbus.EventBus,
	logger logger.Logger,
) DocumentService {
	return &documentService{
		docRepo:     docRepo,
		fileService: fileService,
		ocrService:  ocrService,
		eventBus:    eventBus,
		logger:      logger,
	}
}

func (s *documentService) UploadDocument(ctx context.Context, orgID int32, req *UploadDocumentRequest, content io.Reader) (*domain.Document, error) {
	// Validate content type (only PDFs allowed)
	if !strings.Contains(strings.ToLower(req.ContentType), "pdf") {
		return nil, domain.ErrInvalidFileType
	}

	// Upload file using file manager
	fileReq := &filedomain.FileUploadRequest{
		Filename:    req.FileName,
		Size:        req.FileSize,
		ContentType: req.ContentType,
		Context:     filemanager.ContextGeneral,
		Metadata:    req.Metadata,
	}

	fileAsset, err := s.fileService.UploadFile(ctx, fileReq, content)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrFileUploadFailed, err)
	}

	// Create document record
	doc := &domain.Document{
		OrganizationID: orgID,
		FileAssetID:    fileAsset.ID,
		Title:          req.Title,
		FileName:       req.FileName,
		ContentType:    req.ContentType,
		FileSize:       req.FileSize,
		Status:         domain.DocumentStatusPending,
		Metadata:       req.Metadata,
	}

	createdDoc, err := s.docRepo.Create(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	// Process document asynchronously (extract text)
	go func() {
		processCtx := context.Background()
		s.ProcessDocument(processCtx, orgID, createdDoc.ID)
	}()

	return createdDoc, nil
}

func (s *documentService) GetDocument(ctx context.Context, orgID, docID int32) (*domain.Document, error) {
	doc, err := s.docRepo.GetByID(ctx, orgID, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return doc, nil
}

func (s *documentService) ListDocuments(ctx context.Context, orgID int32, req *ListDocumentsRequest) (*ListDocumentsResponse, error) {
	var docs []*domain.Document
	var total int64
	var err error

	if req.Status != nil {
		docs, err = s.docRepo.ListByStatus(ctx, orgID, *req.Status, req.Limit, req.Offset)
		if err != nil {
			return nil, fmt.Errorf("failed to list documents by status: %w", err)
		}
		total, err = s.docRepo.CountByStatus(ctx, orgID, *req.Status)
	} else {
		docs, err = s.docRepo.List(ctx, orgID, req.Limit, req.Offset)
		if err != nil {
			return nil, fmt.Errorf("failed to list documents: %w", err)
		}
		total, err = s.docRepo.Count(ctx, orgID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to count documents: %w", err)
	}

	return &ListDocumentsResponse{
		Documents: docs,
		Total:     total,
		Limit:     req.Limit,
		Offset:    req.Offset,
	}, nil
}

func (s *documentService) UpdateDocument(ctx context.Context, orgID, docID int32, req *UpdateDocumentRequest) (*domain.Document, error) {
	// Get existing document
	doc, err := s.docRepo.GetByID(ctx, orgID, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	// Update fields
	if req.Title != "" {
		doc.Title = req.Title
	}
	if req.Metadata != nil {
		doc.Metadata = req.Metadata
	}

	updatedDoc, err := s.docRepo.Update(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	return updatedDoc, nil
}

func (s *documentService) DeleteDocument(ctx context.Context, orgID, docID int32) error {
	// Get document to verify it exists
	doc, err := s.docRepo.GetByID(ctx, orgID, docID)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}

	// Delete the file asset
	if err := s.fileService.DeleteFile(ctx, doc.FileAssetID); err != nil {
		// Continue with document deletion even if file deletion fails
	}

	// Delete the document record
	if err := s.docRepo.Delete(ctx, orgID, docID); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

func (s *documentService) GetDocumentStats(ctx context.Context, orgID int32) (*domain.DocumentStats, error) {
	total, err := s.docRepo.Count(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to count documents: %w", err)
	}

	pending, err := s.docRepo.CountByStatus(ctx, orgID, domain.DocumentStatusPending)
	if err != nil {
		return nil, fmt.Errorf("failed to count pending documents: %w", err)
	}

	processed, err := s.docRepo.CountByStatus(ctx, orgID, domain.DocumentStatusProcessed)
	if err != nil {
		return nil, fmt.Errorf("failed to count processed documents: %w", err)
	}

	failed, err := s.docRepo.CountByStatus(ctx, orgID, domain.DocumentStatusFailed)
	if err != nil {
		return nil, fmt.Errorf("failed to count failed documents: %w", err)
	}

	return &domain.DocumentStats{
		TotalCount:     total,
		PendingCount:   pending,
		ProcessedCount: processed,
		FailedCount:    failed,
	}, nil
}

func (s *documentService) ProcessDocument(ctx context.Context, orgID, docID int32) (*domain.Document, error) {
	// Update status to processing
	doc, err := s.docRepo.UpdateStatus(ctx, orgID, docID, domain.DocumentStatusProcessing)
	if err != nil {
		return nil, fmt.Errorf("failed to update document status: %w", err)
	}

	// Download file content
	content, _, err := s.fileService.DownloadFile(ctx, doc.FileAssetID)
	if err != nil {
		s.markDocumentFailed(ctx, orgID, docID, err.Error())
		return nil, fmt.Errorf("%w: %v", domain.ErrFileDownloadFailed, err)
	}
	defer content.Close()

	// Extract text from PDF
	extractedText, err := s.extractTextFromPDF(content)
	if err != nil {
		s.markDocumentFailed(ctx, orgID, docID, err.Error())
		return nil, fmt.Errorf("%w: %v", domain.ErrTextExtractionFailed, err)
	}

	// Update document with extracted text
	doc, err = s.docRepo.UpdateExtractedText(ctx, orgID, docID, extractedText)
	if err != nil {
		s.markDocumentFailed(ctx, orgID, docID, err.Error())
		return nil, fmt.Errorf("failed to update extracted text: %w", err)
	}

	// Publish event for cognitive module to pick up
	event := events.NewDocumentUploaded(docID, orgID, doc.FileAssetID, doc.Title, extractedText)
	if err := s.eventBus.Publish(ctx, event); err != nil {
		// Don't fail the operation just because event publishing failed
	}

	return doc, nil
}

// markDocumentFailed marks a document as failed and publishes failure event
func (s *documentService) markDocumentFailed(ctx context.Context, orgID, docID int32, errMsg string) {
	s.docRepo.UpdateStatus(ctx, orgID, docID, domain.DocumentStatusFailed)

	// Publish failure event
	event := events.NewDocumentFailed(docID, orgID, errMsg)
	s.eventBus.Publish(ctx, event)
}

// extractTextFromPDF extracts text from a PDF file using OCR service
func (s *documentService) extractTextFromPDF(content io.Reader) (string, error) {
	// Read all content into memory
	data, err := io.ReadAll(content)
	if err != nil {
		return "", fmt.Errorf("failed to read PDF content: %w", err)
	}

	// Encode to base64 for OCR service
	base64Data := base64.StdEncoding.EncodeToString(data)

	// Call OCR service
	ctx := context.Background()
	ocrResult, err := s.ocrService.ExtractText(ctx, base64Data, "application/pdf")
	if err != nil {
		s.logger.Error("OCR extraction failed", loggerdomain.Fields{"error": err.Error()})
		return "", fmt.Errorf("OCR extraction failed: %w", err)
	}

	// Check confidence score
	const MinOCRConfidence = 0.7
	if ocrResult.Confidence < MinOCRConfidence {
		s.logger.Warn("OCR confidence below threshold", loggerdomain.Fields{
			"confidence":    ocrResult.Confidence,
			"pages":         ocrResult.Pages,
			"min_threshold": MinOCRConfidence,
		})
		// Still proceed but log the warning
	}

	// Log success
	s.logger.Info("Successfully extracted PDF text via OCR", loggerdomain.Fields{
		"pages":      ocrResult.Pages,
		"chars":      len(ocrResult.Text),
		"confidence": ocrResult.Confidence,
	})

	// Return extracted text (already in markdown format from Mistral)
	return ocrResult.Text, nil
}
