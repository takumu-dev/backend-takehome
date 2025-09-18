package http

import (
	"github.com/labstack/echo/v4"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes(e *echo.Echo) {
	// API v1 group
	v1 := e.Group("/api/v1")
	
	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "ok",
			"service": "blog-platform",
		})
	})
	
	// Auth routes
	auth := v1.Group("/auth")
	_ = auth // TODO: Add auth handlers
	
	// Posts routes
	posts := v1.Group("/posts")
	_ = posts // TODO: Add post handlers
	
	// Documentation route
	e.GET("/docs/*", func(c echo.Context) error {
		// TODO: Add swagger documentation handler
		return c.String(200, "API Documentation - Coming Soon")
	})
}
