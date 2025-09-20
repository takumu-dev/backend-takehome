package service

import (
	"context"
	"fmt"

	"blog-platform/internal/application/logging"
	"blog-platform/internal/domain/user"
)

// UserService implements the user.Service interface
type UserService struct {
	repo   user.Repository
	logger logging.Logger
}

// NewUserService creates a new UserService instance
func NewUserService(repo user.Repository, logger logging.Logger) *UserService {
	return &UserService{
		repo:   repo,
		logger: logger,
	}
}

// Register creates a new user account
func (s *UserService) Register(ctx context.Context, name, email, password string) (*user.User, error) {
	s.logger.Info(ctx, "registering new user", "email", email, "name", name)
	
	// Check if user already exists
	existingUser, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		s.logger.Warn(ctx, "registration attempt for existing user", "email", email)
		return nil, user.ErrUserExists
	}
	if err != nil && err != user.ErrUserNotFound {
		s.logger.Error(ctx, "failed to check existing user during registration", "email", email, "error", err.Error())
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Create new user with validation and password hashing
	u, err := user.NewUser(name, email, password)
	if err != nil {
		s.logger.Error(ctx, "failed to create user entity", "email", email, "name", name, "error", err.Error())
		return nil, err
	}

	// Save to repository
	err = s.repo.Create(ctx, u)
	if err != nil {
		s.logger.Error(ctx, "failed to save user to repository", "email", email, "userID", u.ID, "error", err.Error())
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info(ctx, "user registered successfully", "email", email, "userID", u.ID)
	return u, nil
}

// Login authenticates a user with email and password
func (s *UserService) Login(ctx context.Context, email, password string) (*user.User, error) {
	s.logger.Info(ctx, "user login attempt", "email", email)
	
	// Get user by email
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if err == user.ErrUserNotFound {
			s.logger.Warn(ctx, "login attempt with non-existent email", "email", email)
			return nil, user.ErrInvalidCredentials
		}
		s.logger.Error(ctx, "failed to retrieve user during login", "email", email, "error", err.Error())
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Validate password
	if !u.ValidatePassword(password) {
		s.logger.Warn(ctx, "login attempt with invalid password", "email", email, "userID", u.ID)
		return nil, user.ErrInvalidCredentials
	}

	s.logger.Info(ctx, "user login successful", "email", email, "userID", u.ID)
	return u, nil
}

// GetByID retrieves a user by their ID
func (s *UserService) GetByID(ctx context.Context, id int) (*user.User, error) {
	s.logger.Debug(ctx, "retrieving user by ID", "userID", id)
	
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "failed to retrieve user by ID", "userID", id, "error", err.Error())
		return nil, err
	}
	
	s.logger.Debug(ctx, "user retrieved successfully by ID", "userID", id)
	return u, nil
}

// GetByEmail retrieves a user by their email
func (s *UserService) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	s.logger.Debug(ctx, "retrieving user by email", "email", email)
	
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		s.logger.Error(ctx, "failed to retrieve user by email", "email", email, "error", err.Error())
		return nil, err
	}
	
	s.logger.Debug(ctx, "user retrieved successfully by email", "email", email, "userID", u.ID)
	return u, nil
}

// UpdateProfile updates a user's profile information
func (s *UserService) UpdateProfile(ctx context.Context, id int, name, email string) (*user.User, error) {
	s.logger.Info(ctx, "updating user profile", "userID", id, "newEmail", email, "newName", name)
	
	// Get existing user
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "failed to retrieve user for profile update", "userID", id, "error", err.Error())
		return nil, err
	}

	// Check if email is being changed and if new email already exists
	if u.Email != email {
		s.logger.Debug(ctx, "email change detected, checking for conflicts", "userID", id, "oldEmail", u.Email, "newEmail", email)
		existingUser, err := s.repo.GetByEmail(ctx, email)
		if err == nil && existingUser != nil && existingUser.ID != id {
			s.logger.Warn(ctx, "profile update attempt with existing email", "userID", id, "conflictingEmail", email, "existingUserID", existingUser.ID)
			return nil, user.ErrUserExists
		}
		if err != nil && err != user.ErrUserNotFound {
			s.logger.Error(ctx, "failed to check email conflict during profile update", "userID", id, "email", email, "error", err.Error())
			return nil, fmt.Errorf("failed to check existing email: %w", err)
		}
	}

	// Update profile with validation
	err = u.UpdateProfile(name, email)
	if err != nil {
		s.logger.Error(ctx, "failed to update user profile entity", "userID", id, "error", err.Error())
		return nil, err
	}

	// Save changes
	err = s.repo.Update(ctx, u)
	if err != nil {
		s.logger.Error(ctx, "failed to save profile updates", "userID", id, "error", err.Error())
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	s.logger.Info(ctx, "user profile updated successfully", "userID", id, "email", email)
	return u, nil
}

// UpdatePassword updates a user's password
func (s *UserService) UpdatePassword(ctx context.Context, id int, currentPassword, newPassword string) error {
	s.logger.Info(ctx, "updating user password", "userID", id)
	
	// Get existing user
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "failed to retrieve user for password update", "userID", id, "error", err.Error())
		return err
	}

	// Validate current password
	if !u.ValidatePassword(currentPassword) {
		s.logger.Warn(ctx, "password update attempt with invalid current password", "userID", id)
		return user.ErrInvalidCredentials
	}

	// Update password with validation and hashing
	err = u.UpdatePassword(newPassword)
	if err != nil {
		s.logger.Error(ctx, "failed to update password entity", "userID", id, "error", err.Error())
		return err
	}

	// Save changes
	err = s.repo.Update(ctx, u)
	if err != nil {
		s.logger.Error(ctx, "failed to save password update", "userID", id, "error", err.Error())
		return fmt.Errorf("failed to update password: %w", err)
	}

	s.logger.Info(ctx, "user password updated successfully", "userID", id)
	return nil
}

// Delete removes a user account
func (s *UserService) Delete(ctx context.Context, id int) error {
	s.logger.Info(ctx, "deleting user account", "userID", id)
	
	err := s.repo.Delete(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "failed to delete user account", "userID", id, "error", err.Error())
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	s.logger.Info(ctx, "user account deleted successfully", "userID", id)
	return nil
}

// List retrieves a paginated list of users
func (s *UserService) List(ctx context.Context, limit, offset int) ([]*user.User, error) {
	s.logger.Debug(ctx, "listing users", "limit", limit, "offset", offset)
	
	// Validate pagination parameters
	originalLimit := limit
	if limit <= 0 {
		limit = 10 // Default limit
	}
	if limit > 100 {
		limit = 100 // Maximum limit
	}
	if offset < 0 {
		offset = 0
	}
	
	if originalLimit != limit || offset < 0 {
		s.logger.Debug(ctx, "pagination parameters adjusted", "originalLimit", originalLimit, "adjustedLimit", limit, "adjustedOffset", offset)
	}

	users, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		s.logger.Error(ctx, "failed to retrieve user list", "limit", limit, "offset", offset, "error", err.Error())
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	s.logger.Debug(ctx, "user list retrieved successfully", "count", len(users), "limit", limit, "offset", offset)
	return users, nil
}
