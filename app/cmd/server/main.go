package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"blog-platform/internal/infrastructure/config"
	"blog-platform/internal/infrastructure/database"
	"blog-platform/internal/infrastructure/http"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Health check endpoint with database status
	e.GET("/health", func(c echo.Context) error {
		// Test database connection
		if err := db.Ping(); err != nil {
			return c.JSON(503, map[string]string{
				"status":   "error",
				"service":  "blog-platform",
				"database": "disconnected",
				"error":    err.Error(),
			})
		}

		return c.JSON(200, map[string]string{
			"status":   "ok",
			"service":  "blog-platform",
			"database": "connected",
		})
	})

	// Setup routes
	http.SetupRoutes(e)

	// Start server
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
