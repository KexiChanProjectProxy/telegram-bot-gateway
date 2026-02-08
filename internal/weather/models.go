package weather

import "time"

// RealtimeWeather represents the current weather conditions
type RealtimeWeather struct {
	Temperature float64 `json:"temperature"`
	Skycon      Skycon  `json:"skycon"`
	Humidity    float64 `json:"humidity"`
	Wind        Wind    `json:"wind"`
	Visibility  float64 `json:"visibility"`
	AQI         AQI     `json:"air_quality"`
}

// Wind represents wind information
type Wind struct {
	Speed     float64 `json:"speed"`
	Direction float64 `json:"direction"`
}

// AQI represents air quality index information
type AQI struct {
	CN  int    `json:"chn"`
	USA int    `json:"usa"`
}

// HourlyForecast represents hourly weather forecast
type HourlyForecast struct {
	Datetime    time.Time `json:"datetime"`
	Temperature float64   `json:"temperature"`
	Skycon      Skycon    `json:"skycon"`
	Humidity    float64   `json:"humidity"`
	Wind        Wind      `json:"wind"`
}

// DailyForecast represents daily weather forecast
type DailyForecast struct {
	Date        string  `json:"date"`
	SkyconDay   Skycon  `json:"skycon_day"`
	SkyconNight Skycon  `json:"skycon_night"`
	TempMax     float64 `json:"temp_max"`
	TempMin     float64 `json:"temp_min"`
	Humidity    float64 `json:"humidity"`
	Wind        Wind    `json:"wind"`
	AQI         AQI     `json:"air_quality"`
}

// WeatherResponse matches Caiyun API v2.6 response format
type WeatherResponse struct {
	Status      string          `json:"status"`
	APIVersion  string          `json:"api_version"`
	APIStatus   string          `json:"api_status"`
	Lang        string          `json:"lang"`
	Unit        string          `json:"unit"`
	Tzshift     int             `json:"tzshift"`
	Timezone    string          `json:"timezone"`
	ServerTime  int64           `json:"server_time"`
	Location    []float64       `json:"location"`
	Result      WeatherResult   `json:"result"`
}

// WeatherResult contains the actual weather data
type WeatherResult struct {
	Realtime RealtimeResult `json:"realtime"`
	Hourly   HourlyResult   `json:"hourly"`
	Daily    DailyResult    `json:"daily"`
}

// RealtimeResult contains realtime weather data from API
type RealtimeResult struct {
	Status      string       `json:"status"`
	Temperature float64      `json:"temperature"`
	Humidity    float64      `json:"humidity"`
	Skycon      Skycon       `json:"skycon"`
	Visibility  float64      `json:"visibility"`
	Wind        Wind         `json:"wind"`
	AirQuality  AirQuality   `json:"air_quality"`
}

// HourlyResult contains hourly forecast data
type HourlyResult struct {
	Status      string          `json:"status"`
	Description string          `json:"description"`
	Temperature []ValuePoint    `json:"temperature"`
	Skycon      []SkyconPoint   `json:"skycon"`
	Humidity    []ValuePoint    `json:"humidity"`
	Wind        []WindPoint     `json:"wind"`
}

// DailyResult contains daily forecast data
type DailyResult struct {
	Status      string             `json:"status"`
	Skycon      []DailySkyconPoint `json:"skycon"`
	Temperature []DailyTempPoint   `json:"temperature"`
	Humidity    []ValuePoint       `json:"humidity"`
	Wind        []WindPoint        `json:"wind"`
	AirQuality  AirQualityDaily    `json:"air_quality"`
}

// ValuePoint represents a time-series data point with a single value
type ValuePoint struct {
	Datetime time.Time `json:"datetime"`
	Value    float64   `json:"value"`
}

// SkyconPoint represents a time-series skycon data point
type SkyconPoint struct {
	Datetime time.Time `json:"datetime"`
	Value    Skycon    `json:"value"`
}

// WindPoint represents a time-series wind data point
type WindPoint struct {
	Datetime  time.Time `json:"datetime"`
	Speed     float64   `json:"speed"`
	Direction float64   `json:"direction"`
}

// DailySkyconPoint represents daily skycon forecast
type DailySkyconPoint struct {
	Date  string `json:"date"`
	Value Skycon `json:"value"`
}

// DailyTempPoint represents daily temperature forecast
type DailyTempPoint struct {
	Date string  `json:"date"`
	Max  float64 `json:"max"`
	Min  float64 `json:"min"`
	Avg  float64 `json:"avg"`
}

// AirQuality represents air quality information
type AirQuality struct {
	PM25 float64 `json:"pm25"`
	PM10 float64 `json:"pm10"`
	O3   float64 `json:"o3"`
	SO2  float64 `json:"so2"`
	NO2  float64 `json:"no2"`
	CO   float64 `json:"co"`
	AQI  AQI     `json:"aqi"`
}

// AirQualityDaily represents daily air quality forecast
type AirQualityDaily struct {
	AQI []DailyAQIPoint `json:"aqi"`
}

// DailyAQIPoint represents daily AQI forecast
type DailyAQIPoint struct {
	Date string `json:"date"`
	Max  AQI    `json:"max"`
	Min  AQI    `json:"min"`
	Avg  AQI    `json:"avg"`
}
