package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	baseURL      = "https://api.caiyunapp.com/v2.6"
	hourlySteps  = 360 // 15 days of hourly data
	dailySteps   = 15  // 15 days of daily forecast
)

// Client is the Caiyun Weather API client
type Client struct {
	token      string
	httpClient *http.Client
}

// NewClient creates a new Caiyun Weather API client
func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetWeather fetches weather data for the given coordinates
func (c *Client) GetWeather(ctx context.Context, lon, lat float64) (*WeatherResponse, error) {
	url := fmt.Sprintf("%s/%s/%.6f,%.6f/weather.json?hourlysteps=%d&dailysteps=%d",
		baseURL, c.token, lon, lat, hourlySteps, dailySteps)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var weatherResp WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if weatherResp.Status != "ok" {
		return nil, fmt.Errorf("API returned error status: %s", weatherResp.Status)
	}

	return &weatherResp, nil
}
