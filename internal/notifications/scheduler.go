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
	handlers *NotificationHandlers
	logger   zerolog.Logger
}

// NewScheduler creates a new notification scheduler with configured cron jobs
func NewScheduler(cfg *config.Config, handlers *NotificationHandlers, logger zerolog.Logger) (*Scheduler, error) {
	// Load timezone
	location, err := time.LoadLocation(cfg.Schedule.Timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %s: %w", cfg.Schedule.Timezone, err)
	}

	// Create cron instance with timezone (without verbose logger to avoid interface issues)
	c := cron.New(cron.WithLocation(location))

	scheduler := &Scheduler{
		cron:     c,
		handlers: handlers,
		logger:   logger.With().Str("component", "scheduler").Logger(),
	}

	// Add morning notification job (8:00 AM)
	_, err = c.AddFunc("0 8 * * *", func() {
		scheduler.logger.Info().Msg("triggering morning notification")
		if err := handlers.HandleMorningNotification(); err != nil {
			scheduler.logger.Error().Err(err).Msg("morning notification failed")
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add morning notification job: %w", err)
	}

	// Add evening notification job (11:30 PM)
	_, err = c.AddFunc("30 23 * * *", func() {
		scheduler.logger.Info().Msg("triggering evening notification")
		if err := handlers.HandleEveningNotification(); err != nil {
			scheduler.logger.Error().Err(err).Msg("evening notification failed")
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add evening notification job: %w", err)
	}

	// Add weather polling job (every 15 minutes)
	_, err = c.AddFunc("*/15 * * * *", func() {
		scheduler.logger.Debug().Msg("triggering weather poll")
		if err := handlers.HandleWeatherPoll(); err != nil {
			scheduler.logger.Error().Err(err).Msg("weather poll failed")
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add weather poll job: %w", err)
	}

	scheduler.logger.Info().
		Str("timezone", cfg.Schedule.Timezone).
		Msg("scheduler configured with 3 jobs: morning (8:00), evening (23:30), poll (*/15)")

	return scheduler, nil
}

// Start begins the scheduled job execution
func (s *Scheduler) Start() {
	s.logger.Info().Msg("starting scheduler")
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
