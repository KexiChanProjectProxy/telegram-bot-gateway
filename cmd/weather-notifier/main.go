package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/weather-notice-bot/internal/app"
	"github.com/user/weather-notice-bot/internal/config"
)

const (
	// Exit codes
	ExitSuccess         = 0
	ExitConfigError     = 1
	ExitAppInitError    = 2
	ExitAppStartError   = 3
	ExitInterrupted     = 130
)

var (
	// Version information (set by build flags)
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "", "path to configuration file (default: config.yaml in current directory)")
	version := flag.Bool("version", false, "print version information and exit")
	flag.Parse()

	// Print version and exit if requested
	if *version {
		printVersion()
		os.Exit(ExitSuccess)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Initialize application
	application, err := app.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize application: %v\n", err)
		os.Exit(ExitAppInitError)
	}

	// Start application
	if err := application.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start application: %v\n", err)
		os.Exit(ExitAppStartError)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Wait for shutdown signal
	logger := application.GetLogger()
	sig := <-sigChan
	logger.Info().
		Str("signal", sig.String()).
		Msg("received shutdown signal")

	// Gracefully stop the application
	application.Stop()

	logger.Info().Msg("shutdown complete")
	os.Exit(ExitSuccess)
}

// printVersion prints version information
func printVersion() {
	fmt.Printf("Weather Notification Bot\n")
	fmt.Printf("Version:    %s\n", Version)
	fmt.Printf("Build Time: %s\n", BuildTime)
	fmt.Printf("Git Commit: %s\n", GitCommit)
}
