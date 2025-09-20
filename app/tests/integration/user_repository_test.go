package integration

import (
	"context"
	"testing"
	"time"

	"blog-platform/internal/domain/user"
	"blog-platform/internal/infrastructure/config"
	"blog-platform/internal/infrastructure/database"
	"blog-platform/internal/infrastructure/repository"
)

func setupTestDB(t *testing.T) *database.Database {
	cfg := config.Load()
	db, err := database.NewDatabase(cfg)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}
	return db
}

func cleanupUsers(t *testing.T, db *database.Database) {
	_, err := db.Exec("DELETE FROM users WHERE email LIKE '%test%'")
	if err != nil {
		t.Logf("failed to cleanup test users: %v", err)
	}
}

func TestUserRepository_Integration_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupUsers(t, db)

	repo := repository.NewUserRepository(db.DB)
	ctx := context.Background()

	u, err := user.NewUser("Test User", "test@example.com", "password123")
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

func TestUserRepository_Integration_DuplicateEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupUsers(t, db)

	repo := repository.NewUserRepository(db.DB)
	ctx := context.Background()

	// Create first user
	u1, _ := user.NewUser("Test User 1", "duplicate@example.com", "password123")
	err := repo.Create(ctx, u1)
	if err != nil {
		t.Fatalf("failed to create first user: %v", err)
	}

	// Try to create second user with same email
	u2, _ := user.NewUser("Test User 2", "duplicate@example.com", "password456")
	err = repo.Create(ctx, u2)
	if err != user.ErrUserExists {
		t.Errorf("expected ErrUserExists, got %v", err)
	}
}

func TestUserRepository_Integration_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupUsers(t, db)

	repo := repository.NewUserRepository(db.DB)
	ctx := context.Background()

	// Create and store a user
	u, _ := user.NewUser("Test User", "getbyemail@example.com", "password123")
	repo.Create(ctx, u)

	// Test getting existing user by email
	retrieved, err := repo.GetByEmail(ctx, "getbyemail@example.com")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if retrieved.Email != u.Email {
		t.Errorf("expected email %s, got %s", u.Email, retrieved.Email)
	}

	if retrieved.ID != u.ID {
		t.Errorf("expected ID %d, got %d", u.ID, retrieved.ID)
	}

	// Test getting non-existent user
	_, err = repo.GetByEmail(ctx, "nonexistent@example.com")
	if err != user.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepository_Integration_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupUsers(t, db)

	repo := repository.NewUserRepository(db.DB)
	ctx := context.Background()

	// Create and store a user
	u, _ := user.NewUser("Original Name", "update@example.com", "password123")
	repo.Create(ctx, u)

	// Update the user
	err := u.UpdateProfile("Updated Name", "updated@example.com")
	if err != nil {
		t.Fatalf("failed to update profile: %v", err)
	}

	err = repo.Update(ctx, u)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify the update
	updated, err := repo.GetByID(ctx, u.ID)
	if err != nil {
		t.Errorf("failed to retrieve updated user: %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got %s", updated.Name)
	}

	if updated.Email != "updated@example.com" {
		t.Errorf("expected email 'updated@example.com', got %s", updated.Email)
	}

	// Verify UpdatedAt was changed
	if !updated.UpdatedAt.After(updated.CreatedAt) {
		t.Error("expected UpdatedAt to be after CreatedAt")
	}
}

func TestUserRepository_Integration_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupUsers(t, db)

	repo := repository.NewUserRepository(db.DB)
	ctx := context.Background()

	// Create and store a user
	u, _ := user.NewUser("Delete Me", "delete@example.com", "password123")
	repo.Create(ctx, u)

	// Verify user exists
	_, err := repo.GetByID(ctx, u.ID)
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

func TestUserRepository_Integration_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupUsers(t, db)

	repo := repository.NewUserRepository(db.DB)
	ctx := context.Background()

	// Create multiple test users
	testUsers := make([]*user.User, 5)
	for i := 0; i < 5; i++ {
		u, _ := user.NewUser("List Test User", "listtest"+string(rune('a'+i))+"@example.com", "password123")
		repo.Create(ctx, u)
		testUsers[i] = u
		time.Sleep(1 * time.Millisecond) // Ensure different created_at times
	}

	// Test listing all users
	users, err := repo.List(ctx, 10, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(users) < 5 {
		t.Errorf("expected at least 5 users, got %d", len(users))
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
	users, err = repo.List(ctx, 2, 2)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}
