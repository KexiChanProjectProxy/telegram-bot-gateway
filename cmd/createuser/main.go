package main

import (
	"context"
	"flag"
	"log"

	"github.com/kexi/telegram-bot-gateway/internal/config"
	"github.com/kexi/telegram-bot-gateway/internal/pkg/jwt"
	"github.com/kexi/telegram-bot-gateway/internal/repository"
	"github.com/kexi/telegram-bot-gateway/internal/service"
)

func main() {
	username := flag.String("username", "admin", "Admin username")
	password := flag.String("password", "", "Admin password (required)")
	email := flag.String("email", "", "Admin email (optional)")
	flag.Parse()

	if *password == "" {
		log.Fatal("Password is required. Use -password flag")
	}

	// Load configuration
	cfg, err := config.Load("configs/config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := repository.NewDatabase(&cfg.Database, cfg.Server.Mode)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Initialize repositories and services
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)

	jwtService := jwt.NewService(
		cfg.Auth.JWT.Secret,
		cfg.Auth.JWT.Issuer,
		cfg.Auth.JWT.AccessTokenTTL,
		cfg.Auth.JWT.RefreshTokenTTL,
		cfg.Auth.JWT.RefreshThreshold,
	)

	authService := service.NewAuthService(userRepo, refreshTokenRepo, jwtService)

	// Create admin user
	ctx := context.Background()
	user, err := authService.CreateUser(ctx, *username, *email, *password)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	log.Printf("✓ Created user: %s (ID: %d)", user.Username, user.ID)
	log.Println("✓ Now assign admin role manually:")
	log.Printf("   INSERT INTO user_roles (user_id, role_id) VALUES (%d, 1);", user.ID)
}
