package auth

import (
	"context"
	"time"

	"blog-platform/internal/domain/user"
)

// PasswordValidator defines the interface for password validation operations
type PasswordValidator interface {
	ValidatePassword(password string) error
	HasSequentialChars(password string) bool
	HasRepeatedChars(password string) bool
}

// AuthService defines the interface for authentication operations
type AuthService interface {
	GenerateToken(ctx context.Context, user *user.User) (string, error)
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
	Login(ctx context.Context, email, password string) (*user.User, string, error)
	Register(ctx context.Context, name, email, password string) (*user.User, string, error)
	RefreshToken(ctx context.Context, token string) (string, error)
}

// TokenService defines the interface for JWT token operations
type TokenService interface {
	GenerateToken(userID int, email string, duration time.Duration) (string, error)
	ValidateToken(token string) (*TokenClaims, error)
	RefreshToken(token string) (string, error)
}
