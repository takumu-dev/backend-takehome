package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"blog-platform/internal/infrastructure/config"
)

// CompressionConfig holds compression configuration
type CompressionConfig struct {
	// Enabled controls whether compression is enabled
	Enabled bool
	// Level defines the compression level (1-9, where 9 is best compression)
	Level int
	// MinLength defines the minimum response size to compress (in bytes)
	MinLength int
	// Skipper defines a function to skip compression for certain requests
	Skipper middleware.Skipper
}

// DefaultCompressionConfig returns default compression configuration
func DefaultCompressionConfig() CompressionConfig {
	return CompressionConfig{
		Enabled:   true,
		Level:     6, // Good balance between compression ratio and speed
		MinLength: 1024, // Only compress responses larger than 1KB
		Skipper:   middleware.DefaultSkipper,
	}
}

// GetCompressionConfig returns compression configuration based on application config
func GetCompressionConfig(cfg *config.Config) middleware.GzipConfig {
	return middleware.GzipConfig{
		Level: cfg.Compression.Level,
		Skipper: func(c echo.Context) bool {
			// Skip compression if disabled
			if !cfg.Compression.Enabled {
				return true
			}
			
			// Skip compression for certain content types that are already compressed
			contentType := c.Response().Header().Get(echo.HeaderContentType)
			switch contentType {
			case "image/jpeg", "image/png", "image/gif", "image/webp",
				 "video/mp4", "video/mpeg", "video/quicktime",
				 "application/zip", "application/gzip", "application/x-gzip":
				return true
			}
			
			// Skip compression for small responses (will be checked after response is written)
			return false
		},
	}
}

// Compression returns compression middleware with configuration-based setup
func Compression(cfg *config.Config) echo.MiddlewareFunc {
	if !cfg.Compression.Enabled {
		// Return a no-op middleware if compression is disabled
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}
	
	return middleware.GzipWithConfig(GetCompressionConfig(cfg))
}

// CompressionWithConfig returns compression middleware with custom configuration
func CompressionWithConfig(config CompressionConfig) echo.MiddlewareFunc {
	if !config.Enabled {
		// Return a no-op middleware if compression is disabled
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}
	
	gzipConfig := middleware.GzipConfig{
		Level:   config.Level,
		Skipper: config.Skipper,
	}
	
	return middleware.GzipWithConfig(gzipConfig)
}

// CompressionDefault returns compression middleware with default configuration
func CompressionDefault() echo.MiddlewareFunc {
	return CompressionWithConfig(DefaultCompressionConfig())
}
