package commands

import (
	"fmt"
	"strconv"
)

const deleteUsage = `Delete an API key

Usage:
  apikey delete <id>

Arguments:
  <id>                      API key ID to delete

Options:
  --help, -h                Show this help message

Examples:
  apikey delete 1

Warning: This permanently deletes the API key and all associated permissions.
         This action cannot be undone.
`

func DeleteAPIKey(args []string) {
	if hasFlag(args, "--help", "-h") || len(args) == 0 {
		fmt.Println(deleteUsage)
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

	// Get API key first to show name
	ctx, cancel := getContext()
	defer cancel()

	apiKey, err := apiKeyRepo.GetByID(ctx, uint(id))
	if err != nil {
		fatal("Failed to get API key: %v", err)
	}

	// Delete API key (cascades to all permissions)
	if err := apiKeyRepo.Delete(ctx, uint(id)); err != nil {
		fatal("Failed to delete API key: %v", err)
	}

	success("API key %d (%s) deleted successfully", apiKey.ID, apiKey.Name)
	info("All associated permissions have been removed.")
}
