package service

import (
	"context"
	"errors"

	"blog-platform/internal/application/logging"
	"blog-platform/internal/domain/comment"
)

// CommentService implements the comment.Service interface
type CommentService struct {
	repo   comment.Repository
	logger logging.Logger
}

// NewCommentService creates a new comment service
func NewCommentService(repo comment.Repository, logger logging.Logger) *CommentService {
	return &CommentService{
		repo:   repo,
		logger: logger,
	}
}

// AddComment creates a new comment with validation
func (s *CommentService) AddComment(ctx context.Context, postID int, authorName, content string) (*comment.Comment, error) {
	s.logger.Info(ctx, "adding comment", "postID", postID, "authorName", authorName)
	
	// Create comment with validation
	c, err := comment.NewComment(postID, authorName, content)
	if err != nil {
		s.logger.Error(ctx, "failed to create comment entity", "postID", postID, "authorName", authorName, "error", err.Error())
		return nil, err
	}

	// Save to repository
	err = s.repo.Create(ctx, c)
	if err != nil {
		s.logger.Error(ctx, "failed to save comment to repository", "postID", postID, "commentID", c.ID, "error", err.Error())
		return nil, err
	}

	s.logger.Info(ctx, "comment added successfully", "postID", postID, "commentID", c.ID, "authorName", authorName)
	return c, nil
}

// GetComment retrieves a comment by ID
func (s *CommentService) GetComment(ctx context.Context, id int) (*comment.Comment, error) {
	s.logger.Debug(ctx, "retrieving comment", "commentID", id)
	
	if id <= 0 {
		s.logger.Error(ctx, "invalid comment ID", "commentID", id)
		return nil, errors.New("comment ID must be positive")
	}

	comment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "failed to retrieve comment", "commentID", id, "error", err.Error())
		return nil, err
	}
	
	s.logger.Debug(ctx, "comment retrieved successfully", "commentID", id)
	return comment, nil
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
	s.logger.Info(ctx, "updating comment", "commentID", id, "authorName", authorName)
	
	// Validate ID
	if id <= 0 {
		s.logger.Error(ctx, "invalid comment ID for update", "commentID", id)
		return nil, errors.New("comment ID must be positive")
	}

	// Get existing comment
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "failed to retrieve comment for update", "commentID", id, "error", err.Error())
		return nil, err
	}

	// Check authorization - only the author can update their comment
	if !c.IsAuthor(authorName) {
		s.logger.Warn(ctx, "unauthorized comment update attempt", "commentID", id, "requestedBy", authorName, "actualAuthor", c.AuthorName)
		return nil, errors.New("unauthorized: only the author can update this comment")
	}

	// Update content with validation
	err = c.Update(content)
	if err != nil {
		s.logger.Error(ctx, "failed to update comment entity", "commentID", id, "error", err.Error())
		return nil, err
	}

	// Save to repository
	err = s.repo.Update(ctx, c)
	if err != nil {
		s.logger.Error(ctx, "failed to save updated comment", "commentID", id, "error", err.Error())
		return nil, err
	}

	s.logger.Info(ctx, "comment updated successfully", "commentID", id, "authorName", authorName)
	return c, nil
}

// DeleteComment deletes a comment with authorization check
func (s *CommentService) DeleteComment(ctx context.Context, id int, authorName string) error {
	s.logger.Info(ctx, "deleting comment", "commentID", id, "authorName", authorName)
	
	// Validate ID
	if id <= 0 {
		s.logger.Error(ctx, "invalid comment ID for deletion", "commentID", id)
		return errors.New("comment ID must be positive")
	}

	// Get existing comment
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "failed to retrieve comment for deletion", "commentID", id, "error", err.Error())
		return err
	}

	// Check authorization - only the author can delete their comment
	if !c.IsAuthor(authorName) {
		s.logger.Warn(ctx, "unauthorized comment deletion attempt", "commentID", id, "requestedBy", authorName, "actualAuthor", c.AuthorName)
		return errors.New("unauthorized: only the author can delete this comment")
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "failed to delete comment", "commentID", id, "error", err.Error())
		return err
	}
	
	s.logger.Info(ctx, "comment deleted successfully", "commentID", id, "authorName", authorName)
	return nil
}
