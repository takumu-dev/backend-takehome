package user

import (
	"context"
)

// Service defines the interface for user business logic
type Service interface {
	Register(ctx context.Context, name, email, password string) (*User, error)
	Authenticate(ctx context.Context, email, password string) (*User, error)
	GetProfile(ctx context.Context, userID int) (*User, error)
}
