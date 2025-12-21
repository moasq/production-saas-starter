package cognitive

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/moasq/go-b2b-starter/internal/cognitive/app/services"
	"github.com/moasq/go-b2b-starter/internal/cognitive/domain"
	"github.com/moasq/go-b2b-starter/internal/auth"
	"github.com/moasq/go-b2b-starter/pkg/httperr"
)

type Handler struct {
	ragService       services.RAGService
	embeddingService services.EmbeddingService
}

func NewHandler(ragService services.RAGService, embeddingService services.EmbeddingService) *Handler {
	return &Handler{
		ragService:       ragService,
		embeddingService: embeddingService,
	}
}

// ChatRequest represents the JSON request body for chat
type ChatRequest struct {
	SessionID      int32  `json:"session_id,omitempty"`
	Message        string `json:"message" binding:"required"`
	UseRAG         bool   `json:"use_rag,omitempty"`
	MaxDocuments   int    `json:"max_documents,omitempty"`
	ContextHistory int    `json:"context_history,omitempty"`
}

// Chat sends a message and gets a response
// @Summary Chat with AI
// @Description Sends a message to the AI and gets a response, optionally using RAG
// @Tags Cognitive
// @Accept json
// @Produce json
// @Param request body ChatRequest true "Chat request"
// @Success 200 {object} github_com_moasq_go-b2b-starter_app_example_cognitive_domain.ChatResponse
// @Failure 400 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /example_cognitive/chat [post]
func (h *Handler) Chat(c *gin.Context) {
	reqCtx := auth.GetRequestContext(c)
	if reqCtx == nil {
		c.JSON(http.StatusBadRequest, errors.NewHTTPError(
			http.StatusBadRequest,
			"missing_context",
			"Organization context is required",
		))
		return
	}

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewHTTPError(
			http.StatusBadRequest,
			"invalid_request",
			"Invalid JSON format: "+err.Error(),
		))
		return
	}

	// Create domain request
	chatReq := &domain.ChatRequest{
		SessionID:      req.SessionID,
		Message:        req.Message,
		UseRAG:         req.UseRAG,
		MaxDocuments:   req.MaxDocuments,
		ContextHistory: req.ContextHistory,
	}

	response, err := h.ragService.Chat(c.Request.Context(), reqCtx.OrganizationID, reqCtx.AccountID, chatReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewHTTPError(
			http.StatusInternalServerError,
			"chat_failed",
			"Failed to process chat: "+err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListSessions lists chat sessions for the current user
// @Summary List chat sessions
// @Description Lists chat sessions for the current user with pagination
// @Tags Cognitive
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} errors.HTTPError
// @Router /example_cognitive/sessions [get]
func (h *Handler) ListSessions(c *gin.Context) {
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

	sessions, err := h.ragService.ListSessions(c.Request.Context(), reqCtx.OrganizationID, reqCtx.AccountID, int32(limit), int32(offset))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewHTTPError(
			http.StatusInternalServerError,
			"list_failed",
			"Failed to list sessions: "+err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetSessionHistory retrieves messages for a session
// @Summary Get session history
// @Description Retrieves all messages for a chat session
// @Tags Cognitive
// @Produce json
// @Param id path int true "Session ID"
// @Success 200 {array} github_com_moasq_go-b2b-starter_app_example_cognitive_domain.ChatMessage
// @Failure 400 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /example_cognitive/sessions/{id}/messages [get]
func (h *Handler) GetSessionHistory(c *gin.Context) {
	idParam := c.Param("id")
	var sessionID int32
	if _, err := fmt.Sscanf(idParam, "%d", &sessionID); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewHTTPError(
			http.StatusBadRequest,
			"invalid_id",
			"Session ID must be a valid number",
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

	messages, err := h.ragService.GetSessionHistory(c.Request.Context(), reqCtx.OrganizationID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewHTTPError(
			http.StatusInternalServerError,
			"fetch_failed",
			"Failed to fetch session history: "+err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, messages)
}
