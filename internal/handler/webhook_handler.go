package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kexi/telegram-bot-gateway/internal/service"
)

// WebhookHandler handles webhook endpoints
type WebhookHandler struct {
	webhookService *service.WebhookService
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(webhookService *service.WebhookService) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
	}
}

// CreateWebhook handles webhook registration
// @Summary Register webhook
// @Description Register a new webhook for message notifications
// @Tags webhooks
// @Accept json
// @Produce json
// @Param request body service.CreateWebhookRequest true "Webhook details"
// @Success 201 {object} service.WebhookDTO
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/webhooks [post]
func (h *WebhookHandler) CreateWebhook(c *gin.Context) {
	var req service.CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	webhook, err := h.webhookService.CreateWebhook(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, webhook)
}

// ListWebhooks handles listing webhooks for a chat
// @Summary List webhooks
// @Description Get webhooks for a specific chat
// @Tags webhooks
// @Produce json
// @Param chat_id query int true "Chat ID"
// @Success 200 {array} service.WebhookDTO
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/webhooks [get]
func (h *WebhookHandler) ListWebhooks(c *gin.Context) {
	chatIDStr := c.Query("chat_id")
	if chatIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chat_id is required"})
		return
	}

	chatID, err := strconv.ParseUint(chatIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	webhooks, err := h.webhookService.ListWebhooksByChat(c.Request.Context(), uint(chatID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, webhooks)
}

// GetWebhook handles getting a specific webhook
// @Summary Get webhook
// @Description Get webhook by ID
// @Tags webhooks
// @Produce json
// @Param id path int true "Webhook ID"
// @Success 200 {object} service.WebhookDTO
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/webhooks/{id} [get]
func (h *WebhookHandler) GetWebhook(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	webhook, err := h.webhookService.GetWebhook(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, webhook)
}

// DeleteWebhook handles deleting a webhook
// @Summary Delete webhook
// @Description Delete a webhook
// @Tags webhooks
// @Param id path int true "Webhook ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/webhooks/{id} [delete]
func (h *WebhookHandler) DeleteWebhook(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	if err := h.webhookService.DeleteWebhook(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook deleted successfully"})
}

// UpdateWebhook handles updating a webhook
// @Summary Update webhook
// @Description Update webhook settings
// @Tags webhooks
// @Accept json
// @Produce json
// @Param id path int true "Webhook ID"
// @Param request body UpdateWebhookRequest true "Updated settings"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/webhooks/{id} [put]
func (h *WebhookHandler) UpdateWebhook(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	var req UpdateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.webhookService.UpdateWebhook(c.Request.Context(), uint(id), req.URL, req.IsActive); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook updated successfully"})
}

// UpdateWebhookRequest represents a webhook update request
type UpdateWebhookRequest struct {
	URL      string `json:"url"`
	IsActive bool   `json:"is_active"`
}
