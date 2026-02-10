package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/user/weather-notice-bot/internal/detector"
	"github.com/user/weather-notice-bot/internal/weather"

	"github.com/rs/zerolog"
)

// DetectorAdapter adapts the detector package to the WeatherDetector interface
// expected by the notification handlers. It manages state persistence and
// formats changes into a human-readable description.
type DetectorAdapter struct {
	detector  *detector.ChangeDetector
	stateFile string
	logger    zerolog.Logger
}

// NewDetectorAdapter creates a new detector adapter
func NewDetectorAdapter(thresholds detector.Thresholds, stateFile string, logger zerolog.Logger) *DetectorAdapter {
	return &DetectorAdapter{
		detector:  detector.NewChangeDetector(thresholds),
		stateFile: stateFile,
		logger:    logger.With().Str("component", "detector_adapter").Logger(),
	}
}

// DetectChanges implements the WeatherDetector interface
// It compares current weather with the last known state and returns a description
// of significant changes, or empty string if no significant changes detected.
func (da *DetectorAdapter) DetectChanges(ctx context.Context, current *weather.WeatherResponse) (string, error) {
	// Load previous state
	oldState, err := detector.LoadState(da.stateFile)
	if err != nil {
		da.logger.Error().Err(err).Msg("failed to load previous weather state")
		return "", fmt.Errorf("load previous state: %w", err)
	}

	// Convert RealtimeResult to RealtimeWeather
	realtimeWeather := convertToRealtimeWeather(current.Result.Realtime)

	// If no previous state exists, save current state and return no changes
	if oldState == nil {
		da.logger.Info().Msg("no previous weather state found, initializing")
		newState := detector.NewWeatherState(realtimeWeather)
		if err := detector.SaveState(da.stateFile, newState); err != nil {
			da.logger.Warn().Err(err).Msg("failed to save initial weather state")
		}
		return "", nil
	}

	// Detect changes
	changes := da.detector.DetectChanges(*oldState, realtimeWeather)

	// If no significant changes, return empty string
	if len(changes) == 0 {
		da.logger.Debug().Msg("no significant weather changes detected")
		return "", nil
	}

	// Save new state
	newState := detector.NewWeatherState(realtimeWeather)
	if err := detector.SaveState(da.stateFile, newState); err != nil {
		da.logger.Warn().Err(err).Msg("failed to save new weather state")
		// Don't fail the detection just because we couldn't save state
	}

	// Format changes into a human-readable description
	description := formatChanges(changes)
	da.logger.Info().
		Int("change_count", len(changes)).
		Str("description", description).
		Msg("significant weather changes detected")

	return description, nil
}

// convertToRealtimeWeather converts RealtimeResult to RealtimeWeather
func convertToRealtimeWeather(rt weather.RealtimeResult) weather.RealtimeWeather {
	return weather.RealtimeWeather{
		Temperature: rt.Temperature,
		Skycon:      rt.Skycon,
		Humidity:    rt.Humidity,
		Wind:        rt.Wind,
		Visibility:  rt.Visibility,
		AQI:         rt.AirQuality.AQI,
	}
}

// formatChanges converts a slice of WeatherChange into a formatted string
func formatChanges(changes []detector.WeatherChange) string {
	if len(changes) == 0 {
		return ""
	}

	var parts []string
	for _, change := range changes {
		parts = append(parts, change.Description)
	}

	return strings.Join(parts, "\n")
}

// GetStateFilePath returns the absolute path for the weather state file
func GetStateFilePath(dataDir string) string {
	return filepath.Join(dataDir, "weather_state.json")
}
