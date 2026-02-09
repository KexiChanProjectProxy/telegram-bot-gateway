package commands

import (
	"fmt"
	"strconv"
)

const getUsage = `Get API key details

Usage:
  apikey get <id>

Arguments:
  <id>                      API key ID

Options:
  --help, -h                Show this help message

Examples:
  apikey get 1
`

func GetAPIKey(args []string) {
	if hasFlag(args, "--help", "-h") || len(args) == 0 {
		fmt.Println(getUsage)
		return
	}

	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		fatal("Invalid API key ID: %s", args[0])
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

	// Get API key
	ctx, cancel := getContext()
	defer cancel()

	apiKey, err := apiKeyRepo.GetByID(ctx, uint(id))
	if err != nil {
		fatal("Failed to get API key: %v", err)
	}

	// Display details
	info("API Key Details")
	info("================")
	info("ID:          %d", apiKey.ID)
	info("Key Prefix:  %s...", apiKey.Key)
	info("Name:        %s", apiKey.Name)
	if apiKey.Description != "" {
		info("Description: %s", apiKey.Description)
	}
	info("Rate Limit:  %d requests/hour", apiKey.RateLimit)
	info("Active:      %t", apiKey.IsActive)
	if apiKey.ExpiresAt != nil {
		info("Expires:     %s", apiKey.ExpiresAt.Format("2006-01-02 15:04:05"))
	} else {
		info("Expires:     Never")
	}
	if apiKey.LastUsedAt != nil {
		info("Last Used:   %s", apiKey.LastUsedAt.Format("2006-01-02 15:04:05"))
	} else {
		info("Last Used:   Never")
	}
	info("Created:     %s", apiKey.CreatedAt.Format("2006-01-02 15:04:05"))
	info("")
	info("To see permissions: apikey show-permissions %d", apiKey.ID)
}
