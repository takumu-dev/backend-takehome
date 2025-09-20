package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"blog-platform/internal/infrastructure/config"
)

// GetCORSConfig returns CORS configuration based on application config
func GetCORSConfig(cfg *config.Config) middleware.CORSConfig {

	corsConfig := middleware.CORSConfig{
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			"X-Requested-With",
			"X-Request-ID",
		},
		ExposeHeaders: []string{
			"X-Request-ID",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}

	// Use configured origins if available, otherwise use environment-based defaults
	if len(cfg.CORS.AllowedOrigins) > 0 {
		corsConfig.AllowOrigins = cfg.CORS.AllowedOrigins
	} else {
		// Set default origins based on environment
		switch cfg.CORS.Environment {
		case "production":
			// Default production origins - should be configured via ALLOWED_ORIGINS
			corsConfig.AllowOrigins = []string{
				"https://yourdomain.com",
				"https://www.yourdomain.com",
			}
		case "staging":
			// Staging: Allow staging domains
			corsConfig.AllowOrigins = []string{
				"https://staging.yourdomain.com",
				"https://preview.yourdomain.com",
				"http://localhost:3000",
				"http://localhost:3001",
			}
		default:
			// Development: Allow localhost and common development ports
			corsConfig.AllowOrigins = []string{
				"http://localhost:3000",
				"http://localhost:3001",
				"http://localhost:8080",
				"http://localhost:8081",
				"http://127.0.0.1:3000",
				"http://127.0.0.1:3001",
				"http://127.0.0.1:8080",
				"http://127.0.0.1:8081",
			}
		}
	}

	return corsConfig
}

// CORS returns CORS middleware with configuration-based setup
func CORS(cfg *config.Config) echo.MiddlewareFunc {
	return middleware.CORSWithConfig(GetCORSConfig(cfg))
}

// CORSWithConfig returns CORS middleware with custom configuration
func CORSWithConfig(config middleware.CORSConfig) echo.MiddlewareFunc {
	return middleware.CORSWithConfig(config)
}
