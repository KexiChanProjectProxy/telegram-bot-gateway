package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kexi/telegram-bot-gateway/internal/service"
)

// BotHandler handles bot endpoints (READ-ONLY - use ./bin/bot CLI for write operations)
type BotHandler struct {
	botService *service.BotService
}

// NewBotHandler creates a new bot handler
func NewBotHandler(botService *service.BotService) *BotHandler {
	return &BotHandler{
		botService: botService,
	}
}

// ListBots handles listing bots
// @Summary List bots
// @Description Get list of registered bots
// @Tags bots
// @Produce json
// @Param offset query int false "Offset"
// @Param limit query int false "Limit"
// @Success 200 {array} service.BotDTO
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/bots [get]
func (h *BotHandler) ListBots(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	bots, err := h.botService.ListBots(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bots)
}

// GetBot handles getting a specific bot
// @Summary Get bot
// @Description Get bot by ID
// @Tags bots
// @Produce json
// @Param id path int true "Bot ID"
// @Success 200 {object} service.BotDTO
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/bots/{id} [get]
func (h *BotHandler) GetBot(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bot ID"})
		return
	}

	bot, err := h.botService.GetBot(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bot)
}
