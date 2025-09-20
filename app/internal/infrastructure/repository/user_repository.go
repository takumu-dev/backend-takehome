package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"blog-platform/internal/domain/user"
)

// UserRepository implements the user.Repository interface using SQLX
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user into the database
func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	query := `
		INSERT INTO users (name, email, password_hash, created_at, updated_at)
		VALUES (:name, :email, :password_hash, :created_at, :updated_at)
	`
	
	result, err := r.db.NamedExecContext(ctx, query, u)
	if err != nil {
		// Check for duplicate email constraint
		if isDuplicateKeyError(err) {
			return user.ErrUserExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	u.ID = int(id)
	return nil
}

// GetByID retrieves a user by their ID
func (r *UserRepository) GetByID(ctx context.Context, id int) (*user.User, error) {
	query := `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = ?
	`
	
	var u user.User
	err := r.db.GetContext(ctx, &u, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &u, nil
}

// GetByEmail retrieves a user by their email address
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = ?
	`
	
	var u user.User
	err := r.db.GetContext(ctx, &u, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &u, nil
}

// Update modifies an existing user in the database
func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	query := `
		UPDATE users
		SET name = :name, email = :email, password_hash = :password_hash, updated_at = :updated_at
		WHERE id = :id
	`
	
	result, err := r.db.NamedExecContext(ctx, query, u)
	if err != nil {
		// Check for duplicate email constraint
		if isDuplicateKeyError(err) {
			return user.ErrUserExists
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return user.ErrUserNotFound
	}

	return nil
}

// Delete removes a user from the database
func (r *UserRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = ?`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return user.ErrUserNotFound
	}

	return nil
}

// List retrieves a paginated list of users
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*user.User, error) {
	query := `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	
	var users []user.User
	err := r.db.SelectContext(ctx, &users, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Convert to slice of pointers
	result := make([]*user.User, len(users))
	for i := range users {
		result[i] = &users[i]
	}

	return result, nil
}

// isDuplicateKeyError checks if the error is a duplicate key constraint violation
func isDuplicateKeyError(err error) bool {
	// MySQL error code 1062 is for duplicate entry
	// This is a simplified check - in production, you might want to use a more robust method
	return err != nil && (
		// MySQL duplicate entry error
		containsString(err.Error(), "Duplicate entry") ||
		containsString(err.Error(), "duplicate key") ||
		containsString(err.Error(), "UNIQUE constraint failed"))
}

// containsString checks if a string contains a substring (case-insensitive)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    (len(s) > len(substr) && 
		     (s[:len(substr)] == substr || 
		      s[len(s)-len(substr):] == substr || 
		      containsSubstring(s, substr))))
}

// containsSubstring is a helper function for substring checking
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
