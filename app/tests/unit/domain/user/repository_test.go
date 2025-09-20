package user_test

import (
	"context"
	"testing"

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

func TestUserRepository_Create(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	u, err := user.NewUser("John Doe", "john@example.com", "password123")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	err = repo.Create(ctx, u)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if u.ID == 0 {
		t.Error("expected user ID to be set after creation")
	}

	// Verify user was stored
	stored, err := repo.GetByID(ctx, u.ID)
	if err != nil {
		t.Errorf("failed to retrieve created user: %v", err)
	}

	if stored.Name != u.Name {
		t.Errorf("expected name %s, got %s", u.Name, stored.Name)
	}

	if stored.Email != u.Email {
		t.Errorf("expected email %s, got %s", u.Email, stored.Email)
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	// Test getting non-existent user
	_, err := repo.GetByID(ctx, 999)
	if err != user.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}

	// Create and store a user
	u, _ := user.NewUser("John Doe", "john@example.com", "password123")
	repo.Create(ctx, u)

	// Test getting existing user
	retrieved, err := repo.GetByID(ctx, u.ID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if retrieved.ID != u.ID {
		t.Errorf("expected ID %d, got %d", u.ID, retrieved.ID)
	}

	if retrieved.Email != u.Email {
		t.Errorf("expected email %s, got %s", u.Email, retrieved.Email)
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	// Test getting non-existent user
	_, err := repo.GetByEmail(ctx, "nonexistent@example.com")
	if err != user.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}

	// Create and store a user
	u, _ := user.NewUser("John Doe", "john@example.com", "password123")
	repo.Create(ctx, u)

	// Test getting existing user by email
	retrieved, err := repo.GetByEmail(ctx, "john@example.com")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if retrieved.Email != u.Email {
		t.Errorf("expected email %s, got %s", u.Email, retrieved.Email)
	}

	if retrieved.ID != u.ID {
		t.Errorf("expected ID %d, got %d", u.ID, retrieved.ID)
	}
}

func TestUserRepository_Update(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	// Test updating non-existent user
	u, _ := user.NewUser("John Doe", "john@example.com", "password123")
	u.ID = 999
	err := repo.Update(ctx, u)
	if err != user.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}

	// Create and store a user
	originalUser, _ := user.NewUser("John Doe", "john@example.com", "password123")
	repo.Create(ctx, originalUser)

	// Update the user
	err = originalUser.UpdateProfile("Jane Doe", "jane@example.com")
	if err != nil {
		t.Fatalf("failed to update profile: %v", err)
	}

	err = repo.Update(ctx, originalUser)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify the update
	updated, err := repo.GetByID(ctx, originalUser.ID)
	if err != nil {
		t.Errorf("failed to retrieve updated user: %v", err)
	}

	if updated.Name != "Jane Doe" {
		t.Errorf("expected name 'Jane Doe', got %s", updated.Name)
	}

	if updated.Email != "jane@example.com" {
		t.Errorf("expected email 'jane@example.com', got %s", updated.Email)
	}
}

func TestUserRepository_Delete(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	// Test deleting non-existent user
	err := repo.Delete(ctx, 999)
	if err != user.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}

	// Create and store a user
	u, _ := user.NewUser("John Doe", "john@example.com", "password123")
	repo.Create(ctx, u)

	// Verify user exists
	_, err = repo.GetByID(ctx, u.ID)
	if err != nil {
		t.Errorf("user should exist before deletion")
	}

	// Delete the user
	err = repo.Delete(ctx, u.ID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify user no longer exists
	_, err = repo.GetByID(ctx, u.ID)
	if err != user.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound after deletion, got %v", err)
	}
}

func TestUserRepository_List(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	// Test empty repository
	users, err := repo.List(ctx, 10, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}

	// Create multiple users
	for i := 0; i < 5; i++ {
		u, _ := user.NewUser("User "+string(rune('A'+i)), "user"+string(rune('a'+i))+"@example.com", "password123")
		repo.Create(ctx, u)
	}

	// Test listing all users
	users, err = repo.List(ctx, 10, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(users) != 5 {
		t.Errorf("expected 5 users, got %d", len(users))
	}

	// Test pagination - limit
	users, err = repo.List(ctx, 3, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(users) != 3 {
		t.Errorf("expected 3 users, got %d", len(users))
	}

	// Test pagination - offset
	users, err = repo.List(ctx, 10, 2)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(users) != 3 {
		t.Errorf("expected 3 users (5 total - 2 offset), got %d", len(users))
	}
}

func TestUserRepository_Interface(t *testing.T) {
	// This test ensures our mock implements the UserRepository interface
	var _ user.Repository = (*MockUserRepository)(nil)
}
