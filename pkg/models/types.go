package models

import "time"

// NotificationLevel represents the severity level of a weather notification
type NotificationLevel string

const (
	NotificationLevelInfo    NotificationLevel = "info"
	NotificationLevelWarning NotificationLevel = "warning"
	NotificationLevelCritical NotificationLevel = "critical"
)

// WeatherAlert represents a weather alert notification
type WeatherAlert struct {
	ID        string            `json:"id"`
	Level     NotificationLevel `json:"level"`
	Title     string            `json:"title"`
	Message   string            `json:"message"`
	Timestamp time.Time         `json:"timestamp"`
	Location  Location          `json:"location"`
	Conditions []AlertCondition  `json:"conditions"`
}

// Location represents a geographic location
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Name      string  `json:"name"`
}

// AlertCondition represents a specific weather condition that triggered an alert
type AlertCondition struct {
	Type        string  `json:"type"`        // e.g., "rain", "temperature", "aqi"
	Description string  `json:"description"` // Human-readable description
	Value       float64 `json:"value"`       // Current value
	Threshold   float64 `json:"threshold"`   // Threshold that was crossed
}

// NotificationRequest represents a request to send a notification
type NotificationRequest struct {
	ChatID  int64  `json:"chat_id"`
	Message string `json:"message"`
}

// ServiceStatus represents the overall status of the service
type ServiceStatus struct {
	Healthy          bool      `json:"healthy"`
	LastWeatherCheck time.Time `json:"last_weather_check"`
	LastNotification time.Time `json:"last_notification"`
	Uptime           string    `json:"uptime"`
	Version          string    `json:"version"`
}

// UserSession represents an authenticated user session
type UserSession struct {
	UserID        int64     `json:"user_id"`
	Authenticated bool      `json:"authenticated"`
	LastActive    time.Time `json:"last_active"`
}
