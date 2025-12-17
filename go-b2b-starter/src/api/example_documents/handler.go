package documents

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/moasq/go-b2b-starter/app/example_documents/app/services"
	"github.com/moasq/go-b2b-starter/pkg/auth"
	"github.com/moasq/go-b2b-starter/pkg/common/errors"
)

type Handler struct {
	service services.DocumentService
}

func NewHandler(service services.DocumentService) *Handler {
	return &Handler{service: service}
}

// UploadDocument uploads a new PDF document
// @Summary Upload PDF document
// @Description Uploads a PDF document, extracts text, and creates embeddings
// @Tags Documents
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "PDF file to upload"
// @Param title formData string true "Document title"
// @Success 201 {object} github_com_moasq_go-b2b-starter_app_example_documents_domain.Document
// @Failure 400 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /example_documents/upload [post]
func (h *Handler) UploadDocument(c *gin.Context) {
	reqCtx := auth.GetRequestContext(c)
	if reqCtx == nil {
		c.JSON(http.StatusBadRequest, errors.NewHTTPError(
			http.StatusBadRequest,
			"missing_context",
			"Organization context is required",
		))
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewHTTPError(
			http.StatusBadRequest,
			"invalid_file",
			"Failed to read file: "+err.Error(),
		))
		return
	}
	defer file.Close()

	// Get title from form
	title := c.PostForm("title")
	if title == "" {
		title = header.Filename
	}

	// Create upload request
	req := &services.UploadDocumentRequest{
		Title:       title,
		FileName:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
		FileSize:    header.Size,
	}

	// Upload document
	document, err := h.service.UploadDocument(c.Request.Context(), reqCtx.OrganizationID, req, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewHTTPError(
			http.StatusInternalServerError,
			"upload_failed",
			"Failed to upload document: "+err.Error(),
		))
		return
	}

	c.JSON(http.StatusCreated, document)
}

// ListDocuments lists documents with pagination
// @Summary List documents
// @Description Lists documents with optional filtering and pagination
// @Tags Documents
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Param status query string false "Filter by status (pending, processing, processed, failed)"
// @Success 200 {object} github_com_moasq_go-b2b-starter_app_example_documents_app_services.ListDocumentsResponse
// @Failure 500 {object} errors.HTTPError
// @Router /example_documents [get]
func (h *Handler) ListDocuments(c *gin.Context) {
	reqCtx := auth.GetRequestContext(c)
	if reqCtx == nil {
		c.JSON(http.StatusBadRequest, errors.NewHTTPError(
			http.StatusBadRequest,
			"missing_context",
			"Organization context is required",
		))
		return
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	req := &services.ListDocumentsRequest{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	// Optional status filter
	// Note: Status filtering would need to be added if needed

	response, err := h.service.ListDocuments(c.Request.Context(), reqCtx.OrganizationID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewHTTPError(
			http.StatusInternalServerError,
			"list_failed",
			"Failed to list documents: "+err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Delete document
// @Description Deletes a document and its associated file
// @Tags Documents
// @Param id path int true "Document ID"
// @Success 204
// @Failure 400 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /example_documents/{id} [delete]
func (h *Handler) DeleteDocument(c *gin.Context) {
	idParam := c.Param("id")
	var docID int32
	if _, err := fmt.Sscanf(idParam, "%d", &docID); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewHTTPError(
			http.StatusBadRequest,
			"invalid_id",
			"Document ID must be a valid number",
		))
		return
	}

	reqCtx := auth.GetRequestContext(c)
	if reqCtx == nil {
		c.JSON(http.StatusBadRequest, errors.NewHTTPError(
			http.StatusBadRequest,
			"missing_context",
			"Organization context is required",
		))
		return
	}

	if err := h.service.DeleteDocument(c.Request.Context(), reqCtx.OrganizationID, docID); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewHTTPError(
			http.StatusInternalServerError,
			"delete_failed",
			"Failed to delete document: "+err.Error(),
		))
		return
	}

	c.Status(http.StatusNoContent)
}
