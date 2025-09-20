package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"blog-platform/internal/application/service"
	"blog-platform/internal/infrastructure/config"
)

// NewLogger creates a logger based on configuration
func NewLogger(cfg *config.Config) service.Logger {
	// Determine log level
	var level slog.Level
	switch strings.ToLower(cfg.Logging.Level) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Determine output destination
	var output io.Writer
	switch strings.ToLower(cfg.Logging.Output) {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		// If it's a file path, try to open it
		if cfg.Logging.Output != "" {
			if file, err := os.OpenFile(cfg.Logging.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
				output = file
			} else {
				// Fallback to stdout if file can't be opened
				output = os.Stdout
			}
		} else {
			output = os.Stdout
		}
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level: level,
		AddSource: level == slog.LevelDebug, // Add source info for debug level
	}

	// Determine handler type based on format
	var handler slog.Handler
	switch strings.ToLower(cfg.Logging.Format) {
	case "text":
		handler = slog.NewTextHandler(output, opts)
	case "json":
		handler = slog.NewJSONHandler(output, opts)
	default:
		handler = slog.NewJSONHandler(output, opts)
	}

	// Create slog logger
	slogger := slog.New(handler)

	// Return wrapped logger
	return NewOperationLogger(slogger)
}

// GetLogLevel returns the current log level as a string
func GetLogLevel(cfg *config.Config) string {
	return cfg.Logging.Level
}

// IsDebugEnabled returns true if debug logging is enabled
func IsDebugEnabled(cfg *config.Config) bool {
	return strings.ToLower(cfg.Logging.Level) == "debug"
}
