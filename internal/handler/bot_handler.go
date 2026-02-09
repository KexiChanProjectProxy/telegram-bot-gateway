package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kexi/telegram-bot-gateway/internal/service"
)

// BotHandler handles bot endpoints
type BotHandler struct {
	botService *service.BotService
}

// NewBotHandler creates a new bot handler
func NewBotHandler(botService *service.BotService) *BotHandler {
	return &BotHandler{
		botService: botService,
	}
}

// CreateBot handles bot registration
// @Summary Register bot
// @Description Register a new Telegram bot
// @Tags bots
// @Accept json
// @Produce json
// @Param request body service.CreateBotRequest true "Bot details"
// @Success 201 {object} service.BotDTO
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/bots [post]
func (h *BotHandler) CreateBot(c *gin.Context) {
	var req service.CreateBotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bot, err := h.botService.CreateBot(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, bot)
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

// DeleteBot handles deleting a bot
// @Summary Delete bot
// @Description Delete a bot
// @Tags bots
// @Param id path int true "Bot ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/bots/{id} [delete]
func (h *BotHandler) DeleteBot(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bot ID"})
		return
	}

	if err := h.botService.DeleteBot(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bot deleted successfully"})
}
