package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"blog-platform/internal/application/service"
	"blog-platform/internal/domain/auth"
	"blog-platform/internal/domain/user"
)

// MockUserService implements user.Service for testing
type MockUserService struct {
	users  map[string]*user.User
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
	
	u := &user.User{
		ID:           m.nextID,
		Name:         name,
		Email:        email,
		PasswordHash: "hashed_" + password, // Simple mock hashing
	}
	m.nextID++
	m.users[email] = u
	return u, nil
}

func (m *MockUserService) Login(ctx context.Context, email, password string) (*user.User, error) {
	u, exists := m.users[email]
	if !exists {
		return nil, user.ErrInvalidCredentials
	}
	
	// Simple mock password verification
	if u.PasswordHash != "hashed_"+password {
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
	if u, exists := m.users[email]; exists {
		return u, nil
	}
	return nil, user.ErrUserNotFound
}

func (m *MockUserService) UpdateProfile(ctx context.Context, id int, name, email string) (*user.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			u.Name = name
			u.Email = email
			return u, nil
		}
	}
	return nil, user.ErrUserNotFound
}

func (m *MockUserService) UpdatePassword(ctx context.Context, id int, currentPassword, newPassword string) error {
	for _, u := range m.users {
		if u.ID == id {
			if u.PasswordHash != "hashed_"+currentPassword {
				return user.ErrInvalidCredentials
			}
			u.PasswordHash = "hashed_" + newPassword
			return nil
		}
	}
	return user.ErrUserNotFound
}

func (m *MockUserService) Delete(ctx context.Context, id int) error {
	for email, u := range m.users {
		if u.ID == id {
			delete(m.users, email)
			return nil
		}
	}
	return user.ErrUserNotFound
}

func (m *MockUserService) List(ctx context.Context, limit, offset int) ([]*user.User, error) {
	var users []*user.User
	count := 0
	skipped := 0
	
	for _, u := range m.users {
		if skipped < offset {
			skipped++
			continue
		}
		if count >= limit {
			break
		}
		users = append(users, u)
		count++
	}
	
	return users, nil
}

// MockTokenService implements auth.TokenService for testing
type MockTokenService struct {
	tokens      map[string]*auth.TokenClaims
	shouldError bool
}

func NewMockTokenService() *MockTokenService {
	return &MockTokenService{
		tokens:      make(map[string]*auth.TokenClaims),
		shouldError: false,
	}
}

func (m *MockTokenService) SetError(shouldError bool) {
	m.shouldError = shouldError
}

func (m *MockTokenService) GenerateToken(userID int, email string, duration time.Duration) (string, error) {
	if m.shouldError {
		return "", errors.New("token generation failed")
	}
	token := "mock_token_" + email
	claims := &auth.TokenClaims{
		UserID: userID,
		Email:  email,
	}
	m.tokens[token] = claims
	return token, nil
}

func (m *MockTokenService) ValidateToken(token string) (*auth.TokenClaims, error) {
	if m.shouldError {
		return nil, auth.ErrInvalidToken
	}
	if claims, exists := m.tokens[token]; exists {
		return claims, nil
	}
	return nil, auth.ErrInvalidToken
}

func (m *MockTokenService) RefreshToken(token string) (string, error) {
	if m.shouldError {
		return "", auth.ErrInvalidToken
	}
	if claims, exists := m.tokens[token]; exists {
		newToken := "refreshed_" + token
		m.tokens[newToken] = claims
		delete(m.tokens, token)
		return newToken, nil
	}
	return "", auth.ErrInvalidToken
}

// MockLogger implements service.Logger for testing
type MockLogger struct {
	logs []LogEntry
}

type LogEntry struct {
	Level   string
	Message string
	Args    []any
}

func NewMockLogger() *MockLogger {
	return &MockLogger{
		logs: make([]LogEntry, 0),
	}
}

func (m *MockLogger) Info(ctx context.Context, msg string, args ...any) {
	m.logs = append(m.logs, LogEntry{Level: "info", Message: msg, Args: args})
}

func (m *MockLogger) Error(ctx context.Context, msg string, args ...any) {
	m.logs = append(m.logs, LogEntry{Level: "error", Message: msg, Args: args})
}

func (m *MockLogger) Warn(ctx context.Context, msg string, args ...any) {
	m.logs = append(m.logs, LogEntry{Level: "warn", Message: msg, Args: args})
}

func (m *MockLogger) Debug(ctx context.Context, msg string, args ...any) {
	m.logs = append(m.logs, LogEntry{Level: "debug", Message: msg, Args: args})
}

func TestNewAuthService(t *testing.T) {
	mockUserService := NewMockUserService()
	mockTokenService := NewMockTokenService()
	mockLogger := NewMockLogger()

	authService := service.NewAuthService(mockUserService, mockTokenService, mockLogger)

	assert.NotNil(t, authService)
}

func TestAuthService_GenerateToken_Success(t *testing.T) {
	mockUserService := NewMockUserService()
	mockTokenService := NewMockTokenService()
	mockLogger := NewMockLogger()

	ctx := context.Background()
	user := &user.User{ID: 1, Email: "test@example.com"}
	authService := service.NewAuthService(mockUserService, mockTokenService, mockLogger)

	token, err := authService.GenerateToken(ctx, user)

	assert.NoError(t, err)
	assert.Equal(t, "mock_token_test@example.com", token)
}

func TestAuthService_GenerateToken_Error(t *testing.T) {
	mockUserService := NewMockUserService()
	mockTokenService := NewMockTokenService()
	mockLogger := NewMockLogger()

	mockTokenService.SetError(true)
	testUser := &user.User{ID: 1, Email: "test@example.com"}
	authService := service.NewAuthService(mockUserService, mockTokenService, mockLogger)

	token, err := authService.GenerateToken(context.Background(), testUser)

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "token generation failed")
}

func TestAuthService_ValidateToken_Success(t *testing.T) {
	mockUserService := NewMockUserService()
	mockTokenService := NewMockTokenService()
	mockLogger := NewMockLogger()

	ctx := context.Background()
	testUser := &user.User{ID: 1, Email: "test@example.com"}
	authService := service.NewAuthService(mockUserService, mockTokenService, mockLogger)

	// First generate a token to validate
	token, err := authService.GenerateToken(ctx, testUser)
	assert.NoError(t, err)

	// Now validate the generated token
	claims, err := authService.ValidateToken(ctx, token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, 1, claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
}

func TestAuthService_ValidateToken_InvalidToken(t *testing.T) {
	mockUserService := NewMockUserService()
	mockTokenService := NewMockTokenService()
	mockLogger := NewMockLogger()

	authService := service.NewAuthService(mockUserService, mockTokenService, mockLogger)

	claims, err := authService.ValidateToken(context.Background(), "invalid_token")

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, auth.ErrInvalidToken, err)
}

func TestAuthService_RefreshToken_Success(t *testing.T) {
	mockUserService := NewMockUserService()
	mockTokenService := NewMockTokenService()
	mockLogger := NewMockLogger()

	// First generate a token
	testUser := &user.User{ID: 1, Email: "test@example.com"}
	authService := service.NewAuthService(mockUserService, mockTokenService, mockLogger)
	token, err := authService.GenerateToken(context.Background(), testUser)
	assert.NoError(t, err)
	newToken, err := authService.RefreshToken(context.Background(), token)

	assert.NoError(t, err)
	assert.NotEmpty(t, newToken)
	assert.Contains(t, newToken, "refreshed_")
}

func TestAuthService_RefreshToken_InvalidToken(t *testing.T) {
	mockUserService := NewMockUserService()
	mockTokenService := NewMockTokenService()
	mockLogger := NewMockLogger()

	authService := service.NewAuthService(mockUserService, mockTokenService, mockLogger)

	newToken, err := authService.RefreshToken(context.Background(), "invalid_token")

	assert.Error(t, err)
	assert.Empty(t, newToken)
	assert.Equal(t, auth.ErrInvalidToken, err)
}
