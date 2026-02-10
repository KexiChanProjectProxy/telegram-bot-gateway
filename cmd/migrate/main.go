package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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
		migrations := []string{
			"migrations/001_initial_schema.sql",
			"migrations/003_bot_webhook_secret.sql",
		}
		for _, migration := range migrations {
			if _, err := os.Stat(migration); os.IsNotExist(err) {
				log.Printf("Skipping non-existent migration: %s", migration)
				continue
			}
			log.Printf("Running migration: %s", migration)
			if err := runMigration(db, migration); err != nil {
				log.Fatalf("Migration failed: %v", err)
			}
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

	// Split SQL statements by semicolon
	statements := splitSQLStatements(string(content))

	// Execute each statement
	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}

		_, err = db.Exec(stmt)
		if err != nil {
			return fmt.Errorf("failed to execute statement %d: %w\nStatement: %s", i+1, err, stmt)
		}
	}

	return nil
}

// splitSQLStatements splits SQL content by semicolons, ignoring semicolons in comments
func splitSQLStatements(content string) []string {
	var statements []string
	var current strings.Builder
	inBlockComment := false

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle line comments
		if strings.HasPrefix(trimmed, "--") {
			continue
		}

		// Handle block comments
		if strings.Contains(line, "/*") {
			inBlockComment = true
		}
		if strings.Contains(line, "*/") {
			inBlockComment = false
			continue
		}
		if inBlockComment {
			continue
		}

		// Add line to current statement
		current.WriteString(line)
		current.WriteString("\n")

		// Check if line ends with semicolon (end of statement)
		if strings.HasSuffix(trimmed, ";") {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
		}
	}

	// Add any remaining statement
	if current.Len() > 0 {
		stmt := strings.TrimSpace(current.String())
		if stmt != "" {
			statements = append(statements, stmt)
		}
	}

	return statements
}
