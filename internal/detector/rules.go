package detector

import "github.com/user/weather-notice-bot/internal/weather"

// Thresholds defines the sensitivity thresholds for weather change detection
type Thresholds struct {
	// TemperatureDelta is the minimum temperature change in Celsius to trigger a notification
	TemperatureDelta float64

	// WindSpeedDelta is the minimum wind speed change in m/s to trigger a notification
	WindSpeedDelta float64

	// VisibilityDelta is the minimum visibility change in km to trigger a notification
	VisibilityDelta float64

	// AQICNDelta is the minimum AQI (China) change to trigger a notification
	AQICNDelta int

	// AQIUSADelta is the minimum AQI (USA) change to trigger a notification
	AQIUSADelta int
}

// DefaultThresholds returns recommended default thresholds
func DefaultThresholds() Thresholds {
	return Thresholds{
		TemperatureDelta: 3.0, // 3°C change
		WindSpeedDelta:   5.0, // 5 m/s change
		VisibilityDelta:  5.0, // 5 km change
		AQICNDelta:       50,  // 50 AQI points change
		AQIUSADelta:      50,  // 50 AQI points change
	}
}

// ChangeType represents the type of weather change detected
type ChangeType string

const (
	// SkyconChange indicates a change in weather condition (e.g., sunny to rainy)
	SkyconChange ChangeType = "skycon_change"

	// TemperatureChange indicates a significant temperature change
	TemperatureChange ChangeType = "temperature_change"

	// WindChange indicates a significant wind speed change
	WindChange ChangeType = "wind_change"

	// VisibilityChange indicates a significant visibility change
	VisibilityChange ChangeType = "visibility_change"

	// AQIChange indicates a significant air quality change
	AQIChange ChangeType = "aqi_change"
)

// SkyconSignificance represents the significance level of a skycon transition
type SkyconSignificance int

const (
	// NoSignificance means the change is not significant
	NoSignificance SkyconSignificance = 0

	// MinorSignificance means the change is minor (e.g., clear day to partly cloudy)
	MinorSignificance SkyconSignificance = 1

	// ModerateSignificance means the change is moderate (e.g., cloudy to light rain)
	ModerateSignificance SkyconSignificance = 2

	// MajorSignificance means the change is major (e.g., clear to heavy rain)
	MajorSignificance SkyconSignificance = 3
)

// GetSkyconSignificance returns the significance level of a skycon transition
func GetSkyconSignificance(from, to weather.Skycon) SkyconSignificance {
	// No change
	if from == to {
		return NoSignificance
	}

	// Day/night transitions of same condition are not significant
	if isDayNightTransition(from, to) {
		return NoSignificance
	}

	// Precipitation changes are highly significant
	if isPrecipitationChange(from, to) {
		return evaluatePrecipitationSignificance(from, to)
	}

	// Haze/fog changes are significant
	if isHazeOrFogChange(from, to) {
		return evaluateHazeSignificance(from, to)
	}

	// Clear to cloudy or vice versa
	if isClearToCloudyChange(from, to) {
		return MinorSignificance
	}

	// Default to minor significance for other transitions
	return MinorSignificance
}

// isDayNightTransition checks if the change is just a day/night transition of the same condition
func isDayNightTransition(from, to weather.Skycon) bool {
	pairs := [][2]weather.Skycon{
		{weather.ClearDay, weather.ClearNight},
		{weather.PartlyCloudyDay, weather.PartlyCloudyNight},
	}

	for _, pair := range pairs {
		if (from == pair[0] && to == pair[1]) || (from == pair[1] && to == pair[0]) {
			return true
		}
	}

	return false
}

// isPrecipitationChange checks if the change involves precipitation
func isPrecipitationChange(from, to weather.Skycon) bool {
	precipConditions := map[weather.Skycon]bool{
		weather.LightRain:    true,
		weather.ModerateRain: true,
		weather.HeavyRain:    true,
		weather.StormRain:    true,
		weather.LightSnow:    true,
		weather.ModerateSnow: true,
		weather.HeavySnow:    true,
		weather.StormSnow:    true,
	}

	fromPrecip := precipConditions[from]
	toPrecip := precipConditions[to]

	return fromPrecip != toPrecip || (fromPrecip && toPrecip)
}

// evaluatePrecipitationSignificance evaluates the significance of precipitation changes
func evaluatePrecipitationSignificance(from, to weather.Skycon) SkyconSignificance {
	// Define precipitation severity levels
	severityMap := map[weather.Skycon]int{
		weather.LightRain:    1,
		weather.ModerateRain: 2,
		weather.HeavyRain:    3,
		weather.StormRain:    4,
		weather.LightSnow:    1,
		weather.ModerateSnow: 2,
		weather.HeavySnow:    3,
		weather.StormSnow:    4,
	}

	fromSeverity := severityMap[from]
	toSeverity := severityMap[to]

	// Starting or stopping precipitation
	if fromSeverity == 0 || toSeverity == 0 {
		if toSeverity >= 3 || fromSeverity >= 3 {
			return MajorSignificance
		}
		return ModerateSignificance
	}

	// Intensity change within precipitation
	diff := abs(toSeverity - fromSeverity)
	if diff >= 2 {
		return MajorSignificance
	}
	if diff == 1 {
		return ModerateSignificance
	}

	return MinorSignificance
}

// isHazeOrFogChange checks if the change involves haze or fog
func isHazeOrFogChange(from, to weather.Skycon) bool {
	hazeConditions := map[weather.Skycon]bool{
		weather.LightHaze:    true,
		weather.ModerateHaze: true,
		weather.HeavyHaze:    true,
		weather.Fog:          true,
	}

	return hazeConditions[from] || hazeConditions[to]
}

// evaluateHazeSignificance evaluates the significance of haze/fog changes
func evaluateHazeSignificance(from, to weather.Skycon) SkyconSignificance {
	hazeSeverityMap := map[weather.Skycon]int{
		weather.LightHaze:    1,
		weather.ModerateHaze: 2,
		weather.HeavyHaze:    3,
		weather.Fog:          2,
	}

	fromSeverity := hazeSeverityMap[from]
	toSeverity := hazeSeverityMap[to]

	// Starting or clearing haze/fog
	if fromSeverity == 0 || toSeverity == 0 {
		if toSeverity >= 3 || fromSeverity >= 3 {
			return MajorSignificance
		}
		return ModerateSignificance
	}

	// Intensity change
	diff := abs(toSeverity - fromSeverity)
	if diff >= 2 {
		return ModerateSignificance
	}

	return MinorSignificance
}

// isClearToCloudyChange checks if the change is between clear and cloudy conditions
func isClearToCloudyChange(from, to weather.Skycon) bool {
	clearConditions := map[weather.Skycon]bool{
		weather.ClearDay:          true,
		weather.ClearNight:        true,
		weather.PartlyCloudyDay:   true,
		weather.PartlyCloudyNight: true,
	}

	cloudyConditions := map[weather.Skycon]bool{
		weather.Cloudy: true,
	}

	return (clearConditions[from] && cloudyConditions[to]) ||
		(cloudyConditions[from] && clearConditions[to])
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
