// @title Blog Platform API
// @version 1.0
// @description A RESTful API for a blog platform with user authentication, posts, and comments management.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@blog-platform.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"blog-platform/internal/infrastructure/config"
	"blog-platform/internal/infrastructure/database"
	http "blog-platform/internal/infrastructure/http"
	"blog-platform/internal/infrastructure/logging"
	"blog-platform/internal/infrastructure/repository"

	"blog-platform/internal/application/service"
	infraauth "blog-platform/internal/infrastructure/auth"
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

	// Basic middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

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

	// Initialize logger with configuration
	logger := logging.NewLogger(cfg)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.DB)
	postRepo := repository.NewPostRepository(db.DB)
	commentRepo := repository.NewCommentRepository(db.DB)

	// Initialize JWT service
	jwtService := infraauth.NewJWTService(cfg.JWT.Secret)

	// Initialize domain services
	userService := service.NewUserService(userRepo, logger)
	postService := service.NewPostService(postRepo, logger)
	commentService := service.NewCommentService(commentRepo, logger)
	authService := service.NewAuthService(userService, jwtService, logger)

	// Setup routes
	http.SetupRoutes(e, cfg, userService, authService, postService, commentService, logger)

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
