package user_test

import (
	"context"
	"testing"

	"blog-platform/internal/domain/user"
)

// MockUserService implements the UserService interface for testing
type MockUserService struct {
	repo user.Repository
}

func NewMockUserService(repo user.Repository) *MockUserService {
	return &MockUserService{repo: repo}
}

func (s *MockUserService) Register(ctx context.Context, name, email, password string) (*user.User, error) {
	// Check if user already exists
	existingUser, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, user.ErrUserExists
	}
	if err != nil && err != user.ErrUserNotFound {
		return nil, err
	}

	// Create new user
	u, err := user.NewUser(name, email, password)
	if err != nil {
		return nil, err
	}

	// Save to repository
	err = s.repo.Create(ctx, u)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *MockUserService) Login(ctx context.Context, email, password string) (*user.User, error) {
	// Get user by email
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if err == user.ErrUserNotFound {
			return nil, user.ErrInvalidCredentials
		}
		return nil, err
	}

	// Validate password
	if !u.ValidatePassword(password) {
		return nil, user.ErrInvalidCredentials
	}

	return u, nil
}

func (s *MockUserService) GetByID(ctx context.Context, id int) (*user.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *MockUserService) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *MockUserService) UpdateProfile(ctx context.Context, id int, name, email string) (*user.User, error) {
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
			return nil, err
		}
	}

	// Update profile
	err = u.UpdateProfile(name, email)
	if err != nil {
		return nil, err
	}

	// Save changes
	err = s.repo.Update(ctx, u)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *MockUserService) UpdatePassword(ctx context.Context, id int, currentPassword, newPassword string) error {
	// Get existing user
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Validate current password
	if !u.ValidatePassword(currentPassword) {
		return user.ErrInvalidCredentials
	}

	// Update password
	err = u.UpdatePassword(newPassword)
	if err != nil {
		return err
	}

	// Save changes
	return s.repo.Update(ctx, u)
}

func (s *MockUserService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *MockUserService) List(ctx context.Context, limit, offset int) ([]*user.User, error) {
	return s.repo.List(ctx, limit, offset)
}

func TestUserService_Register(t *testing.T) {
	repo := NewMockUserRepository()
	service := NewMockUserService(repo)
	ctx := context.Background()

	tests := []struct {
		name        string
		userName    string
		email       string
		password    string
		expectError bool
		errorType   error
	}{
		{
			name:        "valid registration",
			userName:    "John Doe",
			email:       "john@example.com",
			password:    "password123",
			expectError: false,
		},
		{
			name:        "invalid email should fail",
			userName:    "John Doe",
			email:       "invalid-email",
			password:    "password123",
			expectError: true,
		},
		{
			name:        "short password should fail",
			userName:    "John Doe",
			email:       "john@example.com",
			password:    "123",
			expectError: true,
		},
		{
			name:        "empty name should fail",
			userName:    "",
			email:       "john@example.com",
			password:    "password123",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := service.Register(ctx, tt.userName, tt.email, tt.password)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if u == nil {
				t.Error("expected user to be returned")
				return
			}

			if u.Name != tt.userName {
				t.Errorf("expected name '%s', got '%s'", tt.userName, u.Name)
			}

			if u.Email != tt.email {
				t.Errorf("expected email '%s', got '%s'", tt.email, u.Email)
			}

			if u.ID == 0 {
				t.Error("expected user ID to be set")
			}
		})
	}
}

func TestUserService_Register_DuplicateEmail(t *testing.T) {
	repo := NewMockUserRepository()
	service := NewMockUserService(repo)
	ctx := context.Background()

	// Register first user
	_, err := service.Register(ctx, "John Doe", "duplicate@example.com", "password123")
	if err != nil {
		t.Fatalf("failed to register first user: %v", err)
	}

	// Try to register second user with same email
	_, err = service.Register(ctx, "Jane Doe", "duplicate@example.com", "password456")
	if err != user.ErrUserExists {
		t.Errorf("expected ErrUserExists, got %v", err)
	}
}

func TestUserService_Login(t *testing.T) {
	repo := NewMockUserRepository()
	service := NewMockUserService(repo)
	ctx := context.Background()

	// Register a user first
	registeredUser, err := service.Register(ctx, "John Doe", "login@example.com", "password123")
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	tests := []struct {
		name        string
		email       string
		password    string
		expectError bool
		errorType   error
	}{
		{
			name:        "valid login",
			email:       "login@example.com",
			password:    "password123",
			expectError: false,
		},
		{
			name:        "wrong password should fail",
			email:       "login@example.com",
			password:    "wrongpassword",
			expectError: true,
			errorType:   user.ErrInvalidCredentials,
		},
		{
			name:        "non-existent user should fail",
			email:       "nonexistent@example.com",
			password:    "password123",
			expectError: true,
			errorType:   user.ErrInvalidCredentials,
		},
		{
			name:        "empty password should fail",
			email:       "login@example.com",
			password:    "",
			expectError: true,
			errorType:   user.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := service.Login(ctx, tt.email, tt.password)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorType != nil && err != tt.errorType {
					t.Errorf("expected error %v, got %v", tt.errorType, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if u == nil {
				t.Error("expected user to be returned")
				return
			}

			if u.ID != registeredUser.ID {
				t.Errorf("expected user ID %d, got %d", registeredUser.ID, u.ID)
			}

			if u.Email != tt.email {
				t.Errorf("expected email '%s', got '%s'", tt.email, u.Email)
			}
		})
	}
}

func TestUserService_UpdateProfile(t *testing.T) {
	repo := NewMockUserRepository()
	service := NewMockUserService(repo)
	ctx := context.Background()

	// Register a user first
	registeredUser, err := service.Register(ctx, "John Doe", "update@example.com", "password123")
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	tests := []struct {
		name        string
		newName     string
		newEmail    string
		expectError bool
	}{
		{
			name:        "valid profile update",
			newName:     "Jane Doe",
			newEmail:    "jane@example.com",
			expectError: false,
		},
		{
			name:        "invalid email should fail",
			newName:     "Jane Doe",
			newEmail:    "invalid-email",
			expectError: true,
		},
		{
			name:        "empty name should fail",
			newName:     "",
			newEmail:    "jane@example.com",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := service.UpdateProfile(ctx, registeredUser.ID, tt.newName, tt.newEmail)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if u.Name != tt.newName {
				t.Errorf("expected name '%s', got '%s'", tt.newName, u.Name)
			}

			if u.Email != tt.newEmail {
				t.Errorf("expected email '%s', got '%s'", tt.newEmail, u.Email)
			}
		})
	}
}

func TestUserService_UpdatePassword(t *testing.T) {
	repo := NewMockUserRepository()
	service := NewMockUserService(repo)
	ctx := context.Background()

	// Register a user first
	registeredUser, err := service.Register(ctx, "John Doe", "password@example.com", "oldpassword123")
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	tests := []struct {
		name            string
		currentPassword string
		newPassword     string
		expectError     bool
		errorType       error
	}{
		{
			name:            "valid password update",
			currentPassword: "oldpassword123",
			newPassword:     "newpassword123",
			expectError:     false,
		},
		{
			name:            "wrong current password should fail",
			currentPassword: "wrongpassword",
			newPassword:     "newpassword123",
			expectError:     true,
			errorType:       user.ErrInvalidCredentials,
		},
		{
			name:            "short new password should fail",
			currentPassword: "oldpassword123",
			newPassword:     "123",
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdatePassword(ctx, registeredUser.ID, tt.currentPassword, tt.newPassword)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorType != nil && err != tt.errorType {
					t.Errorf("expected error %v, got %v", tt.errorType, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify password was changed by trying to login with new password
			_, err = service.Login(ctx, registeredUser.Email, tt.newPassword)
			if err != nil {
				t.Errorf("failed to login with new password: %v", err)
			}
		})
	}
}
