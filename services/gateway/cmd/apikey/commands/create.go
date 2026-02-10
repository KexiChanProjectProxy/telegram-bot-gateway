package commands

import (
	"fmt"
	"time"

	"github.com/kexi/telegram-bot-gateway/internal/domain"
)

const createUsage = `Create a new API key

Usage:
  apikey create --name <name> [options]

Required:
  --name <name>             Name for the API key

Options:
  --description <desc>      Description of the API key
  --rate-limit <n>          Requests per hour (default: 1000)
  --expires <duration>      Expiration time (e.g., 1y, 30d, 24h)
  --help, -h                Show this help message

Examples:
  apikey create --name "Production Service"
  apikey create --name "Dev API" --rate-limit 5000 --expires 1y
  apikey create --name "Test" --description "Testing purposes" --expires 30d
`

func CreateAPIKey(args []string) {
	if hasFlag(args, "--help", "-h") {
		fmt.Println(createUsage)
		return
	}

	// Parse arguments
	name := getFlagValue(args, "--name", "-n")
	if name == "" {
		fatal("--name is required\n\n%s", createUsage)
	}

	description := getFlagValue(args, "--description", "-d")
	rateLimitStr := getFlagValue(args, "--rate-limit", "-r")
	expiresStr := getFlagValue(args, "--expires", "-e")

	rateLimit := 1000
	if rateLimitStr != "" {
		_, err := fmt.Sscanf(rateLimitStr, "%d", &rateLimit)
		if err != nil {
			fatal("Invalid rate limit: %s", rateLimitStr)
		}
	}

	var expiresAt *time.Time
	if expiresStr != "" {
		duration, err := parseDuration(expiresStr)
		if err != nil {
			fatal("Invalid expiration duration: %v", err)
		}
		expires := time.Now().Add(duration)
		expiresAt = &expires
	}

	// Initialize database and services
	db, err := initDB()
	if err != nil {
		fatal("Failed to initialize database: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	apiKeyRepo, _, _, _, _, _ := initRepositories(db)
	apiKeyService := initAPIKeyService()

	// Generate API key
	plainKey, hashedKey, err := apiKeyService.Generate()
	if err != nil {
		fatal("Failed to generate API key: %v", err)
	}

	// Create API key record
	ctx, cancel := getContext()
	defer cancel()

	apiKey := &domain.APIKey{
		Key:         plainKey, // Store full key for lookup
		HashedKey:   hashedKey,
		Name:        name,
		Description: description,
		RateLimit:   rateLimit,
		IsActive:    true,
		ExpiresAt:   expiresAt,
	}

	if err := apiKeyRepo.Create(ctx, apiKey); err != nil {
		fatal("Failed to create API key: %v", err)
	}

	success("API key created successfully!")
	info("")
	info("API Key ID: %d", apiKey.ID)
	info("API Key:    %s", plainKey)
	info("")
	info("⚠️  IMPORTANT: Save this key now! It cannot be retrieved later.")
	info("")
	info("Details:")
	info("  Name:        %s", name)
	if description != "" {
		info("  Description: %s", description)
	}
	info("  Rate Limit:  %d requests/hour", rateLimit)
	if expiresAt != nil {
		info("  Expires:     %s", expiresAt.Format("2006-01-02 15:04:05"))
	} else {
		info("  Expires:     Never")
	}
	info("")
	info("Next steps:")
	info("  1. Grant chat permissions:  apikey grant-chat %d <chat-id> --read --send", apiKey.ID)
	info("  2. View permissions:        apikey show-permissions %d", apiKey.ID)
}
