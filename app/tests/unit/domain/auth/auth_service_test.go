package auth_test

import (
	"context"
	"testing"

	"blog-platform/internal/domain/auth"
	"blog-platform/internal/domain/user"
)

// MockUserRepository implements the user.Repository interface for testing
type MockUserRepository struct {
	users  map[int]*user.User
	emails map[string]*user.User
	nextID int
}

// NewMockUserRepository creates a new mock user repository
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:  make(map[int]*user.User),
		emails: make(map[string]*user.User),
		nextID: 1,
	}
}

// Create adds a new user to the mock repository
func (m *MockUserRepository) Create(ctx context.Context, u *user.User) error {
	u.ID = m.nextID
	m.users[u.ID] = u
	m.emails[u.Email] = u
	m.nextID++
	return nil
}

// GetByID retrieves a user by ID
func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*user.User, error) {
	u, exists := m.users[id]
	if !exists {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

// GetByEmail retrieves a user by email
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	u, exists := m.emails[email]
	if !exists {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

// Update modifies an existing user
func (m *MockUserRepository) Update(ctx context.Context, u *user.User) error {
	if _, exists := m.users[u.ID]; !exists {
		return user.ErrUserNotFound
	}
	m.users[u.ID] = u
	m.emails[u.Email] = u
	return nil
}

// List retrieves users with pagination
func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*user.User, error) {
	var result []*user.User
	count := 0
	
	for _, u := range m.users {
		if count >= offset {
			result = append(result, u)
			if len(result) >= limit {
				break
			}
		}
		count++
	}
	
	return result, nil
}

// Delete removes a user from the repository
func (m *MockUserRepository) Delete(ctx context.Context, id int) error {
	u, exists := m.users[id]
	if !exists {
		return user.ErrUserNotFound
	}
	delete(m.users, id)
	delete(m.emails, u.Email)
	return nil
}

// MockAuthService implements the auth.AuthService interface for testing
type MockAuthService struct {
	userRepo    user.Repository
	tokenService auth.TokenService
}

// NewMockAuthService creates a new mock auth service
func NewMockAuthService(userRepo user.Repository, tokenService auth.TokenService) *MockAuthService {
	return &MockAuthService{
		userRepo:     userRepo,
		tokenService: tokenService,
	}
}

// GenerateToken generates a JWT token for a user
func (s *MockAuthService) GenerateToken(ctx context.Context, u *user.User) (string, error) {
	if u == nil {
		return "", auth.ErrInvalidCredentials
	}
	return s.tokenService.GenerateToken(u.ID, u.Email, 24*3600) // 24 hours in seconds
}

// ValidateToken validates a JWT token
func (s *MockAuthService) ValidateToken(ctx context.Context, token string) (*auth.TokenClaims, error) {
	return s.tokenService.ValidateToken(token)
}

// Login authenticates a user with email and password
func (s *MockAuthService) Login(ctx context.Context, email, password string) (*user.User, string, error) {
	// Validate inputs
	if email == "" {
		return nil, "", auth.ErrInvalidEmail
	}
	if password == "" {
		return nil, "", auth.ErrInvalidCredentials
	}

	// Get user by email
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if err == user.ErrUserNotFound {
			return nil, "", auth.ErrUserNotFound
		}
		return nil, "", err
	}

	// Check password
	if !u.ValidatePassword(password) {
		return nil, "", auth.ErrInvalidCredentials
	}

	// Generate token
	token, err := s.GenerateToken(ctx, u)
	if err != nil {
		return nil, "", err
	}

	return u, token, nil
}

// Register creates a new user account
func (s *MockAuthService) Register(ctx context.Context, name, email, password string) (*user.User, string, error) {
	// Validate inputs
	if name == "" {
		return nil, "", user.ErrInvalidUserData
	}
	if email == "" {
		return nil, "", auth.ErrInvalidEmail
	}
	if password == "" {
		return nil, "", auth.ErrWeakPassword
	}

	// Check if user already exists
	_, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil {
		return nil, "", auth.ErrEmailAlreadyExists
	}
	if err != user.ErrUserNotFound {
		return nil, "", err
	}

	// Create new user
	u, err := user.NewUser(name, email, password)
	if err != nil {
		return nil, "", err
	}

	// Save user
	err = s.userRepo.Create(ctx, u)
	if err != nil {
		return nil, "", err
	}

	// Generate token
	token, err := s.GenerateToken(ctx, u)
	if err != nil {
		return nil, "", err
	}

	return u, token, nil
}

// RefreshToken generates a new token from an existing token
func (s *MockAuthService) RefreshToken(ctx context.Context, token string) (string, error) {
	return s.tokenService.RefreshToken(token)
}

func TestAuthService_Login(t *testing.T) {
	// Setup mock dependencies
	userRepo := NewMockUserRepository()
	tokenService := NewMockTokenService("test-secret")
	authService := NewMockAuthService(userRepo, tokenService)
	ctx := context.Background()

	// Create a test user
	testUser, err := user.NewUser("John Doe", "john@example.com", "password123")
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	err = userRepo.Create(ctx, testUser)
	if err != nil {
		t.Fatalf("failed to save test user: %v", err)
	}

	tests := []struct {
		name        string
		email       string
		password    string
		expectError bool
		expectedErr error
	}{
		{
			name:        "valid login",
			email:       "john@example.com",
			password:    "password123",
			expectError: false,
		},
		{
			name:        "empty email should fail",
			email:       "",
			password:    "password123",
			expectError: true,
			expectedErr: auth.ErrInvalidEmail,
		},
		{
			name:        "empty password should fail",
			email:       "john@example.com",
			password:    "",
			expectError: true,
			expectedErr: auth.ErrInvalidCredentials,
		},
		{
			name:        "non-existent user should fail",
			email:       "nonexistent@example.com",
			password:    "password123",
			expectError: true,
			expectedErr: auth.ErrUserNotFound,
		},
		{
			name:        "wrong password should fail",
			email:       "john@example.com",
			password:    "wrongpassword",
			expectError: true,
			expectedErr: auth.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, token, err := authService.Login(ctx, tt.email, tt.password)

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

			if u == nil {
				t.Error("expected user to be returned, got nil")
				return
			}

			if token == "" {
				t.Error("expected token to be generated, got empty string")
			}

			if u.Email != tt.email {
				t.Errorf("expected email %s, got %s", tt.email, u.Email)
			}
		})
	}
}

func TestAuthService_Register(t *testing.T) {
	// Setup mock dependencies
	userRepo := NewMockUserRepository()
	tokenService := NewMockTokenService("test-secret")
	authService := NewMockAuthService(userRepo, tokenService)
	ctx := context.Background()

	// Create an existing user to test duplicate email
	existingUser, err := user.NewUser("Existing User", "existing@example.com", "password123")
	if err != nil {
		t.Fatalf("failed to create existing user: %v", err)
	}
	err = userRepo.Create(ctx, existingUser)
	if err != nil {
		t.Fatalf("failed to save existing user: %v", err)
	}

	tests := []struct {
		name        string
		userName    string
		email       string
		password    string
		expectError bool
		expectedErr error
	}{
		{
			name:        "valid registration",
			userName:    "New User",
			email:       "newuser@example.com",
			password:    "password123",
			expectError: false,
		},
		{
			name:        "empty name should fail",
			userName:    "",
			email:       "test@example.com",
			password:    "password123",
			expectError: true,
			expectedErr: user.ErrInvalidUserData,
		},
		{
			name:        "empty email should fail",
			userName:    "Test User",
			email:       "",
			password:    "password123",
			expectError: true,
			expectedErr: auth.ErrInvalidEmail,
		},
		{
			name:        "empty password should fail",
			userName:    "Test User",
			email:       "test@example.com",
			password:    "",
			expectError: true,
			expectedErr: auth.ErrWeakPassword,
		},
		{
			name:        "duplicate email should fail",
			userName:    "Another User",
			email:       "existing@example.com",
			password:    "password123",
			expectError: true,
			expectedErr: auth.ErrEmailAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, token, err := authService.Register(ctx, tt.userName, tt.email, tt.password)

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

			if u == nil {
				t.Error("expected user to be returned, got nil")
				return
			}

			if token == "" {
				t.Error("expected token to be generated, got empty string")
			}

			if u.Name != tt.userName {
				t.Errorf("expected name %s, got %s", tt.userName, u.Name)
			}

			if u.Email != tt.email {
				t.Errorf("expected email %s, got %s", tt.email, u.Email)
			}

			// Verify user was saved to repository
			savedUser, err := userRepo.GetByEmail(ctx, tt.email)
			if err != nil {
				t.Errorf("failed to retrieve saved user: %v", err)
			}

			if savedUser.ID != u.ID {
				t.Errorf("expected saved user ID %d, got %d", u.ID, savedUser.ID)
			}
		})
	}
}

func TestAuthService_GenerateToken(t *testing.T) {
	tokenService := NewMockTokenService("test-secret")
	authService := NewMockAuthService(nil, tokenService)
	ctx := context.Background()

	// Create a test user
	testUser, err := user.NewUser("John Doe", "john@example.com", "password123")
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	testUser.ID = 1 // Set ID for token generation

	tests := []struct {
		name        string
		user        *user.User
		expectError bool
		expectedErr error
	}{
		{
			name:        "valid token generation",
			user:        testUser,
			expectError: false,
		},
		{
			name:        "nil user should fail",
			user:        nil,
			expectError: true,
			expectedErr: auth.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := authService.GenerateToken(ctx, tt.user)

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

func TestAuthService_ValidateToken(t *testing.T) {
	tokenService := NewMockTokenService("test-secret")
	authService := NewMockAuthService(nil, tokenService)
	ctx := context.Background()

	// Generate a valid token first
	validToken, err := tokenService.GenerateToken(1, "test@example.com", 24*3600)
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}

	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "valid token validation",
			token:       validToken,
			expectError: false,
		},
		{
			name:        "invalid token should fail",
			token:       "invalid_token",
			expectError: true,
		},
		{
			name:        "empty token should fail",
			token:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := authService.ValidateToken(ctx, tt.token)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				return
			}

			if err != nil {
				t.Errorf("expected no error, got %v", err)
				return
			}

			if claims == nil {
				t.Error("expected claims to be returned, got nil")
			}
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	tokenService := NewMockTokenService("test-secret")
	authService := NewMockAuthService(nil, tokenService)
	ctx := context.Background()

	// Generate a valid token first
	validToken, err := tokenService.GenerateToken(1, "test@example.com", 24*3600)
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}

	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "valid token refresh",
			token:       validToken,
			expectError: false,
		},
		{
			name:        "invalid token refresh should fail",
			token:       "invalid_token",
			expectError: true,
		},
		{
			name:        "empty token refresh should fail",
			token:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newToken, err := authService.RefreshToken(ctx, tt.token)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
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
		})
	}
}

func TestAuthService_Interface(t *testing.T) {
	// Verify that MockAuthService implements the AuthService interface
	var _ auth.AuthService = (*MockAuthService)(nil)
}
