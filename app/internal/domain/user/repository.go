package user

import (
	"context"
	"errors"
)

// Repository errors
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidUserData    = errors.New("invalid user data")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Repository defines the interface for user data access
type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id int) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, limit, offset int) ([]*User, error)
}
