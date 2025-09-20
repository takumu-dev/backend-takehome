package middleware

import (
	"net/http"
	"strings"

	"blog-platform/internal/domain/auth"

	"github.com/labstack/echo/v4"
)

// AuthMiddleware handles JWT authentication for HTTP requests
type AuthMiddleware struct {
	tokenService auth.TokenService
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(tokenService auth.TokenService) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
	}
}

// RequireAuth returns middleware that requires valid authentication
func (m *AuthMiddleware) RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, err := m.extractToken(c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "missing or invalid authorization header",
				})
			}

			claims, err := m.tokenService.ValidateToken(token)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "invalid or expired token",
				})
			}

			// Set claims in context for use by handlers
			c.Set("user", claims)
			return next(c)
		}
	}
}

// OptionalAuth returns middleware that optionally validates authentication
// If a valid token is provided, the user is set in context
// If no token or invalid token is provided, the request continues without user context
func (m *AuthMiddleware) OptionalAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, err := m.extractToken(c)
			if err != nil {
				// No token provided, continue without authentication
				return next(c)
			}

			claims, err := m.tokenService.ValidateToken(token)
			if err != nil {
				// Invalid token, continue without authentication
				return next(c)
			}

			// Set claims in context for use by handlers
			c.Set("user", claims)
			return next(c)
		}
	}
}

// extractToken extracts the JWT token from the Authorization header
func (m *AuthMiddleware) extractToken(c echo.Context) (string, error) {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return "", auth.ErrMissingToken
	}

	// Check if header starts with "Bearer "
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", auth.ErrInvalidToken
	}

	// Extract token (remove "Bearer " prefix)
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", auth.ErrInvalidToken
	}

	return token, nil
}
