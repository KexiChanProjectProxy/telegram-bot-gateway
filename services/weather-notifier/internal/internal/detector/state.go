package detector

import (
	"encoding/json"
	"os"
	"time"

	"github.com/user/weather-notice-bot/internal/weather"
)

// WeatherState represents the last known weather conditions
type WeatherState struct {
	// LastUpdated is the timestamp when this state was recorded
	LastUpdated time.Time `json:"last_updated"`

	// Temperature in Celsius
	Temperature float64 `json:"temperature"`

	// Skycon represents the weather condition
	Skycon weather.Skycon `json:"skycon"`

	// Humidity percentage (0-100)
	Humidity float64 `json:"humidity"`

	// Wind speed in m/s
	WindSpeed float64 `json:"wind_speed"`

	// Wind direction in degrees (0-360)
	WindDirection float64 `json:"wind_direction"`

	// Visibility in km
	Visibility float64 `json:"visibility"`

	// AQI China standard
	AQICN int `json:"aqi_cn"`

	// AQI USA standard
	AQIUSA int `json:"aqi_usa"`
}

// NewWeatherState creates a WeatherState from RealtimeWeather
func NewWeatherState(realtime weather.RealtimeWeather) WeatherState {
	return WeatherState{
		LastUpdated:   time.Now(),
		Temperature:   realtime.Temperature,
		Skycon:        realtime.Skycon,
		Humidity:      realtime.Humidity,
		WindSpeed:     realtime.Wind.Speed,
		WindDirection: realtime.Wind.Direction,
		Visibility:    realtime.Visibility,
		AQICN:         realtime.AQI.CN,
		AQIUSA:        realtime.AQI.USA,
	}
}

// LoadState loads the weather state from a JSON file
// Returns nil if the file doesn't exist or is invalid
func LoadState(filePath string) (*WeatherState, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // File doesn't exist yet, not an error
		}
		return nil, err
	}

	var state WeatherState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// SaveState saves the weather state to a JSON file
func SaveState(filePath string, state WeatherState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}
