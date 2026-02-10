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
}

// NewScheduler creates a new notification scheduler with configured cron jobs
func NewScheduler(cfg *config.Config, handlers []*ChatHandler, logger zerolog.Logger) (*Scheduler, error) {
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

	// Add morning notification job
	_, err = c.AddFunc(cfg.Schedule.MorningCron, func() {
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

	// Add evening notification job
	_, err = c.AddFunc(cfg.Schedule.EveningCron, func() {
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

	// Add weather polling job
	_, err = c.AddFunc(cfg.Schedule.PollCron, func() {
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
		Str("morning_cron", cfg.Schedule.MorningCron).
		Str("evening_cron", cfg.Schedule.EveningCron).
		Str("poll_cron", cfg.Schedule.PollCron).
		Int("chat_handlers", len(handlers)).
		Msg("scheduler configured with 3 jobs")

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
