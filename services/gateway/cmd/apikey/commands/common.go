package commands

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/gorm"

	"github.com/kexi/telegram-bot-gateway/internal/config"
	"github.com/kexi/telegram-bot-gateway/internal/pkg/apikey"
	"github.com/kexi/telegram-bot-gateway/internal/repository"
)

// initDB initializes the database connection
func initDB() (*gorm.DB, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.json"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	db, err := repository.NewDatabase(&cfg.Database, "release")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

// getContext returns a context with timeout
func getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

// fatal logs an error and exits
func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}

// success prints a success message
func success(format string, args ...interface{}) {
	fmt.Printf("✓ "+format+"\n", args...)
}

// info prints an info message
func info(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// parseDuration parses duration strings like "1y", "30d", "24h"
func parseDuration(s string) (time.Duration, error) {
	if len(s) < 2 {
		return 0, fmt.Errorf("invalid duration: %s", s)
	}

	unit := s[len(s)-1]
	value := s[:len(s)-1]

	var multiplier time.Duration
	switch unit {
	case 'y':
		multiplier = 365 * 24 * time.Hour
	case 'M':
		multiplier = 30 * 24 * time.Hour
	case 'd':
		multiplier = 24 * time.Hour
	case 'h':
		multiplier = time.Hour
	case 'm':
		multiplier = time.Minute
	case 's':
		multiplier = time.Second
	default:
		return 0, fmt.Errorf("invalid duration unit: %c (use y, M, d, h, m, s)", unit)
	}

	var count int
	_, err := fmt.Sscanf(value, "%d", &count)
	if err != nil {
		return 0, fmt.Errorf("invalid duration value: %s", value)
	}

	return time.Duration(count) * multiplier, nil
}

// initRepositories initializes all required repositories
func initRepositories(db *gorm.DB) (
	apiKeyRepo repository.APIKeyRepository,
	chatPermRepo repository.ChatPermissionRepository,
	botPermRepo repository.APIKeyBotPermissionRepository,
	feedbackPermRepo repository.APIKeyFeedbackPermissionRepository,
	chatRepo repository.ChatRepository,
	botRepo repository.BotRepository,
) {
	apiKeyRepo = repository.NewAPIKeyRepository(db)
	chatPermRepo = repository.NewChatPermissionRepository(db)
	botPermRepo = repository.NewAPIKeyBotPermissionRepository(db)
	feedbackPermRepo = repository.NewAPIKeyFeedbackPermissionRepository(db)
	chatRepo = repository.NewChatRepository(db)
	botRepo = repository.NewBotRepository(db)
	return
}

// initAPIKeyService initializes the API key service
func initAPIKeyService() *apikey.Service {
	// Default values - could be loaded from config if needed
	prefix := "tgw_"
	length := 32
	return apikey.NewService(prefix, length)
}

// printAPIKeyTable prints a formatted table of API keys
func printAPIKeyTable(keys []struct {
	ID          uint
	Key         string
	Name        string
	RateLimit   int
	IsActive    bool
	ExpiresAt   *time.Time
	LastUsedAt  *time.Time
	CreatedAt   time.Time
}) {
	if len(keys) == 0 {
		info("No API keys found")
		return
	}

	fmt.Println("\nID | Key Prefix       | Name                 | Rate Limit | Active | Expires           | Last Used         | Created")
	fmt.Println("---|------------------|----------------------|------------|--------|-------------------|-------------------|-------------------")

	for _, k := range keys {
		keyPrefix := k.Key
		if len(keyPrefix) > 16 {
			keyPrefix = keyPrefix[:13] + "..."
		}

		name := k.Name
		if len(name) > 20 {
			name = name[:17] + "..."
		}

		active := "Yes"
		if !k.IsActive {
			active = "No"
		}

		expires := "Never"
		if k.ExpiresAt != nil {
			expires = k.ExpiresAt.Format("2006-01-02 15:04")
		}

		lastUsed := "Never"
		if k.LastUsedAt != nil {
			lastUsed = k.LastUsedAt.Format("2006-01-02 15:04")
		}

		created := k.CreatedAt.Format("2006-01-02 15:04")

		fmt.Printf("%-2d | %-16s | %-20s | %-10d | %-6s | %-17s | %-17s | %s\n",
			k.ID, keyPrefix, name, k.RateLimit, active, expires, lastUsed, created)
	}
	fmt.Println()
}

// hasFlag checks if a flag is present in args
func hasFlag(args []string, flag ...string) bool {
	for _, arg := range args {
		for _, f := range flag {
			if arg == f {
				return true
			}
		}
	}
	return false
}

// getFlagValue gets the value of a flag
func getFlagValue(args []string, flag ...string) string {
	for i, arg := range args {
		for _, f := range flag {
			if arg == f && i+1 < len(args) {
				return args[i+1]
			}
		}
	}
	return ""
}

func init() {
	// Disable log timestamps for cleaner output
	log.SetFlags(0)
}
