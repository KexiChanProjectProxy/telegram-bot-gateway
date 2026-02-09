package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	"github.com/kexi/telegram-bot-gateway/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/migrate/main.go [up|down]")
		os.Exit(1)
	}

	command := os.Args[1]

	// Load config
	cfg, err := config.Load("configs/config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := sql.Open(cfg.Database.Driver, cfg.Database.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	switch command {
	case "up":
		if err := runMigration(db, "migrations/001_initial_schema.sql"); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Migration completed successfully")
	case "down":
		if err := runMigration(db, "migrations/001_initial_schema_down.sql"); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		log.Println("Rollback completed successfully")
	default:
		log.Fatalf("Unknown command: %s. Use 'up' or 'down'", command)
	}
}

func runMigration(db *sql.DB, migrationFile string) error {
	absPath, err := filepath.Abs(migrationFile)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file %s: %w", migrationFile, err)
	}

	_, err = db.Exec(string(content))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}
