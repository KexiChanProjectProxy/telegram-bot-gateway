package notifications

import (
	"fmt"
	"time"

	"github.com/user/weather-notice-bot/internal/config"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
)

// Scheduler manages scheduled weather notifications
type Scheduler struct {
	cron     *cron.Cron
	handlers []*ChatHandler
	logger   zerolog.Logger
	location *time.Location
}

// NewScheduler creates a new notification scheduler with time-based schedules
func NewScheduler(cfg *config.Config, handlers []*ChatHandler, logger zerolog.Logger) (*Scheduler, error) {
	// Load timezone
	location, err := time.LoadLocation(cfg.Schedule.Timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %s: %w", cfg.Schedule.Timezone, err)
	}

	// Create cron instance with timezone and seconds support
	c := cron.New(cron.WithLocation(location), cron.WithSeconds())

	scheduler := &Scheduler{
		cron:     c,
		handlers: handlers,
		location: location,
		logger:   logger.With().Str("component", "scheduler").Logger(),
	}

	// Parse and add morning notification job
	morningCron, err := timeStringToCron(cfg.Schedule.MorningTime)
	if err != nil {
		return nil, fmt.Errorf("invalid morning_time format: %w", err)
	}
	_, err = c.AddFunc(morningCron, func() {
		scheduler.logger.Info().Msg("triggering morning notification")
		for _, handler := range handlers {
			if err := handler.HandleMorningNotification(); err != nil {
				scheduler.logger.Error().Err(err).Int64("chat_id", handler.chatID).Msg("morning notification failed")
			}
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add morning notification job: %w", err)
	}

	// Parse and add evening notification job
	eveningCron, err := timeStringToCron(cfg.Schedule.EveningTime)
	if err != nil {
		return nil, fmt.Errorf("invalid evening_time format: %w", err)
	}
	_, err = c.AddFunc(eveningCron, func() {
		scheduler.logger.Info().Msg("triggering evening notification")
		for _, handler := range handlers {
			if err := handler.HandleEveningNotification(); err != nil {
				scheduler.logger.Error().Err(err).Int64("chat_id", handler.chatID).Msg("evening notification failed")
			}
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add evening notification job: %w", err)
	}

	// Parse and add weather polling job
	pollCron, err := durationToCron(cfg.Schedule.PollInterval)
	if err != nil {
		return nil, fmt.Errorf("invalid poll_interval format: %w", err)
	}
	_, err = c.AddFunc(pollCron, func() {
		scheduler.logger.Debug().Msg("triggering weather poll")
		for _, handler := range handlers {
			if err := handler.HandleWeatherPoll(); err != nil {
				scheduler.logger.Error().Err(err).Int64("chat_id", handler.chatID).Msg("weather poll failed")
			}
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add weather poll job: %w", err)
	}

	scheduler.logger.Info().
		Str("timezone", cfg.Schedule.Timezone).
		Str("morning_time", cfg.Schedule.MorningTime).
		Str("evening_time", cfg.Schedule.EveningTime).
		Str("poll_interval", cfg.Schedule.PollInterval).
		Int("chat_handlers", len(handlers)).
		Msg("scheduler configured with 3 jobs")

	return scheduler, nil
}

// Start begins the scheduled job execution and sends initial weather updates
func (s *Scheduler) Start() {
	s.logger.Info().Msg("starting scheduler")

	// Send initial weather update to all chats on startup
	s.logger.Info().Msg("sending initial weather update on startup")
	for _, handler := range s.handlers {
		if err := handler.HandleMorningNotification(); err != nil {
			s.logger.Error().Err(err).Int64("chat_id", handler.chatID).Msg("startup weather notification failed")
		}
	}

	s.cron.Start()
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop() {
	s.logger.Info().Msg("stopping scheduler")
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.logger.Info().Msg("scheduler stopped")
}

// GetCron returns the underlying cron instance for testing or advanced usage
func (s *Scheduler) GetCron() *cron.Cron {
	return s.cron
}

// timeStringToCron converts a time string (HH:MM:SS) to cron expression
// Examples: "08:00:00" -> "0 0 8 * * *", "23:30:00" -> "0 30 23 * * *"
func timeStringToCron(timeStr string) (string, error) {
	// Parse time in format HH:MM:SS
	t, err := time.Parse("15:04:05", timeStr)
	if err != nil {
		// Try HH:MM format
		t, err = time.Parse("15:04", timeStr)
		if err != nil {
			return "", fmt.Errorf("time must be in HH:MM:SS or HH:MM format: %w", err)
		}
	}

	// Convert to cron format: "second minute hour * * *"
	return fmt.Sprintf("%d %d %d * * *", t.Second(), t.Minute(), t.Hour()), nil
}

// durationToCron converts a duration string to cron expression for polling
// Examples: "15m" -> "0 */15 * * * *", "1h" -> "0 0 * * * *"
func durationToCron(durationStr string) (string, error) {
	d, err := time.ParseDuration(durationStr)
	if err != nil {
		return "", fmt.Errorf("invalid duration format: %w", err)
	}

	// Convert duration to appropriate cron expression (with seconds field)
	minutes := int(d.Minutes())
	if minutes < 1 {
		return "", fmt.Errorf("poll interval must be at least 1 minute")
	}

	if minutes < 60 {
		// Every N minutes: "0 */N * * * *"
		return fmt.Sprintf("0 */%d * * * *", minutes), nil
	}

	// For hourly or longer intervals
	hours := int(d.Hours())
	if hours < 24 {
		// Every N hours: "0 0 */N * * *"
		return fmt.Sprintf("0 0 */%d * * *", hours), nil
	}

	// For daily intervals: "0 0 0 */N * *"
	days := hours / 24
	return fmt.Sprintf("0 0 0 */%d * *", days), nil
}
