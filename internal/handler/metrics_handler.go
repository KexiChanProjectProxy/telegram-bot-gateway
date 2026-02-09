package handler

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/kexi/telegram-bot-gateway/internal/pubsub"
	"github.com/kexi/telegram-bot-gateway/internal/repository"
	ws "github.com/kexi/telegram-bot-gateway/internal/websocket"
)

// MetricsHandler handles metrics endpoints
type MetricsHandler struct {
	db            interface{ Stats() interface{} }
	redisClient   *redis.Client
	wsHub         *ws.Hub
	messageBroker *pubsub.MessageBroker
	userRepo      repository.UserRepository
	messageRepo   repository.MessageRepository
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(
	db interface{ Stats() interface{} },
	redisClient *redis.Client,
	wsHub *ws.Hub,
	messageBroker *pubsub.MessageBroker,
	userRepo repository.UserRepository,
	messageRepo repository.MessageRepository,
) *MetricsHandler {
	return &MetricsHandler{
		db:            db,
		redisClient:   redisClient,
		wsHub:         wsHub,
		messageBroker: messageBroker,
		userRepo:      userRepo,
		messageRepo:   messageRepo,
	}
}

// GetMetrics returns system metrics
// @Summary Get system metrics
// @Description Get performance and usage metrics
// @Tags metrics
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /metrics [get]
func (h *MetricsHandler) GetMetrics(c *gin.Context) {
	ctx := context.Background()

	// Get memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Get Redis info
	redisPing := h.redisClient.Ping(ctx).Val()

	// Get pending webhook deliveries
	pendingDeliveries, _ := h.messageBroker.GetPendingWebhookDeliveryCount(ctx)

	metrics := gin.H{
		"timestamp": time.Now().Unix(),
		"uptime":    time.Since(startTime).Seconds(),
		"system": gin.H{
			"goroutines":    runtime.NumGoroutine(),
			"cpu_cores":     runtime.NumCPU(),
			"go_version":    runtime.Version(),
			"memory": gin.H{
				"alloc_mb":       m.Alloc / 1024 / 1024,
				"total_alloc_mb": m.TotalAlloc / 1024 / 1024,
				"sys_mb":         m.Sys / 1024 / 1024,
				"heap_alloc_mb":  m.HeapAlloc / 1024 / 1024,
				"heap_sys_mb":    m.HeapSys / 1024 / 1024,
				"gc_runs":        m.NumGC,
			},
		},
		"database": gin.H{
			"status": "connected",
			"stats":  h.db.Stats(),
		},
		"redis": gin.H{
			"status": redisPing,
			"connected": redisPing == "PONG",
		},
		"websocket": gin.H{
			"connected_clients": h.wsHub.GetClientCount(),
		},
		"webhooks": gin.H{
			"pending_deliveries": pendingDeliveries,
		},
	}

	c.JSON(http.StatusOK, metrics)
}

var startTime = time.Now()
