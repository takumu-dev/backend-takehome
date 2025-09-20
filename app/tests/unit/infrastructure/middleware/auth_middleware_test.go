package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"blog-platform/internal/domain/auth"
	"blog-platform/internal/infrastructure/middleware"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// MockTokenService implements the auth.TokenService interface for testing
type MockTokenService struct {
	validTokens map[string]*auth.TokenClaims
}

// NewMockTokenService creates a new mock token service
func NewMockTokenService() *MockTokenService {
	return &MockTokenService{
		validTokens: make(map[string]*auth.TokenClaims),
	}
}

// AddValidToken adds a valid token for testing
func (m *MockTokenService) AddValidToken(token string, claims *auth.TokenClaims) {
	m.validTokens[token] = claims
}

// GenerateToken generates a JWT token for the given user ID and email
func (m *MockTokenService) GenerateToken(userID int, email string, duration time.Duration) (string, error) {
	if userID <= 0 {
		return "", auth.ErrInvalidToken
	}
	if email == "" {
		return "", auth.ErrInvalidToken
	}
	
	token := "valid_token_" + email
	claims := &auth.TokenClaims{
		UserID:    userID,
		Email:     email,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(duration).Unix(),
	}
	m.validTokens[token] = claims
	return token, nil
}

// ValidateToken validates a JWT token and returns the token claims
func (m *MockTokenService) ValidateToken(token string) (*auth.TokenClaims, error) {
	if token == "" {
		return nil, auth.ErrInvalidToken
	}
	
	claims, exists := m.validTokens[token]
	if !exists {
		return nil, auth.ErrInvalidToken
	}
	
	return claims, nil
}

// RefreshToken generates a new token from an existing valid token
func (m *MockTokenService) RefreshToken(token string) (string, error) {
	claims, err := m.ValidateToken(token)
	if err != nil {
		return "", err
	}
	
	newToken := "refreshed_token_" + claims.Email
	newClaims := &auth.TokenClaims{
		UserID:    claims.UserID,
		Email:     claims.Email,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}
	m.validTokens[newToken] = newClaims
	return newToken, nil
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	// Setup
	tokenService := NewMockTokenService()
	testClaims := &auth.TokenClaims{
		UserID:    1,
		Email:     "test@example.com",
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}
	
	token := "valid_token_test@example.com"
	tokenService.AddValidToken(token, testClaims)
	
	authMiddleware := middleware.NewAuthMiddleware(tokenService)
	
	// Create Echo instance and test handler
	e := echo.New()
	handler := func(c echo.Context) error {
		// Check if user is set in context
		user := c.Get("user")
		assert.NotNil(t, user)
		assert.Equal(t, testClaims, user)
		return c.JSON(http.StatusOK, map[string]string{"message": "success"})
	}
	
	// Create request with valid Authorization header
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	// Apply middleware
	middlewareFunc := authMiddleware.RequireAuth()
	h := middlewareFunc(handler)
	
	// Execute
	err := h(c)
	
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	// Setup
	tokenService := NewMockTokenService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService)
	
	// Create Echo instance and test handler
	e := echo.New()
	handler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "success"})
	}
	
	// Create request without Authorization header
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	// Apply middleware
	middlewareFunc := authMiddleware.RequireAuth()
	h := middlewareFunc(handler)
	
	// Execute
	err := h(c)
	
	// Assert
	assert.NoError(t, err) // Middleware should handle the error internally
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_InvalidTokenFormat(t *testing.T) {
	// Setup
	tokenService := NewMockTokenService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService)
	
	// Create Echo instance and test handler
	e := echo.New()
	handler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "success"})
	}
	
	// Create request with invalid Authorization header format
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	// Apply middleware
	middlewareFunc := authMiddleware.RequireAuth()
	h := middlewareFunc(handler)
	
	// Execute
	err := h(c)
	
	// Assert
	assert.NoError(t, err) // Middleware should handle the error internally
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	// Setup
	tokenService := NewMockTokenService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService)
	
	// Create Echo instance and test handler
	e := echo.New()
	handler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "success"})
	}
	
	// Create request with invalid token
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	// Apply middleware
	middlewareFunc := authMiddleware.RequireAuth()
	h := middlewareFunc(handler)
	
	// Execute
	err := h(c)
	
	// Assert
	assert.NoError(t, err) // Middleware should handle the error internally
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_EmptyToken(t *testing.T) {
	// Setup
	tokenService := NewMockTokenService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService)
	
	// Create Echo instance and test handler
	e := echo.New()
	handler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "success"})
	}
	
	// Create request with empty token
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer ")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	// Apply middleware
	middlewareFunc := authMiddleware.RequireAuth()
	h := middlewareFunc(handler)
	
	// Execute
	err := h(c)
	
	// Assert
	assert.NoError(t, err) // Middleware should handle the error internally
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_OptionalAuth_ValidToken(t *testing.T) {
	// Setup
	tokenService := NewMockTokenService()
	testClaims := &auth.TokenClaims{
		UserID:    1,
		Email:     "test@example.com",
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}
	
	token := "valid_token_test@example.com"
	tokenService.AddValidToken(token, testClaims)
	
	authMiddleware := middleware.NewAuthMiddleware(tokenService)
	
	// Create Echo instance and test handler
	e := echo.New()
	handler := func(c echo.Context) error {
		// Check if user is set in context
		user := c.Get("user")
		assert.NotNil(t, user)
		assert.Equal(t, testClaims, user)
		return c.JSON(http.StatusOK, map[string]string{"message": "success"})
	}
	
	// Create request with valid Authorization header
	req := httptest.NewRequest(http.MethodGet, "/optional", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	// Apply optional middleware
	middlewareFunc := authMiddleware.OptionalAuth()
	h := middlewareFunc(handler)
	
	// Execute
	err := h(c)
	
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthMiddleware_OptionalAuth_NoToken(t *testing.T) {
	// Setup
	tokenService := NewMockTokenService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService)
	
	// Create Echo instance and test handler
	e := echo.New()
	handler := func(c echo.Context) error {
		// Check that no user is set in context
		user := c.Get("user")
		assert.Nil(t, user)
		return c.JSON(http.StatusOK, map[string]string{"message": "success"})
	}
	
	// Create request without Authorization header
	req := httptest.NewRequest(http.MethodGet, "/optional", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	// Apply optional middleware
	middlewareFunc := authMiddleware.OptionalAuth()
	h := middlewareFunc(handler)
	
	// Execute
	err := h(c)
	
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthMiddleware_Interface(t *testing.T) {
	tokenService := NewMockTokenService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService)
	
	// Verify that AuthMiddleware implements the expected interface
	assert.NotNil(t, authMiddleware)
	assert.NotNil(t, authMiddleware.RequireAuth())
	assert.NotNil(t, authMiddleware.OptionalAuth())
}
