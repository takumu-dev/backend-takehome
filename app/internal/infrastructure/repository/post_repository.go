package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"blog-platform/internal/domain/post"
)

// PostRepository implements the post.Repository interface using SQLX
type PostRepository struct {
	db *sqlx.DB
}

// NewPostRepository creates a new PostRepository instance
func NewPostRepository(db *sqlx.DB) *PostRepository {
	return &PostRepository{db: db}
}

// Create inserts a new post into the database
func (r *PostRepository) Create(ctx context.Context, p *post.Post) error {
	if p == nil {
		return post.ErrInvalidPostData
	}

	query := `
		INSERT INTO posts (title, content, author_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query, p.Title, p.Content, p.AuthorID, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create post: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	p.ID = int(id)
	return nil
}

// GetByID retrieves a post by its ID
func (r *PostRepository) GetByID(ctx context.Context, id int) (*post.Post, error) {
	query := `
		SELECT id, title, content, author_id, created_at, updated_at
		FROM posts
		WHERE id = ?
	`

	var p post.Post
	err := r.db.GetContext(ctx, &p, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, post.ErrPostNotFound
		}
		return nil, fmt.Errorf("failed to get post by ID: %w", err)
	}

	return &p, nil
}

// GetByAuthorID retrieves posts by author ID with pagination
func (r *PostRepository) GetByAuthorID(ctx context.Context, authorID int, limit, offset int) ([]*post.Post, error) {
	query := `
		SELECT id, title, content, author_id, created_at, updated_at
		FROM posts
		WHERE author_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	var posts []*post.Post
	err := r.db.SelectContext(ctx, &posts, query, authorID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts by author ID: %w", err)
	}

	return posts, nil
}

// List retrieves all posts with pagination
func (r *PostRepository) List(ctx context.Context, limit, offset int) ([]*post.Post, error) {
	query := `
		SELECT id, title, content, author_id, created_at, updated_at
		FROM posts
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	var posts []*post.Post
	err := r.db.SelectContext(ctx, &posts, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list posts: %w", err)
	}

	return posts, nil
}

// Update updates an existing post in the database
func (r *PostRepository) Update(ctx context.Context, p *post.Post) error {
	if p == nil {
		return post.ErrInvalidPostData
	}

	query := `
		UPDATE posts
		SET title = ?, content = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, p.Title, p.Content, p.UpdatedAt, p.ID)
	if err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return post.ErrPostNotFound
	}

	return nil
}

// Delete removes a post from the database
func (r *PostRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM posts WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return post.ErrPostNotFound
	}

	return nil
}
