package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"blog-platform/internal/domain/auth"
	"blog-platform/internal/domain/user"
	"blog-platform/internal/infrastructure/http/handlers"
	"blog-platform/internal/infrastructure/http/middleware"
)

// MockUserService implements the user.Service interface for testing
type MockUserService struct {
	users map[string]*user.User
	nextID int
}

func NewMockUserService() *MockUserService {
	return &MockUserService{
		users:  make(map[string]*user.User),
		nextID: 1,
	}
}

func (m *MockUserService) Register(ctx context.Context, name, email, password string) (*user.User, error) {
	if _, exists := m.users[email]; exists {
		return nil, user.ErrUserExists
	}
	
	u, err := user.NewUser(name, email, password)
	if err != nil {
		return nil, err
	}
	
	u.ID = m.nextID
	m.nextID++
	m.users[email] = u
	return u, nil
}

func (m *MockUserService) Login(ctx context.Context, email, password string) (*user.User, error) {
	u, exists := m.users[email]
	if !exists {
		return nil, user.ErrInvalidCredentials
	}
	
	if !u.ValidatePassword(password) {
		return nil, user.ErrInvalidCredentials
	}
	
	return u, nil
}

func (m *MockUserService) GetByID(ctx context.Context, id int) (*user.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, user.ErrUserNotFound
}

func (m *MockUserService) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	u, exists := m.users[email]
	if !exists {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

func (m *MockUserService) UpdateProfile(ctx context.Context, id int, name, email string) (*user.User, error) {
	return nil, nil // Not needed for auth tests
}

func (m *MockUserService) UpdatePassword(ctx context.Context, id int, currentPassword, newPassword string) error {
	return nil // Not needed for auth tests
}

func (m *MockUserService) Delete(ctx context.Context, id int) error {
	return nil // Not needed for auth tests
}

func (m *MockUserService) List(ctx context.Context, limit, offset int) ([]*user.User, error) {
	return nil, nil // Not needed for auth tests
}

// MockAuthService implements the auth.AuthService interface for testing
type MockAuthService struct{
	userService *MockUserService
}

func NewMockAuthService(userService *MockUserService) *MockAuthService {
	return &MockAuthService{
		userService: userService,
	}
}

func (m *MockAuthService) GenerateToken(ctx context.Context, user *user.User) (string, error) {
	return "mock-jwt-token", nil
}

func (m *MockAuthService) ValidateToken(ctx context.Context, token string) (*auth.TokenClaims, error) {
	if token == "mock-jwt-token" {
		return &auth.TokenClaims{
			UserID: 1,
			Email:  "test@example.com",
		}, nil
	}
	return nil, auth.ErrInvalidToken
}

func (m *MockAuthService) Login(ctx context.Context, email, password string) (*user.User, string, error) {
	u, err := m.userService.Login(ctx, email, password)
	if err != nil {
		return nil, "", err
	}
	token, _ := m.GenerateToken(ctx, u)
	return u, token, nil
}

func (m *MockAuthService) Register(ctx context.Context, name, email, password string) (*user.User, string, error) {
	u, err := m.userService.Register(ctx, name, email, password)
	if err != nil {
		return nil, "", err
	}
	token, _ := m.GenerateToken(ctx, u)
	return u, token, nil
}

func (m *MockAuthService) RefreshToken(ctx context.Context, token string) (string, error) {
	return "mock-refreshed-token", nil
}

// MockLogger implements the service.Logger interface for testing
type MockLogger struct{}

func (m *MockLogger) Info(ctx context.Context, msg string, args ...any)  {}
func (m *MockLogger) Error(ctx context.Context, msg string, args ...any) {}
func (m *MockLogger) Warn(ctx context.Context, msg string, args ...any)  {}
func (m *MockLogger) Debug(ctx context.Context, msg string, args ...any) {}

func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func setupTestServer() (*echo.Echo, *handlers.AuthHandler) {
	e := echo.New()
	e.Validator = middleware.NewValidator()
	
	userService := NewMockUserService()
	authService := NewMockAuthService(userService)
	logger := NewMockLogger()
	
	authHandler := handlers.NewAuthHandler(userService, authService, logger)
	
	return e, authHandler
}

func TestAuthHandler_Register_Success(t *testing.T) {
	e, authHandler := setupTestServer()
	
	registerReq := handlers.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	
	reqBody, err := json.Marshal(registerReq)
	require.NoError(t, err)
	
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	err = authHandler.Register(c)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusCreated, rec.Code)
	
	var response handlers.AuthResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "John Doe", response.User.Name)
	assert.Equal(t, "john@example.com", response.User.Email)
	assert.Equal(t, "mock-jwt-token", response.Token)
	assert.NotZero(t, response.User.ID)
}

func TestAuthHandler_Register_ValidationError(t *testing.T) {
	e, authHandler := setupTestServer()
	
	registerReq := handlers.RegisterRequest{
		Name:     "", // Invalid: empty name
		Email:    "invalid-email", // Invalid: not an email
		Password: "123", // Invalid: too short
	}
	
	reqBody, err := json.Marshal(registerReq)
	require.NoError(t, err)
	
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	err = authHandler.Register(c)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	
	var response handlers.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "validation_error", response.Error)
}

func TestAuthHandler_Register_UserExists(t *testing.T) {
	e, authHandler := setupTestServer()
	
	// First registration
	registerReq := handlers.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	
	reqBody, err := json.Marshal(registerReq)
	require.NoError(t, err)
	
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	err = authHandler.Register(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	
	// Second registration with same email
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(reqBody))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	
	err = authHandler.Register(c2)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusConflict, rec2.Code)
	
	var response handlers.ErrorResponse
	err = json.Unmarshal(rec2.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "conflict", response.Error)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	e, authHandler := setupTestServer()
	
	// First register a user
	registerReq := handlers.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	
	regReqBody, err := json.Marshal(registerReq)
	require.NoError(t, err)
	
	regReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regReqBody))
	regReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	regRec := httptest.NewRecorder()
	regC := e.NewContext(regReq, regRec)
	
	err = authHandler.Register(regC)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, regRec.Code)
	
	// Now test login
	loginReq := handlers.LoginRequest{
		Email:    "john@example.com",
		Password: "password123",
	}
	
	reqBody, err := json.Marshal(loginReq)
	require.NoError(t, err)
	
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	err = authHandler.Login(c)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusOK, rec.Code)
	
	var response handlers.AuthResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "John Doe", response.User.Name)
	assert.Equal(t, "john@example.com", response.User.Email)
	assert.Equal(t, "mock-jwt-token", response.Token)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	e, authHandler := setupTestServer()
	
	loginReq := handlers.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "wrongpassword",
	}
	
	reqBody, err := json.Marshal(loginReq)
	require.NoError(t, err)
	
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	err = authHandler.Login(c)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	
	var response handlers.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "invalid_credentials", response.Error)
}

func TestAuthHandler_Login_ValidationError(t *testing.T) {
	e, authHandler := setupTestServer()
	
	loginReq := handlers.LoginRequest{
		Email:    "invalid-email", // Invalid email format
		Password: "", // Empty password
	}
	
	reqBody, err := json.Marshal(loginReq)
	require.NoError(t, err)
	
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	err = authHandler.Login(c)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	
	var response handlers.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "validation_error", response.Error)
}
