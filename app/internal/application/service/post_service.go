package service

import (
	"context"

	"blog-platform/internal/domain/post"
)

// PostService implements the post.Service interface
type PostService struct {
	repo   post.Repository
	logger Logger
}

// NewPostService creates a new PostService instance
func NewPostService(repo post.Repository, logger Logger) *PostService {
	return &PostService{
		repo:   repo,
		logger: logger,
	}
}

// CreatePost creates a new post with validation
func (s *PostService) CreatePost(ctx context.Context, userID int, title, content string) (*post.Post, error) {
	s.logger.Info(ctx, "creating post", "userID", userID, "title", title)
	
	// Create post entity with validation
	p, err := post.NewPost(title, content, userID)
	if err != nil {
		s.logger.Error(ctx, "failed to create post entity", "userID", userID, "error", err.Error())
		return nil, err
	}

	// Save to repository
	err = s.repo.Create(ctx, p)
	if err != nil {
		s.logger.Error(ctx, "failed to save post to repository", "userID", userID, "postID", p.ID, "error", err.Error())
		return nil, err
	}

	s.logger.Info(ctx, "post created successfully", "userID", userID, "postID", p.ID, "title", title)
	return p, nil
}

// GetPost retrieves a post by ID
func (s *PostService) GetPost(ctx context.Context, id int) (*post.Post, error) {
	s.logger.Debug(ctx, "retrieving post", "postID", id)
	
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "failed to retrieve post", "postID", id, "error", err.Error())
		return nil, err
	}
	
	s.logger.Debug(ctx, "post retrieved successfully", "postID", id)
	return post, nil
}

// GetPostsByAuthor retrieves posts by author ID with pagination
func (s *PostService) GetPostsByAuthor(ctx context.Context, authorID int, limit, offset int) ([]*post.Post, error) {
	// Validate and normalize pagination parameters
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetByAuthorID(ctx, authorID, limit, offset)
}

// ListPosts retrieves all posts with pagination
func (s *PostService) ListPosts(ctx context.Context, limit, offset int) ([]*post.Post, error) {
	// Validate and normalize pagination parameters
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(ctx, limit, offset)
}

// UpdatePost updates a post with authorization checks
func (s *PostService) UpdatePost(ctx context.Context, userID, postID int, title, content string) (*post.Post, error) {
	s.logger.Info(ctx, "updating post", "userID", userID, "postID", postID)
	
	// Get the existing post
	existingPost, err := s.repo.GetByID(ctx, postID)
	if err != nil {
		s.logger.Error(ctx, "failed to retrieve post for update", "postID", postID, "error", err.Error())
		return nil, err
	}

	// Check authorization - only the author can update the post
	if !existingPost.IsAuthor(userID) {
		s.logger.Warn(ctx, "unauthorized post update attempt", "userID", userID, "postID", postID, "authorID", existingPost.AuthorID)
		return nil, post.ErrUnauthorized
	}

	// Update the post with validation
	err = existingPost.Update(title, content)
	if err != nil {
		s.logger.Error(ctx, "failed to update post entity", "postID", postID, "error", err.Error())
		return nil, err
	}

	// Save the updated post
	err = s.repo.Update(ctx, existingPost)
	if err != nil {
		s.logger.Error(ctx, "failed to save updated post", "postID", postID, "error", err.Error())
		return nil, err
	}

	s.logger.Info(ctx, "post updated successfully", "userID", userID, "postID", postID)
	return existingPost, nil
}

// DeletePost deletes a post with authorization checks
func (s *PostService) DeletePost(ctx context.Context, userID, postID int) error {
	s.logger.Info(ctx, "deleting post", "userID", userID, "postID", postID)
	
	// Get the existing post
	existingPost, err := s.repo.GetByID(ctx, postID)
	if err != nil {
		s.logger.Error(ctx, "failed to retrieve post for deletion", "postID", postID, "error", err.Error())
		return err
	}

	// Check authorization - only the author can delete the post
	if !existingPost.IsAuthor(userID) {
		s.logger.Warn(ctx, "unauthorized post deletion attempt", "userID", userID, "postID", postID, "authorID", existingPost.AuthorID)
		return post.ErrUnauthorized
	}

	// Delete the post
	err = s.repo.Delete(ctx, postID)
	if err != nil {
		s.logger.Error(ctx, "failed to delete post", "postID", postID, "error", err.Error())
		return err
	}
	
	s.logger.Info(ctx, "post deleted successfully", "userID", userID, "postID", postID)
	return nil
}
