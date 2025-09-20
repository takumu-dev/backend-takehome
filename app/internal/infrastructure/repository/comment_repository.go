package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"blog-platform/internal/domain/comment"
)

// CommentRepository implements the comment.Repository interface using SQLX
type CommentRepository struct {
	db *sqlx.DB
}

// NewCommentRepository creates a new comment repository
func NewCommentRepository(db *sqlx.DB) *CommentRepository {
	return &CommentRepository{
		db: db,
	}
}

// Create inserts a new comment into the database
func (r *CommentRepository) Create(ctx context.Context, c *comment.Comment) error {
	query := `
		INSERT INTO comments (post_id, author_name, content, created_at)
		VALUES (?, ?, ?, ?)
	`
	
	result, err := r.db.ExecContext(ctx, query, c.PostID, c.AuthorName, c.Content, c.CreatedAt)
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	c.ID = int(id)
	return nil
}

// GetByID retrieves a comment by its ID
func (r *CommentRepository) GetByID(ctx context.Context, id int) (*comment.Comment, error) {
	query := `
		SELECT id, post_id, author_name, content, created_at
		FROM comments
		WHERE id = ?
	`
	
	var c comment.Comment
	err := r.db.GetContext(ctx, &c, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, comment.ErrCommentNotFound
		}
		return nil, err
	}
	
	return &c, nil
}

// GetByPostID retrieves comments for a specific post with pagination
func (r *CommentRepository) GetByPostID(ctx context.Context, postID int, limit, offset int) ([]*comment.Comment, error) {
	query := `
		SELECT id, post_id, author_name, content, created_at
		FROM comments
		WHERE post_id = ?
		ORDER BY created_at ASC
		LIMIT ? OFFSET ?
	`
	
	var comments []*comment.Comment
	err := r.db.SelectContext(ctx, &comments, query, postID, limit, offset)
	if err != nil {
		return nil, err
	}
	
	return comments, nil
}

// Update modifies an existing comment in the database
func (r *CommentRepository) Update(ctx context.Context, c *comment.Comment) error {
	query := `
		UPDATE comments
		SET content = ?
		WHERE id = ?
	`
	
	result, err := r.db.ExecContext(ctx, query, c.Content, c.ID)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return comment.ErrCommentNotFound
	}
	
	return nil
}

// Delete removes a comment from the database
func (r *CommentRepository) Delete(ctx context.Context, id int) error {
	query := `
		DELETE FROM comments
		WHERE id = ?
	`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return comment.ErrCommentNotFound
	}
	
	return nil
}
