package auth_test

import (
	"testing"
	"time"

	"blog-platform/internal/domain/auth"
)

// MockTokenService implements the auth.TokenService interface for testing
type MockTokenService struct {
	secretKey string
}

// NewMockTokenService creates a new mock token service
func NewMockTokenService(secretKey string) *MockTokenService {
	return &MockTokenService{
		secretKey: secretKey,
	}
}

// GenerateToken creates a JWT token with the given claims
func (m *MockTokenService) GenerateToken(userID int, email string, duration time.Duration) (string, error) {
	// For testing, we'll create a simple mock token
	// In real implementation, this would use JWT library
	if userID <= 0 {
		return "", auth.ErrInvalidUserID
	}
	if email == "" {
		return "", auth.ErrInvalidEmail
	}
	if duration <= 0 {
		return "", auth.ErrInvalidDuration
	}
	if m.secretKey == "" {
		return "", auth.ErrInvalidSecretKey
	}

	// Mock token format: "token_userID_email_duration"
	return "mock_token_" + string(rune(userID)) + "_" + email, nil
}

// ValidateToken validates a JWT token and returns claims
func (m *MockTokenService) ValidateToken(token string) (*auth.TokenClaims, error) {
	if token == "" {
		return nil, auth.ErrEmptyToken
	}
	if m.secretKey == "" {
		return nil, auth.ErrInvalidSecretKey
	}

	// Mock validation - in real implementation, this would parse JWT
	if token == "invalid_token" {
		return nil, auth.ErrInvalidToken
	}
	if token == "expired_token" {
		return nil, auth.ErrTokenExpired
	}

	// Return mock claims for valid tokens
	now := time.Now()
	return &auth.TokenClaims{
		UserID:    1,
		Email:     "test@example.com",
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(24 * time.Hour).Unix(),
	}, nil
}

// RefreshToken generates a new token from an existing valid token
func (m *MockTokenService) RefreshToken(token string) (string, error) {
	claims, err := m.ValidateToken(token)
	if err != nil {
		return "", err
	}

	// Generate new token with same user info but extended expiry
	return m.GenerateToken(claims.UserID, claims.Email, 24*time.Hour)
}

func TestTokenService_GenerateToken(t *testing.T) {
	service := NewMockTokenService("test-secret-key")

	tests := []struct {
		name        string
		userID      int
		email       string
		duration    time.Duration
		expectError bool
		expectedErr error
	}{
		{
			name:        "valid token generation",
			userID:      1,
			email:       "test@example.com",
			duration:    24 * time.Hour,
			expectError: false,
		},
		{
			name:        "invalid user ID should fail",
			userID:      0,
			email:       "test@example.com",
			duration:    24 * time.Hour,
			expectError: true,
			expectedErr: auth.ErrInvalidUserID,
		},
		{
			name:        "empty email should fail",
			userID:      1,
			email:       "",
			duration:    24 * time.Hour,
			expectError: true,
			expectedErr: auth.ErrInvalidEmail,
		},
		{
			name:        "invalid duration should fail",
			userID:      1,
			email:       "test@example.com",
			duration:    0,
			expectError: true,
			expectedErr: auth.ErrInvalidDuration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.GenerateToken(tt.userID, tt.email, tt.duration)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if err != tt.expectedErr {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("expected no error, got %v", err)
				return
			}

			if token == "" {
				t.Error("expected token to be generated, got empty string")
			}
		})
	}
}

func TestTokenService_ValidateToken(t *testing.T) {
	service := NewMockTokenService("test-secret-key")

	tests := []struct {
		name        string
		token       string
		expectError bool
		expectedErr error
	}{
		{
			name:        "valid token validation",
			token:       "valid_token",
			expectError: false,
		},
		{
			name:        "empty token should fail",
			token:       "",
			expectError: true,
			expectedErr: auth.ErrEmptyToken,
		},
		{
			name:        "invalid token should fail",
			token:       "invalid_token",
			expectError: true,
			expectedErr: auth.ErrInvalidToken,
		},
		{
			name:        "expired token should fail",
			token:       "expired_token",
			expectError: true,
			expectedErr: auth.ErrTokenExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := service.ValidateToken(tt.token)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if err != tt.expectedErr {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("expected no error, got %v", err)
				return
			}

			if claims == nil {
				t.Error("expected claims to be returned, got nil")
				return
			}

			if claims.UserID <= 0 {
				t.Errorf("expected valid user ID, got %d", claims.UserID)
			}

			if claims.Email == "" {
				t.Error("expected email to be set in claims")
			}

			if claims.IssuedAt <= 0 {
				t.Error("expected issued at to be set in claims")
			}

			if claims.ExpiresAt <= claims.IssuedAt {
				t.Error("expected expires at to be after issued at")
			}
		})
	}
}

func TestTokenService_RefreshToken(t *testing.T) {
	service := NewMockTokenService("test-secret-key")

	tests := []struct {
		name        string
		token       string
		expectError bool
		expectedErr error
	}{
		{
			name:        "valid token refresh",
			token:       "valid_token",
			expectError: false,
		},
		{
			name:        "invalid token refresh should fail",
			token:       "invalid_token",
			expectError: true,
			expectedErr: auth.ErrInvalidToken,
		},
		{
			name:        "expired token refresh should fail",
			token:       "expired_token",
			expectError: true,
			expectedErr: auth.ErrTokenExpired,
		},
		{
			name:        "empty token refresh should fail",
			token:       "",
			expectError: true,
			expectedErr: auth.ErrEmptyToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newToken, err := service.RefreshToken(tt.token)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if err != tt.expectedErr {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("expected no error, got %v", err)
				return
			}

			if newToken == "" {
				t.Error("expected new token to be generated, got empty string")
			}

			if newToken == tt.token {
				t.Error("expected new token to be different from original token")
			}
		})
	}
}

func TestTokenService_InvalidSecretKey(t *testing.T) {
	service := NewMockTokenService("")

	// Test that operations fail with invalid secret key
	_, err := service.GenerateToken(1, "test@example.com", 24*time.Hour)
	if err != auth.ErrInvalidSecretKey {
		t.Errorf("expected ErrInvalidSecretKey, got %v", err)
	}

	_, err = service.ValidateToken("valid_token")
	if err != auth.ErrInvalidSecretKey {
		t.Errorf("expected ErrInvalidSecretKey, got %v", err)
	}
}

func TestTokenService_Interface(t *testing.T) {
	// Verify that MockTokenService implements the TokenService interface
	var _ auth.TokenService = (*MockTokenService)(nil)
}
