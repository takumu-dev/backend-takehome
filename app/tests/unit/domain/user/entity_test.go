package user_test

import (
	"testing"
	"time"

	"blog-platform/internal/domain/user"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name        string
		userName    string
		email       string
		password    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid user creation",
			userName:    "John Doe",
			email:       "john@example.com",
			password:    "password123",
			expectError: false,
		},
		{
			name:        "empty name should fail",
			userName:    "",
			email:       "john@example.com",
			password:    "password123",
			expectError: true,
			errorMsg:    "name cannot be empty",
		},
		{
			name:        "empty email should fail",
			userName:    "John Doe",
			email:       "",
			password:    "password123",
			expectError: true,
			errorMsg:    "email cannot be empty",
		},
		{
			name:        "invalid email format should fail",
			userName:    "John Doe",
			email:       "invalid-email",
			password:    "password123",
			expectError: true,
			errorMsg:    "invalid email format",
		},
		{
			name:        "empty password should fail",
			userName:    "John Doe",
			email:       "john@example.com",
			password:    "",
			expectError: true,
			errorMsg:    "password cannot be empty",
		},
		{
			name:        "short password should fail",
			userName:    "John Doe",
			email:       "john@example.com",
			password:    "123",
			expectError: true,
			errorMsg:    "password must be at least 6 characters long",
		},
		{
			name:        "name too long should fail",
			userName:    string(make([]byte, 256)), // 256 characters
			email:       "john@example.com",
			password:    "password123",
			expectError: true,
			errorMsg:    "name must be less than 255 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := user.NewUser(tt.userName, tt.email, tt.password)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if u.Name != tt.userName {
				t.Errorf("expected name '%s', got '%s'", tt.userName, u.Name)
			}

			if u.Email != tt.email {
				t.Errorf("expected email '%s', got '%s'", tt.email, u.Email)
			}

			if u.PasswordHash == "" {
				t.Error("expected password hash to be set")
			}

			if u.PasswordHash == tt.password {
				t.Error("password should be hashed, not stored in plain text")
			}

			if u.CreatedAt.IsZero() {
				t.Error("expected CreatedAt to be set")
			}

			if u.UpdatedAt.IsZero() {
				t.Error("expected UpdatedAt to be set")
			}

			if u.ID != 0 {
				t.Error("expected ID to be 0 for new user")
			}
		})
	}
}

func TestUser_UpdateProfile(t *testing.T) {
	// Create a valid user first
	u, err := user.NewUser("John Doe", "john@example.com", "password123")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	originalUpdatedAt := u.UpdatedAt
	time.Sleep(1 * time.Millisecond) // Ensure time difference

	tests := []struct {
		name        string
		newName     string
		newEmail    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid profile update",
			newName:     "Jane Doe",
			newEmail:    "jane@example.com",
			expectError: false,
		},
		{
			name:        "empty name should fail",
			newName:     "",
			newEmail:    "jane@example.com",
			expectError: true,
			errorMsg:    "name cannot be empty",
		},
		{
			name:        "empty email should fail",
			newName:     "Jane Doe",
			newEmail:    "",
			expectError: true,
			errorMsg:    "email cannot be empty",
		},
		{
			name:        "invalid email format should fail",
			newName:     "Jane Doe",
			newEmail:    "invalid-email",
			expectError: true,
			errorMsg:    "invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh user for each test
			testUser, _ := user.NewUser("John Doe", "john@example.com", "password123")
			
			err := testUser.UpdateProfile(tt.newName, tt.newEmail)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if testUser.Name != tt.newName {
				t.Errorf("expected name '%s', got '%s'", tt.newName, testUser.Name)
			}

			if testUser.Email != tt.newEmail {
				t.Errorf("expected email '%s', got '%s'", tt.newEmail, testUser.Email)
			}

			if !testUser.UpdatedAt.After(originalUpdatedAt) {
				t.Error("expected UpdatedAt to be updated")
			}
		})
	}
}

func TestUser_ValidatePassword(t *testing.T) {
	u, err := user.NewUser("John Doe", "john@example.com", "password123")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{
			name:     "correct password should return true",
			password: "password123",
			expected: true,
		},
		{
			name:     "incorrect password should return false",
			password: "wrongpassword",
			expected: false,
		},
		{
			name:     "empty password should return false",
			password: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := u.ValidatePassword(tt.password)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestUser_EmailValidation(t *testing.T) {
	validEmails := []string{
		"test@example.com",
		"user.name@domain.co.uk",
		"user+tag@example.org",
		"123@example.com",
	}

	invalidEmails := []string{
		"invalid-email",
		"@example.com",
		"test@",
		"test.example.com",
		"test@.com",
		"test@example.",
	}

	for _, email := range validEmails {
		t.Run("valid_"+email, func(t *testing.T) {
			_, err := user.NewUser("Test User", email, "password123")
			if err != nil {
				t.Errorf("expected valid email '%s' to pass validation, got error: %v", email, err)
			}
		})
	}

	for _, email := range invalidEmails {
		t.Run("invalid_"+email, func(t *testing.T) {
			_, err := user.NewUser("Test User", email, "password123")
			if err == nil {
				t.Errorf("expected invalid email '%s' to fail validation", email)
			}
		})
	}
}
