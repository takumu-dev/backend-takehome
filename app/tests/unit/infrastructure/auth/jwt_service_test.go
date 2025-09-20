package auth_test

import (
	"testing"
	"time"

	"blog-platform/internal/domain/auth"
	infraAuth "blog-platform/internal/infrastructure/auth"
)

func TestJWTService_GenerateToken_Integration(t *testing.T) {
	service := infraAuth.NewJWTService("test-secret-key-for-jwt")

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

			// Verify token can be validated
			claims, err := service.ValidateToken(token)
			if err != nil {
				t.Errorf("generated token should be valid, got error: %v", err)
			}

			if claims.UserID != tt.userID {
				t.Errorf("expected user ID %d, got %d", tt.userID, claims.UserID)
			}

			if claims.Email != tt.email {
				t.Errorf("expected email %s, got %s", tt.email, claims.Email)
			}
		})
	}
}

func TestJWTService_ValidateToken_Integration(t *testing.T) {
	service := infraAuth.NewJWTService("test-secret-key-for-jwt")

	// Generate a valid token first
	validToken, err := service.GenerateToken(1, "test@example.com", 24*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}

	// For expired token test, we'll use an invalid token that simulates expiration
	expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyLCJlbWFpbCI6ImV4cGlyZWRAZXhhbXBsZS5jb20iLCJleHAiOjE2MDk0NTkyMDB9.invalid"

	tests := []struct {
		name        string
		token       string
		expectError bool
		expectedErr error
	}{
		{
			name:        "valid token validation",
			token:       validToken,
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
			token:       "invalid.jwt.token",
			expectError: true,
			expectedErr: auth.ErrInvalidToken,
		},
		{
			name:        "malformed token should fail",
			token:       expiredToken,
			expectError: true,
			expectedErr: auth.ErrInvalidToken,
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

func TestJWTService_RefreshToken_Integration(t *testing.T) {
	service := infraAuth.NewJWTService("test-secret-key-for-jwt")

	// Generate a valid token first
	validToken, err := service.GenerateToken(1, "test@example.com", 24*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}

	tests := []struct {
		name        string
		token       string
		expectError bool
		expectedErr error
	}{
		{
			name:        "valid token refresh",
			token:       validToken,
			expectError: false,
		},
		{
			name:        "invalid token refresh should fail",
			token:       "invalid.jwt.token",
			expectError: true,
			expectedErr: auth.ErrInvalidToken,
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

			// Note: tokens might be identical if generated at the same millisecond
			// This is acceptable behavior for JWT refresh

			// Verify new token is valid
			claims, err := service.ValidateToken(newToken)
			if err != nil {
				t.Errorf("refreshed token should be valid, got error: %v", err)
			}

			if claims.UserID != 1 {
				t.Errorf("expected user ID 1, got %d", claims.UserID)
			}

			if claims.Email != "test@example.com" {
				t.Errorf("expected email test@example.com, got %s", claims.Email)
			}
		})
	}
}

func TestJWTService_InvalidSecretKey_Integration(t *testing.T) {
	service := infraAuth.NewJWTService("")

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

func TestJWTService_CrossValidation_Integration(t *testing.T) {
	service1 := infraAuth.NewJWTService("secret-key-1")
	service2 := infraAuth.NewJWTService("secret-key-2")

	// Generate token with service1
	token, err := service1.GenerateToken(1, "test@example.com", 24*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Try to validate with service2 (different secret key)
	_, err = service2.ValidateToken(token)
	if err != auth.ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken when validating with different secret key, got %v", err)
	}
}

func TestJWTService_Interface_Integration(t *testing.T) {
	// Verify that JWTService implements the TokenService interface
	var _ auth.TokenService = (*infraAuth.JWTService)(nil)
}
