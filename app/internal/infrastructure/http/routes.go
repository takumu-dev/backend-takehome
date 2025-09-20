package http

import (
	"github.com/labstack/echo/v4"

	"blog-platform/internal/application/service"
	"blog-platform/internal/domain/auth"
	"blog-platform/internal/domain/post"
	"blog-platform/internal/domain/user"
	"blog-platform/internal/infrastructure/http/handlers"
	"blog-platform/internal/infrastructure/http/middleware"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes(e *echo.Echo, userService user.Service, authService auth.AuthService, postService post.Service, logger service.Logger) {
	// Set up validator
	e.Validator = middleware.NewValidator()
	
	// API v1 group
	v1 := e.Group("/api/v1")
	
	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "ok",
			"service": "blog-platform",
		})
	})
	
	// Auth handlers
	authHandler := handlers.NewAuthHandler(userService, authService, logger)
	
	// Post handlers
	postHandler := handlers.NewPostHandler(postService, logger)
	
	// Auth middleware for protected routes
	authMiddleware := middleware.NewAuthMiddleware(authService, logger)
	
	// Auth routes
	auth := v1.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	
	// Posts routes
	posts := v1.Group("/posts")
	posts.GET("", postHandler.ListPosts)                                    // GET /api/v1/posts
	posts.GET("/:id", postHandler.GetPost)                                  // GET /api/v1/posts/{id}
	posts.POST("", postHandler.CreatePost, authMiddleware.RequireAuth)      // POST /api/v1/posts (protected)
	posts.PUT("/:id", postHandler.UpdatePost, authMiddleware.RequireAuth)   // PUT /api/v1/posts/{id} (protected)
	posts.DELETE("/:id", postHandler.DeletePost, authMiddleware.RequireAuth) // DELETE /api/v1/posts/{id} (protected)
	
	// Documentation route
	e.GET("/docs/*", func(c echo.Context) error {
		// TODO: Add swagger documentation handler
		return c.String(200, "API Documentation - Coming Soon")
	})
}
