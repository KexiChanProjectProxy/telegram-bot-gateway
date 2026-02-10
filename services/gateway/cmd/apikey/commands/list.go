package commands

import (
	"fmt"
	"time"
)

const listUsage = `List all API keys

Usage:
  apikey list [options]

Options:
  --format <format>         Output format: table (default), json
  --help, -h                Show this help message

Examples:
  apikey list
  apikey list --format json
`

func ListAPIKeys(args []string) {
	if hasFlag(args, "--help", "-h") {
		fmt.Println(listUsage)
		return
	}

	format := getFlagValue(args, "--format", "-f")
	if format == "" {
		format = "table"
	}

	// Initialize database
	db, err := initDB()
	if err != nil {
		fatal("Failed to initialize database: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	apiKeyRepo, _, _, _, _, _ := initRepositories(db)

	// Get all API keys
	ctx, cancel := getContext()
	defer cancel()

	apiKeys, err := apiKeyRepo.List(ctx, 0, 1000)
	if err != nil {
		fatal("Failed to list API keys: %v", err)
	}

	if format == "json" {
		// Simple JSON output
		fmt.Println("[")
		for i, key := range apiKeys {
			expires := "null"
			if key.ExpiresAt != nil {
				expires = fmt.Sprintf("\"%s\"", key.ExpiresAt.Format("2006-01-02T15:04:05Z"))
			}
			lastUsed := "null"
			if key.LastUsedAt != nil {
				lastUsed = fmt.Sprintf("\"%s\"", key.LastUsedAt.Format("2006-01-02T15:04:05Z"))
			}

			fmt.Printf("  {\"id\":%d,\"key\":\"%s\",\"name\":\"%s\",\"rate_limit\":%d,\"is_active\":%t,\"expires_at\":%s,\"last_used_at\":%s,\"created_at\":\"%s\"}",
				key.ID, key.Key, key.Name, key.RateLimit, key.IsActive, expires, lastUsed, key.CreatedAt.Format("2006-01-02T15:04:05Z"))
			if i < len(apiKeys)-1 {
				fmt.Println(",")
			} else {
				fmt.Println()
			}
		}
		fmt.Println("]")
	} else {
		// Table format
		var tableData []struct {
			ID         uint
			Key        string
			Name       string
			RateLimit  int
			IsActive   bool
			ExpiresAt  *time.Time
			LastUsedAt *time.Time
			CreatedAt  time.Time
		}

		for _, key := range apiKeys {
			tableData = append(tableData, struct {
				ID         uint
				Key        string
				Name       string
				RateLimit  int
				IsActive   bool
				ExpiresAt  *time.Time
				LastUsedAt *time.Time
				CreatedAt  time.Time
			}{
				ID:         key.ID,
				Key:        key.Key,
				Name:       key.Name,
				RateLimit:  key.RateLimit,
				IsActive:   key.IsActive,
				ExpiresAt:  key.ExpiresAt,
				LastUsedAt: key.LastUsedAt,
				CreatedAt:  key.CreatedAt,
			})
		}

		printAPIKeyTable(tableData)
	}
}
