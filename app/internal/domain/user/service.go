package user

import (
	"context"
)

// Service defines the interface for user business logic
type Service interface {
	Register(ctx context.Context, name, email, password string) (*User, error)
	Login(ctx context.Context, email, password string) (*User, error)
	GetByID(ctx context.Context, id int) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	UpdateProfile(ctx context.Context, id int, name, email string) (*User, error)
	UpdatePassword(ctx context.Context, id int, currentPassword, newPassword string) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, limit, offset int) ([]*User, error)
}
