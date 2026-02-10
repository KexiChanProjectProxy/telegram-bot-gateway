package utils

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger wraps zerolog.Logger with additional functionality
type Logger struct {
	zerolog.Logger
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level       string // "debug", "info", "warn", "error"
	PrettyPrint bool   // Enable pretty console output
	Output      io.Writer // Output writer (defaults to os.Stdout)
}

// InitLogger initializes the global logger with the provided configuration
func InitLogger(cfg LogConfig) *Logger {
	// Set log level
	level := parseLogLevel(cfg.Level)
	zerolog.SetGlobalLevel(level)

	// Configure output
	var output io.Writer
	if cfg.Output != nil {
		output = cfg.Output
	} else {
		output = os.Stdout
	}

	// Enable pretty printing for development
	if cfg.PrettyPrint {
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
	}

	// Create logger
	logger := zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()

	// Set as global logger
	log.Logger = logger

	return &Logger{Logger: logger}
}

// NewLogger creates a new logger instance with the provided configuration
func NewLogger(cfg LogConfig) *Logger {
	// Set log level
	level := parseLogLevel(cfg.Level)

	// Configure output
	var output io.Writer
	if cfg.Output != nil {
		output = cfg.Output
	} else {
		output = os.Stdout
	}

	// Enable pretty printing for development
	if cfg.PrettyPrint {
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
	}

	// Create logger with level
	logger := zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()

	return &Logger{Logger: logger}
}

// parseLogLevel converts string log level to zerolog.Level
func parseLogLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

// WithComponent returns a logger with a component field
func (l *Logger) WithComponent(component string) *Logger {
	newLogger := l.Logger.With().Str("component", component).Logger()
	return &Logger{Logger: newLogger}
}

// WithField returns a logger with an additional field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	newLogger := l.Logger.With().Interface(key, value).Logger()
	return &Logger{Logger: newLogger}
}

// WithFields returns a logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newLogger := l.Logger.With().Fields(fields).Logger()
	return &Logger{Logger: newLogger}
}

// WithError returns a logger with an error field
func (l *Logger) WithError(err error) *Logger {
	newLogger := l.Logger.With().Err(err).Logger()
	return &Logger{Logger: newLogger}
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	return &Logger{Logger: log.Logger}
}

// Debug logs a debug message
func Debug(msg string) {
	log.Debug().Msg(msg)
}

// Info logs an info message
func Info(msg string) {
	log.Info().Msg(msg)
}

// Warn logs a warning message
func Warn(msg string) {
	log.Warn().Msg(msg)
}

// Error logs an error message
func Error(msg string) {
	log.Error().Msg(msg)
}

// Fatal logs a fatal message and exits
func Fatal(msg string) {
	log.Fatal().Msg(msg)
}

// Debugf logs a debug message with formatting
func Debugf(format string, args ...interface{}) {
	log.Debug().Msgf(format, args...)
}

// Infof logs an info message with formatting
func Infof(format string, args ...interface{}) {
	log.Info().Msgf(format, args...)
}

// Warnf logs a warning message with formatting
func Warnf(format string, args ...interface{}) {
	log.Warn().Msgf(format, args...)
}

// Errorf logs an error message with formatting
func Errorf(format string, args ...interface{}) {
	log.Error().Msgf(format, args...)
}

// Fatalf logs a fatal message with formatting and exits
func Fatalf(format string, args ...interface{}) {
	log.Fatal().Msgf(format, args...)
}
