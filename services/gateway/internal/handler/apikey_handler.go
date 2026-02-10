package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kexi/telegram-bot-gateway/internal/service"
)

// APIKeyHandler handles API key endpoints
type APIKeyHandler struct {
	apiKeyService *service.APIKeyService
}

// NewAPIKeyHandler creates a new API key handler
func NewAPIKeyHandler(apiKeyService *service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyService: apiKeyService,
	}
}

// CreateAPIKey handles API key creation
// @Summary Create API key
// @Description Generate a new API key
// @Tags apikeys
// @Accept json
// @Produce json
// @Param request body service.CreateAPIKeyRequest true "API key details"
// @Success 201 {object} service.APIKeyDTO
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/apikeys [post]
func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	var req service.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiKey, err := h.apiKeyService.CreateAPIKey(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, apiKey)
}

// ListAPIKeys handles listing API keys
// @Summary List API keys
// @Description Get list of API keys
// @Tags apikeys
// @Produce json
// @Param offset query int false "Offset"
// @Param limit query int false "Limit"
// @Success 200 {array} service.APIKeyDTO
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/apikeys [get]
func (h *APIKeyHandler) ListAPIKeys(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	apiKeys, err := h.apiKeyService.ListAPIKeys(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, apiKeys)
}

// GetAPIKey handles getting a specific API key
// @Summary Get API key
// @Description Get API key by ID
// @Tags apikeys
// @Produce json
// @Param id path int true "API Key ID"
// @Success 200 {object} service.APIKeyDTO
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/apikeys/{id} [get]
func (h *APIKeyHandler) GetAPIKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid API key ID"})
		return
	}

	apiKey, err := h.apiKeyService.GetAPIKey(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, apiKey)
}

// RevokeAPIKey handles revoking an API key
// @Summary Revoke API key
// @Description Deactivate an API key
// @Tags apikeys
// @Param id path int true "API Key ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/apikeys/{id}/revoke [post]
func (h *APIKeyHandler) RevokeAPIKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid API key ID"})
		return
	}

	if err := h.apiKeyService.RevokeAPIKey(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key revoked successfully"})
}

// DeleteAPIKey handles deleting an API key
// @Summary Delete API key
// @Description Permanently delete an API key
// @Tags apikeys
// @Param id path int true "API Key ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/apikeys/{id} [delete]
func (h *APIKeyHandler) DeleteAPIKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid API key ID"})
		return
	}

	if err := h.apiKeyService.DeleteAPIKey(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key deleted successfully"})
}
