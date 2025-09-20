package service

import (
	"context"

	"blog-platform/internal/domain/post"
)

// PostService implements the post.Service interface
type PostService struct {
	repo post.Repository
}

// NewPostService creates a new PostService instance
func NewPostService(repo post.Repository) *PostService {
	return &PostService{repo: repo}
}

// CreatePost creates a new post with validation
func (s *PostService) CreatePost(ctx context.Context, userID int, title, content string) (*post.Post, error) {
	// Create post entity with validation
	p, err := post.NewPost(title, content, userID)
	if err != nil {
		return nil, err
	}

	// Save to repository
	err = s.repo.Create(ctx, p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// GetPost retrieves a post by ID
func (s *PostService) GetPost(ctx context.Context, id int) (*post.Post, error) {
	return s.repo.GetByID(ctx, id)
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
	// Get the existing post
	existingPost, err := s.repo.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	// Check authorization - only the author can update the post
	if !existingPost.IsAuthor(userID) {
		return nil, post.ErrUnauthorized
	}

	// Update the post with validation
	err = existingPost.Update(title, content)
	if err != nil {
		return nil, err
	}

	// Save the updated post
	err = s.repo.Update(ctx, existingPost)
	if err != nil {
		return nil, err
	}

	return existingPost, nil
}

// DeletePost deletes a post with authorization checks
func (s *PostService) DeletePost(ctx context.Context, userID, postID int) error {
	// Get the existing post
	existingPost, err := s.repo.GetByID(ctx, postID)
	if err != nil {
		return err
	}

	// Check authorization - only the author can delete the post
	if !existingPost.IsAuthor(userID) {
		return post.ErrUnauthorized
	}

	// Delete the post
	return s.repo.Delete(ctx, postID)
}
