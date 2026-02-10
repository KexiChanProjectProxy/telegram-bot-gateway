package notifications

import (
	"context"
	"fmt"
	"strings"

	"github.com/user/weather-notice-bot/internal/llm"
	"github.com/user/weather-notice-bot/internal/telegram"
	"github.com/user/weather-notice-bot/internal/weather"

	"github.com/rs/zerolog"
)

// WeatherDetector defines the interface for weather change detection
type WeatherDetector interface {
	// DetectChanges compares current weather with previous state
	// Returns a description of significant changes, or empty string if no significant changes
	DetectChanges(ctx context.Context, current *weather.WeatherResponse) (string, error)
}

// LocationContext holds all the context needed for a single location
type LocationContext struct {
	Name      string
	Latitude  float64
	Longitude float64
	Detector  WeatherDetector
}

// ChatHandler manages weather notification operations for a single chat
// with potentially multiple locations
type ChatHandler struct {
	weatherClient  *weather.Client
	telegramClient *telegram.Client
	llmClient      *llm.Client
	chatID         int64
	chatName       string
	locations      []LocationContext
	logger         zerolog.Logger
}

// NewChatHandler creates a new chat handler instance
func NewChatHandler(
	weatherClient *weather.Client,
	telegramClient *telegram.Client,
	llmClient *llm.Client,
	chatID int64,
	chatName string,
	locations []LocationContext,
	logger zerolog.Logger,
) *ChatHandler {
	return &ChatHandler{
		weatherClient:  weatherClient,
		telegramClient: telegramClient,
		llmClient:      llmClient,
		chatID:         chatID,
		chatName:       chatName,
		locations:      locations,
		logger:         logger.With().Str("component", "chat_handler").Int64("chat_id", chatID).Logger(),
	}
}

// HandleMorningNotification sends the morning weather notification
func (h *ChatHandler) HandleMorningNotification() error {
	ctx := context.Background()

	h.logger.Info().Msg("handling morning notification")

	var sections []string

	// Process each location
	for _, loc := range h.locations {
		// Fetch current weather
		weatherData, err := h.weatherClient.GetWeather(ctx, loc.Longitude, loc.Latitude)
		if err != nil {
			h.logger.Error().Err(err).Str("location", loc.Name).Msg("failed to fetch weather data")
			sections = append(sections, fmt.Sprintf("❌ <b>%s</b>: 获取天气数据失败", loc.Name))
			continue
		}

		// Generate LLM advice
		prompt := buildMorningPromptFromResponse(loc.Name, weatherData)
		advice, err := h.llmClient.GenerateAdvice(ctx, prompt)
		if err != nil {
			h.logger.Warn().Err(err).Str("location", loc.Name).Msg("failed to generate LLM advice, continuing without it")
			advice = "" // Continue without advice rather than failing
		}

		// Format message for this location
		message := FormatMorningMessage(loc.Name, weatherData, advice)
		sections = append(sections, message)
	}

	// Combine all sections
	fullMessage := strings.Join(sections, "\n\n" + strings.Repeat("─", 30) + "\n\n")

	// Split message if needed for Telegram's 4096 char limit
	messages := splitMessage(fullMessage, 4096)

	// Send via Telegram
	for i, msg := range messages {
		if err := h.telegramClient.SendMessage(ctx, h.chatID, msg, "HTML"); err != nil {
			h.logger.Error().Err(err).Int("part", i+1).Msg("failed to send morning notification")
			return fmt.Errorf("send telegram message part %d: %w", i+1, err)
		}
	}

	h.logger.Info().Int("locations", len(h.locations)).Int("messages", len(messages)).Msg("morning notification sent successfully")
	return nil
}

// HandleEveningNotification sends the evening weather notification
func (h *ChatHandler) HandleEveningNotification() error {
	ctx := context.Background()

	h.logger.Info().Msg("handling evening notification")

	var sections []string

	// Process each location
	for _, loc := range h.locations {
		// Fetch current weather
		weatherData, err := h.weatherClient.GetWeather(ctx, loc.Longitude, loc.Latitude)
		if err != nil {
			h.logger.Error().Err(err).Str("location", loc.Name).Msg("failed to fetch weather data")
			sections = append(sections, fmt.Sprintf("❌ <b>%s</b>: 获取天气数据失败", loc.Name))
			continue
		}

		// Generate LLM advice
		prompt := buildEveningPromptFromResponse(loc.Name, weatherData)
		advice, err := h.llmClient.GenerateAdvice(ctx, prompt)
		if err != nil {
			h.logger.Warn().Err(err).Str("location", loc.Name).Msg("failed to generate LLM advice, continuing without it")
			advice = "" // Continue without advice rather than failing
		}

		// Format message for this location
		message := FormatEveningMessage(loc.Name, weatherData, advice)
		sections = append(sections, message)
	}

	// Combine all sections
	fullMessage := strings.Join(sections, "\n\n" + strings.Repeat("─", 30) + "\n\n")

	// Split message if needed for Telegram's 4096 char limit
	messages := splitMessage(fullMessage, 4096)

	// Send via Telegram
	for i, msg := range messages {
		if err := h.telegramClient.SendMessage(ctx, h.chatID, msg, "HTML"); err != nil {
			h.logger.Error().Err(err).Int("part", i+1).Msg("failed to send evening notification")
			return fmt.Errorf("send telegram message part %d: %w", i+1, err)
		}
	}

	h.logger.Info().Int("locations", len(h.locations)).Int("messages", len(messages)).Msg("evening notification sent successfully")
	return nil
}

// HandleWeatherPoll checks for weather changes and sends alerts if needed
func (h *ChatHandler) HandleWeatherPoll() error {
	ctx := context.Background()

	h.logger.Debug().Msg("polling weather for changes")

	var sections []string
	hasChanges := false

	// Process each location
	for _, loc := range h.locations {
		// Fetch current weather
		weatherData, err := h.weatherClient.GetWeather(ctx, loc.Longitude, loc.Latitude)
		if err != nil {
			h.logger.Error().Err(err).Str("location", loc.Name).Msg("failed to fetch weather data")
			continue
		}

		// Check for significant changes
		changes, err := loc.Detector.DetectChanges(ctx, weatherData)
		if err != nil {
			h.logger.Error().Err(err).Str("location", loc.Name).Msg("failed to detect weather changes")
			continue
		}

		// If no significant changes for this location, skip it
		if changes == "" {
			h.logger.Debug().Str("location", loc.Name).Msg("no significant weather changes detected")
			continue
		}

		hasChanges = true
		h.logger.Info().Str("location", loc.Name).Str("changes", changes).Msg("significant weather changes detected")

		// Generate LLM advice for the changes
		prompt := buildChangeAlertPromptFromResponse(loc.Name, changes, weatherData)
		advice, err := h.llmClient.GenerateAdvice(ctx, prompt)
		if err != nil {
			h.logger.Warn().Err(err).Str("location", loc.Name).Msg("failed to generate LLM advice, continuing without it")
			advice = "" // Continue without advice rather than failing
		}

		// Format alert message for this location
		message := FormatChangeAlert(loc.Name, changes, weatherData, advice)
		sections = append(sections, message)
	}

	// If no locations had changes, skip notification
	if !hasChanges {
		h.logger.Debug().Msg("no significant weather changes detected across all locations")
		return nil
	}

	// Combine all sections
	fullMessage := strings.Join(sections, "\n\n" + strings.Repeat("─", 30) + "\n\n")

	// Split message if needed for Telegram's 4096 char limit
	messages := splitMessage(fullMessage, 4096)

	// Send via Telegram
	for i, msg := range messages {
		if err := h.telegramClient.SendMessage(ctx, h.chatID, msg, "HTML"); err != nil {
			h.logger.Error().Err(err).Int("part", i+1).Msg("failed to send weather change alert")
			return fmt.Errorf("send telegram message part %d: %w", i+1, err)
		}
	}

	h.logger.Info().Int("locations_with_changes", len(sections)).Int("messages", len(messages)).Msg("weather change alert sent successfully")
	return nil
}

// splitMessage splits a message into chunks that fit Telegram's character limit
// Tries to split at section boundaries (separator lines) when possible
func splitMessage(message string, maxLen int) []string {
	if len(message) <= maxLen {
		return []string{message}
	}

	var messages []string
	separator := "\n\n" + strings.Repeat("─", 30) + "\n\n"

	// Try to split at separator boundaries
	parts := strings.Split(message, separator)

	currentMsg := ""
	for i, part := range parts {
		// If adding this part would exceed limit, save current and start new
		testMsg := currentMsg
		if currentMsg != "" && i < len(parts) {
			testMsg += separator
		}
		testMsg += part

		if len(testMsg) > maxLen && currentMsg != "" {
			// Save current message
			messages = append(messages, strings.TrimSpace(currentMsg))
			currentMsg = part
		} else {
			currentMsg = testMsg
		}
	}

	// Add final message
	if currentMsg != "" {
		messages = append(messages, strings.TrimSpace(currentMsg))
	}

	return messages
}

// Helper functions to build prompts from WeatherResponse
func buildMorningPromptFromResponse(locationName string, w *weather.WeatherResponse) string {
	realtime := w.Result.Realtime
	daily := w.Result.Daily

	var todayTemp weather.DailyTempPoint
	if len(daily.Temperature) > 0 {
		todayTemp = daily.Temperature[0]
	}

	return fmt.Sprintf(`位置：%s

今天早上的天气情况：
- 天气：%s
- 温度：当前 %.1f°C，最高 %.1f°C，最低 %.1f°C
- 湿度：%.0f%%
- 风速：%.1f m/s
- 空气质量：AQI %d

请为用户提供今天出门的建议，包括穿衣、携带物品等提醒。`,
		locationName,
		realtime.Skycon.Chinese(),
		realtime.Temperature,
		todayTemp.Max,
		todayTemp.Min,
		realtime.Humidity*100,
		realtime.Wind.Speed,
		realtime.AirQuality.AQI.CN,
	)
}

func buildEveningPromptFromResponse(locationName string, w *weather.WeatherResponse) string {
	realtime := w.Result.Realtime
	daily := w.Result.Daily

	var tomorrowTemp weather.DailyTempPoint
	var tomorrowSkycon weather.Skycon
	if len(daily.Temperature) > 1 {
		tomorrowTemp = daily.Temperature[1]
	}
	if len(daily.Skycon) > 1 {
		tomorrowSkycon = daily.Skycon[1].Value
	}

	return fmt.Sprintf(`位置：%s

今天晚上和明天的天气情况：
- 今晚天气：%s，温度：%.1f°C
- 明天天气：%s，温度：%.1f°C ~ %.1f°C
- 湿度：%.0f%%
- 风速：%.1f m/s

请为用户提供今晚和明天的天气提醒，以及休息时的注意事项。`,
		locationName,
		realtime.Skycon.Chinese(),
		realtime.Temperature,
		tomorrowSkycon.Chinese(),
		tomorrowTemp.Min,
		tomorrowTemp.Max,
		realtime.Humidity*100,
		realtime.Wind.Speed,
	)
}

func buildChangeAlertPromptFromResponse(locationName string, changes string, w *weather.WeatherResponse) string {
	realtime := w.Result.Realtime

	return fmt.Sprintf(`位置：%s

检测到显著的天气变化：
%s

当前天气情况：
- 天气：%s
- 温度：%.1f°C
- 湿度：%.0f%%
- 风速：%.1f m/s
- 空气质量：AQI %d

请提醒用户注意这些天气变化，并给出应对建议。`,
		locationName,
		changes,
		realtime.Skycon.Chinese(),
		realtime.Temperature,
		realtime.Humidity*100,
		realtime.Wind.Speed,
		realtime.AirQuality.AQI.CN,
	)
}
