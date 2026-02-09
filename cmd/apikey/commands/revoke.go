package commands

import (
	"fmt"
	"strconv"
)

const revokeUsage = `Revoke (deactivate) an API key

Usage:
  apikey revoke <id>

Arguments:
  <id>                      API key ID to revoke

Options:
  --help, -h                Show this help message

Examples:
  apikey revoke 1

Note: Revoked keys can be reactivated by setting is_active=true in the database
`

func RevokeAPIKey(args []string) {
	if hasFlag(args, "--help", "-h") || len(args) == 0 {
		fmt.Println(revokeUsage)
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

	// Get and revoke API key
	ctx, cancel := getContext()
	defer cancel()

	apiKey, err := apiKeyRepo.GetByID(ctx, uint(id))
	if err != nil {
		fatal("Failed to get API key: %v", err)
	}

	apiKey.IsActive = false
	if err := apiKeyRepo.Update(ctx, apiKey); err != nil {
		fatal("Failed to revoke API key: %v", err)
	}

	success("API key %d (%s) revoked successfully", apiKey.ID, apiKey.Name)
}
