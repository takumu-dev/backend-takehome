package service_test

import (
	"context"
	"testing"

	"blog-platform/internal/application/service"
	"blog-platform/internal/domain/user"
)

// MockUserRepository implements the UserRepository interface for testing
type MockUserRepository struct {
	users  map[int]*user.User
	nextID int
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:  make(map[int]*user.User),
		nextID: 1,
	}
}

func (m *MockUserRepository) Create(ctx context.Context, u *user.User) error {
	u.ID = m.nextID
	m.nextID++
	m.users[u.ID] = u
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*user.User, error) {
	if u, exists := m.users[id]; exists {
		return u, nil
	}
	return nil, user.ErrUserNotFound
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, user.ErrUserNotFound
}

func (m *MockUserRepository) Update(ctx context.Context, u *user.User) error {
	if _, exists := m.users[u.ID]; !exists {
		return user.ErrUserNotFound
	}
	m.users[u.ID] = u
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id int) error {
	if _, exists := m.users[id]; !exists {
		return user.ErrUserNotFound
	}
	delete(m.users, id)
	return nil
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*user.User, error) {
	var users []*user.User
	count := 0
	skipped := 0
	
	for _, u := range m.users {
		if skipped < offset {
			skipped++
			continue
		}
		if count >= limit {
			break
		}
		users = append(users, u)
		count++
	}
	
	return users, nil
}

func TestUserService_Implementation(t *testing.T) {
	// Test that our concrete service implements the interface
	repo := NewMockUserRepository()
	var _ user.Service = service.NewUserService(repo, NewMockLogger())
}

func TestUserService_Register_Integration(t *testing.T) {
	repo := NewMockUserRepository()
	userService := service.NewUserService(repo, NewMockLogger())
	ctx := context.Background()

	// Test successful registration
	u, err := userService.Register(ctx, "John Doe", "john@example.com", "password123")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if u == nil {
		t.Fatal("expected user to be returned")
	}

	if u.Name != "John Doe" {
		t.Errorf("expected name 'John Doe', got %s", u.Name)
	}

	if u.Email != "john@example.com" {
		t.Errorf("expected email 'john@example.com', got %s", u.Email)
	}

	if u.ID == 0 {
		t.Error("expected user ID to be set")
	}

	// Verify password is hashed
	if u.PasswordHash == "password123" {
		t.Error("password should be hashed, not stored in plain text")
	}

	// Test duplicate email registration
	_, err = userService.Register(ctx, "Jane Doe", "john@example.com", "password456")
	if err != user.ErrUserExists {
		t.Errorf("expected ErrUserExists, got %v", err)
	}
}

func TestUserService_Login_Integration(t *testing.T) {
	repo := NewMockUserRepository()
	userService := service.NewUserService(repo, NewMockLogger())
	ctx := context.Background()

	// Register a user first
	registeredUser, err := userService.Register(ctx, "John Doe", "login@example.com", "password123")
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	// Test successful login
	loginUser, err := userService.Login(ctx, "login@example.com", "password123")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if loginUser.ID != registeredUser.ID {
		t.Errorf("expected user ID %d, got %d", registeredUser.ID, loginUser.ID)
	}

	// Test login with wrong password
	_, err = userService.Login(ctx, "login@example.com", "wrongpassword")
	if err != user.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}

	// Test login with non-existent email
	_, err = userService.Login(ctx, "nonexistent@example.com", "password123")
	if err != user.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestUserService_UpdateProfile_Integration(t *testing.T) {
	repo := NewMockUserRepository()
	userService := service.NewUserService(repo, NewMockLogger())
	ctx := context.Background()

	// Register a user first
	registeredUser, err := userService.Register(ctx, "John Doe", "update@example.com", "password123")
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	// Test successful profile update
	updatedUser, err := userService.UpdateProfile(ctx, registeredUser.ID, "Jane Doe", "jane@example.com")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if updatedUser.Name != "Jane Doe" {
		t.Errorf("expected name 'Jane Doe', got %s", updatedUser.Name)
	}

	if updatedUser.Email != "jane@example.com" {
		t.Errorf("expected email 'jane@example.com', got %s", updatedUser.Email)
	}

	// Verify the update persisted
	retrievedUser, err := userService.GetByID(ctx, registeredUser.ID)
	if err != nil {
		t.Errorf("failed to retrieve updated user: %v", err)
	}

	if retrievedUser.Name != "Jane Doe" {
		t.Errorf("expected persisted name 'Jane Doe', got %s", retrievedUser.Name)
	}
}

func TestUserService_UpdatePassword_Integration(t *testing.T) {
	repo := NewMockUserRepository()
	userService := service.NewUserService(repo, NewMockLogger())
	ctx := context.Background()

	// Register a user first
	registeredUser, err := userService.Register(ctx, "John Doe", "password@example.com", "oldpassword123")
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	// Test successful password update
	err = userService.UpdatePassword(ctx, registeredUser.ID, "oldpassword123", "newpassword123")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify old password no longer works
	_, err = userService.Login(ctx, "password@example.com", "oldpassword123")
	if err != user.ErrInvalidCredentials {
		t.Errorf("expected old password to be invalid, got %v", err)
	}

	// Verify new password works
	_, err = userService.Login(ctx, "password@example.com", "newpassword123")
	if err != nil {
		t.Errorf("expected new password to work, got %v", err)
	}

	// Test password update with wrong current password
	err = userService.UpdatePassword(ctx, registeredUser.ID, "wrongpassword", "anotherpassword")
	if err != user.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestUserService_List_Integration(t *testing.T) {
	repo := NewMockUserRepository()
	userService := service.NewUserService(repo, NewMockLogger())
	ctx := context.Background()

	// Register multiple users
	for i := 0; i < 5; i++ {
		name := "User " + string(rune('A'+i))
		email := "user" + string(rune('a'+i)) + "@example.com"
		_, err := userService.Register(ctx, name, email, "password123")
		if err != nil {
			t.Fatalf("failed to register user %d: %v", i, err)
		}
	}

	// Test listing users
	users, err := userService.List(ctx, 10, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(users) != 5 {
		t.Errorf("expected 5 users, got %d", len(users))
	}

	// Test pagination
	users, err = userService.List(ctx, 3, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(users) != 3 {
		t.Errorf("expected 3 users, got %d", len(users))
	}

	// Test limit validation (should cap at 100)
	users, err = userService.List(ctx, 200, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Should still return all users since we only have 5
	if len(users) != 5 {
		t.Errorf("expected 5 users, got %d", len(users))
	}
}
