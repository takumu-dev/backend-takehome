package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"blog-platform/internal/application/service"
	"blog-platform/internal/domain/auth"
)

// AuthMiddleware handles authentication for protected routes
type AuthMiddleware struct {
	authService auth.AuthService
	logger      service.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authService auth.AuthService, logger service.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
	}
}

// RequireAuth is a middleware that requires authentication
func (m *AuthMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		
		// Get Authorization header
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			m.logger.Warn(ctx, "missing authorization header")
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error":   "Unauthorized",
				"message": "Authorization header required",
			})
		}

		// Check if it's a Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.logger.Warn(ctx, "invalid authorization header format")
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error":   "Unauthorized",
				"message": "Invalid authorization header format",
			})
		}

		token := parts[1]
		if token == "" {
			m.logger.Warn(ctx, "empty bearer token")
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error":   "Unauthorized",
				"message": "Bearer token required",
			})
		}

		// Validate token
		claims, err := m.authService.ValidateToken(ctx, token)
		if err != nil {
			m.logger.Warn(ctx, "token validation failed", "error", err.Error())
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error":   "Unauthorized",
				"message": "Invalid or expired token",
			})
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)

		m.logger.Debug(ctx, "user authenticated", "user_id", claims.UserID, "email", claims.Email)
		return next(c)
	}
}
