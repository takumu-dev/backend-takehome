package service

import (
	"context"
	"fmt"

	"blog-platform/internal/domain/user"
)

// UserService implements the user.Service interface
type UserService struct {
	repo user.Repository
}

// NewUserService creates a new UserService instance
func NewUserService(repo user.Repository) *UserService {
	return &UserService{repo: repo}
}

// Register creates a new user account
func (s *UserService) Register(ctx context.Context, name, email, password string) (*user.User, error) {
	// Check if user already exists
	existingUser, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, user.ErrUserExists
	}
	if err != nil && err != user.ErrUserNotFound {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Create new user with validation and password hashing
	u, err := user.NewUser(name, email, password)
	if err != nil {
		return nil, err
	}

	// Save to repository
	err = s.repo.Create(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return u, nil
}

// Login authenticates a user with email and password
func (s *UserService) Login(ctx context.Context, email, password string) (*user.User, error) {
	// Get user by email
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if err == user.ErrUserNotFound {
			return nil, user.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Validate password
	if !u.ValidatePassword(password) {
		return nil, user.ErrInvalidCredentials
	}

	return u, nil
}

// GetByID retrieves a user by their ID
func (s *UserService) GetByID(ctx context.Context, id int) (*user.User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// GetByEmail retrieves a user by their email
func (s *UserService) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// UpdateProfile updates a user's profile information
func (s *UserService) UpdateProfile(ctx context.Context, id int, name, email string) (*user.User, error) {
	// Get existing user
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if email is being changed and if new email already exists
	if u.Email != email {
		existingUser, err := s.repo.GetByEmail(ctx, email)
		if err == nil && existingUser != nil && existingUser.ID != id {
			return nil, user.ErrUserExists
		}
		if err != nil && err != user.ErrUserNotFound {
			return nil, fmt.Errorf("failed to check existing email: %w", err)
		}
	}

	// Update profile with validation
	err = u.UpdateProfile(name, email)
	if err != nil {
		return nil, err
	}

	// Save changes
	err = s.repo.Update(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return u, nil
}

// UpdatePassword updates a user's password
func (s *UserService) UpdatePassword(ctx context.Context, id int, currentPassword, newPassword string) error {
	// Get existing user
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Validate current password
	if !u.ValidatePassword(currentPassword) {
		return user.ErrInvalidCredentials
	}

	// Update password with validation and hashing
	err = u.UpdatePassword(newPassword)
	if err != nil {
		return err
	}

	// Save changes
	err = s.repo.Update(ctx, u)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// Delete removes a user account
func (s *UserService) Delete(ctx context.Context, id int) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// List retrieves a paginated list of users
func (s *UserService) List(ctx context.Context, limit, offset int) ([]*user.User, error) {
	// Validate pagination parameters
	if limit <= 0 {
		limit = 10 // Default limit
	}
	if limit > 100 {
		limit = 100 // Maximum limit
	}
	if offset < 0 {
		offset = 0
	}

	users, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}
