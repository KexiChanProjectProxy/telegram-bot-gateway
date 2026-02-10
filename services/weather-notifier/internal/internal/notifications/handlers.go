package notifications

import (
	"context"
	"fmt"

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

// NotificationHandlers manages weather notification operations
type NotificationHandlers struct {
	weatherClient  *weather.Client
	telegramClient *telegram.Client
	llmClient      *llm.Client
	detector       WeatherDetector
	chatID         int64
	latitude       float64
	longitude      float64
	logger         zerolog.Logger
}

// NewNotificationHandlers creates a new notification handlers instance
func NewNotificationHandlers(
	weatherClient *weather.Client,
	telegramClient *telegram.Client,
	llmClient *llm.Client,
	detector WeatherDetector,
	chatID int64,
	latitude float64,
	longitude float64,
	logger zerolog.Logger,
) *NotificationHandlers {
	return &NotificationHandlers{
		weatherClient:  weatherClient,
		telegramClient: telegramClient,
		llmClient:      llmClient,
		detector:       detector,
		chatID:         chatID,
		latitude:       latitude,
		longitude:      longitude,
		logger:         logger.With().Str("component", "notification_handlers").Logger(),
	}
}

// HandleMorningNotification sends the morning weather notification
func (h *NotificationHandlers) HandleMorningNotification() error {
	ctx := context.Background()

	h.logger.Info().Msg("handling morning notification")

	// Fetch current weather
	weatherData, err := h.weatherClient.GetWeather(ctx, h.longitude, h.latitude)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to fetch weather data")
		return fmt.Errorf("fetch weather: %w", err)
	}

	// Generate LLM advice
	prompt := buildMorningPromptFromResponse(weatherData)
	advice, err := h.llmClient.GenerateAdvice(ctx, prompt)
	if err != nil {
		h.logger.Warn().Err(err).Msg("failed to generate LLM advice, continuing without it")
		advice = "" // Continue without advice rather than failing
	}

	// Format message
	message := FormatMorningMessage(weatherData, advice)

	// Send via Telegram
	if err := h.telegramClient.SendMessage(ctx, h.chatID, message, "HTML"); err != nil {
		h.logger.Error().Err(err).Msg("failed to send morning notification")
		return fmt.Errorf("send telegram message: %w", err)
	}

	h.logger.Info().Msg("morning notification sent successfully")
	return nil
}

// HandleEveningNotification sends the evening weather notification
func (h *NotificationHandlers) HandleEveningNotification() error {
	ctx := context.Background()

	h.logger.Info().Msg("handling evening notification")

	// Fetch current weather
	weatherData, err := h.weatherClient.GetWeather(ctx, h.longitude, h.latitude)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to fetch weather data")
		return fmt.Errorf("fetch weather: %w", err)
	}

	// Generate LLM advice
	prompt := buildEveningPromptFromResponse(weatherData)
	advice, err := h.llmClient.GenerateAdvice(ctx, prompt)
	if err != nil {
		h.logger.Warn().Err(err).Msg("failed to generate LLM advice, continuing without it")
		advice = "" // Continue without advice rather than failing
	}

	// Format message
	message := FormatEveningMessage(weatherData, advice)

	// Send via Telegram
	if err := h.telegramClient.SendMessage(ctx, h.chatID, message, "HTML"); err != nil {
		h.logger.Error().Err(err).Msg("failed to send evening notification")
		return fmt.Errorf("send telegram message: %w", err)
	}

	h.logger.Info().Msg("evening notification sent successfully")
	return nil
}

// HandleWeatherPoll checks for weather changes and sends alerts if needed
func (h *NotificationHandlers) HandleWeatherPoll() error {
	ctx := context.Background()

	h.logger.Debug().Msg("polling weather for changes")

	// Fetch current weather
	weatherData, err := h.weatherClient.GetWeather(ctx, h.longitude, h.latitude)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to fetch weather data")
		return fmt.Errorf("fetch weather: %w", err)
	}

	// Check for significant changes
	changes, err := h.detector.DetectChanges(ctx, weatherData)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to detect weather changes")
		return fmt.Errorf("detect changes: %w", err)
	}

	// If no significant changes, skip notification
	if changes == "" {
		h.logger.Debug().Msg("no significant weather changes detected")
		return nil
	}

	h.logger.Info().Str("changes", changes).Msg("significant weather changes detected")

	// Generate LLM advice for the changes
	prompt := buildChangeAlertPromptFromResponse(changes, weatherData)
	advice, err := h.llmClient.GenerateAdvice(ctx, prompt)
	if err != nil {
		h.logger.Warn().Err(err).Msg("failed to generate LLM advice, continuing without it")
		advice = "" // Continue without advice rather than failing
	}

	// Format alert message
	message := FormatChangeAlert(changes, weatherData, advice)

	// Send via Telegram
	if err := h.telegramClient.SendMessage(ctx, h.chatID, message, "HTML"); err != nil {
		h.logger.Error().Err(err).Msg("failed to send weather change alert")
		return fmt.Errorf("send telegram message: %w", err)
	}

	h.logger.Info().Msg("weather change alert sent successfully")
	return nil
}

// Helper functions to build prompts from WeatherResponse
func buildMorningPromptFromResponse(w *weather.WeatherResponse) string {
	realtime := w.Result.Realtime
	daily := w.Result.Daily

	var todayTemp weather.DailyTempPoint
	if len(daily.Temperature) > 0 {
		todayTemp = daily.Temperature[0]
	}

	return fmt.Sprintf(`今天早上的天气情况：
- 天气：%s
- 温度：当前 %.1f°C，最高 %.1f°C，最低 %.1f°C
- 湿度：%.0f%%
- 风速：%.1f m/s
- 空气质量：AQI %d

请为用户提供今天出门的建议，包括穿衣、携带物品等提醒。`,
		realtime.Skycon.Chinese(),
		realtime.Temperature,
		todayTemp.Max,
		todayTemp.Min,
		realtime.Humidity*100,
		realtime.Wind.Speed,
		realtime.AirQuality.AQI.CN,
	)
}

func buildEveningPromptFromResponse(w *weather.WeatherResponse) string {
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

	return fmt.Sprintf(`今天晚上和明天的天气情况：
- 今晚天气：%s，温度：%.1f°C
- 明天天气：%s，温度：%.1f°C ~ %.1f°C
- 湿度：%.0f%%
- 风速：%.1f m/s

请为用户提供今晚和明天的天气提醒，以及休息时的注意事项。`,
		realtime.Skycon.Chinese(),
		realtime.Temperature,
		tomorrowSkycon.Chinese(),
		tomorrowTemp.Min,
		tomorrowTemp.Max,
		realtime.Humidity*100,
		realtime.Wind.Speed,
	)
}

func buildChangeAlertPromptFromResponse(changes string, w *weather.WeatherResponse) string {
	realtime := w.Result.Realtime

	return fmt.Sprintf(`检测到显著的天气变化：
%s

当前天气情况：
- 天气：%s
- 温度：%.1f°C
- 湿度：%.0f%%
- 风速：%.1f m/s
- 空气质量：AQI %d

请提醒用户注意这些天气变化，并给出应对建议。`,
		changes,
		realtime.Skycon.Chinese(),
		realtime.Temperature,
		realtime.Humidity*100,
		realtime.Wind.Speed,
		realtime.AirQuality.AQI.CN,
	)
}
