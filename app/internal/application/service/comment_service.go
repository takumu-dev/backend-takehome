package service

import (
	"context"
	"errors"

	"blog-platform/internal/domain/comment"
)

// CommentService implements the comment.Service interface
type CommentService struct {
	repo comment.Repository
}

// NewCommentService creates a new comment service
func NewCommentService(repo comment.Repository) *CommentService {
	return &CommentService{
		repo: repo,
	}
}

// AddComment creates a new comment with validation
func (s *CommentService) AddComment(ctx context.Context, postID int, authorName, content string) (*comment.Comment, error) {
	// Create comment with validation
	c, err := comment.NewComment(postID, authorName, content)
	if err != nil {
		return nil, err
	}

	// Save to repository
	err = s.repo.Create(ctx, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// GetComment retrieves a comment by ID
func (s *CommentService) GetComment(ctx context.Context, id int) (*comment.Comment, error) {
	if id <= 0 {
		return nil, errors.New("comment ID must be positive")
	}

	return s.repo.GetByID(ctx, id)
}

// GetCommentsByPost retrieves comments for a post with pagination validation
func (s *CommentService) GetCommentsByPost(ctx context.Context, postID int, limit, offset int) ([]*comment.Comment, error) {
	// Validate post ID
	if postID <= 0 {
		return nil, errors.New("post ID must be positive")
	}

	// Validate pagination parameters
	if limit <= 0 || limit > 100 {
		return nil, errors.New("limit must be between 1 and 100")
	}
	if offset < 0 {
		return nil, errors.New("offset must be non-negative")
	}

	return s.repo.GetByPostID(ctx, postID, limit, offset)
}

// UpdateComment updates a comment's content with authorization check
func (s *CommentService) UpdateComment(ctx context.Context, id int, authorName, content string) (*comment.Comment, error) {
	// Validate ID
	if id <= 0 {
		return nil, errors.New("comment ID must be positive")
	}

	// Get existing comment
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check authorization - only the author can update their comment
	if !c.IsAuthor(authorName) {
		return nil, errors.New("unauthorized: only the author can update this comment")
	}

	// Update content with validation
	err = c.Update(content)
	if err != nil {
		return nil, err
	}

	// Save to repository
	err = s.repo.Update(ctx, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// DeleteComment deletes a comment with authorization check
func (s *CommentService) DeleteComment(ctx context.Context, id int, authorName string) error {
	// Validate ID
	if id <= 0 {
		return errors.New("comment ID must be positive")
	}

	// Get existing comment
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check authorization - only the author can delete their comment
	if !c.IsAuthor(authorName) {
		return errors.New("unauthorized: only the author can delete this comment")
	}

	return s.repo.Delete(ctx, id)
}
