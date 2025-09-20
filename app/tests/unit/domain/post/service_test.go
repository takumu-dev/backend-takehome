package post_test

import (
	"context"
	"testing"

	"blog-platform/internal/domain/post"
)

func TestPostService_CreatePost(t *testing.T) {
	tests := []struct {
		name        string
		userID      int
		title       string
		content     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid post creation",
			userID:      1,
			title:       "Test Post",
			content:     "This is a test post content with sufficient length.",
			expectError: false,
		},
		{
			name:        "invalid user ID should fail",
			userID:      0,
			title:       "Test Post",
			content:     "This is a test post content with sufficient length.",
			expectError: true,
			errorMsg:    "author ID must be positive",
		},
		{
			name:        "empty title should fail",
			userID:      1,
			title:       "",
			content:     "This is a test post content with sufficient length.",
			expectError: true,
			errorMsg:    "title cannot be empty",
		},
		{
			name:        "empty content should fail",
			userID:      1,
			title:       "Test Post",
			content:     "",
			expectError: true,
			errorMsg:    "content cannot be empty",
		},
		{
			name:        "content too short should fail",
			userID:      1,
			title:       "Test Post",
			content:     "Short",
			expectError: true,
			errorMsg:    "content must be at least 10 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockPostRepository()
			service := NewMockPostService(repo)
			ctx := context.Background()

			p, err := service.CreatePost(ctx, tt.userID, tt.title, tt.content)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("expected no error, got %v", err)
				return
			}

			if p == nil {
				t.Error("expected post to be created, got nil")
				return
			}

			if p.Title != tt.title {
				t.Errorf("expected title '%s', got '%s'", tt.title, p.Title)
			}

			if p.Content != tt.content {
				t.Errorf("expected content '%s', got '%s'", tt.content, p.Content)
			}

			if p.AuthorID != tt.userID {
				t.Errorf("expected author ID %d, got %d", tt.userID, p.AuthorID)
			}
		})
	}
}

func TestPostService_GetPost(t *testing.T) {
	repo := NewMockPostRepository()
	service := NewMockPostService(repo)
	ctx := context.Background()

	// Create a post first
	createdPost, err := service.CreatePost(ctx, 1, "Test Post", "Test content with sufficient length.")
	if err != nil {
		t.Fatalf("failed to create post: %v", err)
	}

	// Test successful retrieval
	retrievedPost, err := service.GetPost(ctx, createdPost.ID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if retrievedPost == nil {
		t.Error("expected post to be retrieved, got nil")
		return
	}

	if retrievedPost.ID != createdPost.ID {
		t.Errorf("expected post ID %d, got %d", createdPost.ID, retrievedPost.ID)
	}

	// Test retrieval of non-existent post
	_, err = service.GetPost(ctx, 999)
	if err != post.ErrPostNotFound {
		t.Errorf("expected ErrPostNotFound, got %v", err)
	}
}

func TestPostService_UpdatePost(t *testing.T) {
	repo := NewMockPostRepository()
	service := NewMockPostService(repo)
	ctx := context.Background()

	// Create a post first
	createdPost, err := service.CreatePost(ctx, 1, "Original Title", "Original content with sufficient length.")
	if err != nil {
		t.Fatalf("failed to create post: %v", err)
	}

	tests := []struct {
		name        string
		userID      int
		postID      int
		title       string
		content     string
		expectError bool
		errorType   error
	}{
		{
			name:        "valid update by author",
			userID:      1,
			postID:      createdPost.ID,
			title:       "Updated Title",
			content:     "Updated content with sufficient length for validation.",
			expectError: false,
		},
		{
			name:        "unauthorized update should fail",
			userID:      2,
			postID:      createdPost.ID,
			title:       "Updated Title",
			content:     "Updated content with sufficient length for validation.",
			expectError: true,
			errorType:   post.ErrUnauthorized,
		},
		{
			name:        "update non-existent post should fail",
			userID:      1,
			postID:      999,
			title:       "Updated Title",
			content:     "Updated content with sufficient length for validation.",
			expectError: true,
			errorType:   post.ErrPostNotFound,
		},
		{
			name:        "update with empty title should fail",
			userID:      1,
			postID:      createdPost.ID,
			title:       "",
			content:     "Updated content with sufficient length for validation.",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedPost, err := service.UpdatePost(ctx, tt.userID, tt.postID, tt.title, tt.content)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorType != nil && err != tt.errorType {
					t.Errorf("expected error %v, got %v", tt.errorType, err)
				}
				return
			}

			if err != nil {
				t.Errorf("expected no error, got %v", err)
				return
			}

			if updatedPost == nil {
				t.Error("expected updated post, got nil")
				return
			}

			if updatedPost.Title != tt.title {
				t.Errorf("expected title '%s', got '%s'", tt.title, updatedPost.Title)
			}

			if updatedPost.Content != tt.content {
				t.Errorf("expected content '%s', got '%s'", tt.content, updatedPost.Content)
			}
		})
	}
}

func TestPostService_DeletePost(t *testing.T) {
	repo := NewMockPostRepository()
	service := NewMockPostService(repo)
	ctx := context.Background()

	// Create a post first
	createdPost, err := service.CreatePost(ctx, 1, "Test Post", "Test content with sufficient length.")
	if err != nil {
		t.Fatalf("failed to create post: %v", err)
	}

	tests := []struct {
		name        string
		userID      int
		postID      int
		expectError bool
		errorType   error
	}{
		{
			name:        "valid deletion by author",
			userID:      1,
			postID:      createdPost.ID,
			expectError: false,
		},
		{
			name:        "unauthorized deletion should fail",
			userID:      2,
			postID:      createdPost.ID,
			expectError: true,
			errorType:   post.ErrUnauthorized,
		},
		{
			name:        "delete non-existent post should fail",
			userID:      1,
			postID:      999,
			expectError: true,
			errorType:   post.ErrPostNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the correct post ID for each test case
			postID := tt.postID
			if tt.name != "delete non-existent post should fail" {
				// Create a fresh post for each test
				testPost, _ := service.CreatePost(ctx, 1, "Test Post", "Test content with sufficient length.")
				postID = testPost.ID
			}

			err := service.DeletePost(ctx, tt.userID, postID)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorType != nil && err != tt.errorType {
					t.Errorf("expected error %v, got %v", tt.errorType, err)
				}
				return
			}

			if err != nil {
				t.Errorf("expected no error, got %v", err)
				return
			}

			// Verify the post is deleted (only for successful deletion)
			if !tt.expectError {
				_, err = service.GetPost(ctx, postID)
				if err != post.ErrPostNotFound {
					t.Error("expected post to be deleted")
				}
			}
		})
	}
}

func TestPostService_GetPostsByAuthor(t *testing.T) {
	repo := NewMockPostRepository()
	service := NewMockPostService(repo)
	ctx := context.Background()

	// Create posts by different authors
	service.CreatePost(ctx, 1, "Post 1", "Content for post 1 with sufficient length.")
	service.CreatePost(ctx, 1, "Post 2", "Content for post 2 with sufficient length.")
	service.CreatePost(ctx, 2, "Post 3", "Content for post 3 with sufficient length.")

	// Test getting posts by author ID 1
	posts, err := service.GetPostsByAuthor(ctx, 1, 10, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(posts) != 2 {
		t.Errorf("expected 2 posts, got %d", len(posts))
	}

	// Test getting posts by author ID 2
	posts, err = service.GetPostsByAuthor(ctx, 2, 10, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(posts) != 1 {
		t.Errorf("expected 1 post, got %d", len(posts))
	}

	// Test pagination
	posts, err = service.GetPostsByAuthor(ctx, 1, 1, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(posts) != 1 {
		t.Errorf("expected 1 post with limit, got %d", len(posts))
	}
}

func TestPostService_ListPosts(t *testing.T) {
	repo := NewMockPostRepository()
	service := NewMockPostService(repo)
	ctx := context.Background()

	// Create multiple posts
	for i := 1; i <= 5; i++ {
		service.CreatePost(ctx, i, "Post Title", "Post content with sufficient length for validation.")
	}

	// Test listing all posts
	posts, err := service.ListPosts(ctx, 10, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(posts) != 5 {
		t.Errorf("expected 5 posts, got %d", len(posts))
	}

	// Test pagination with limit
	posts, err = service.ListPosts(ctx, 3, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(posts) != 3 {
		t.Errorf("expected 3 posts with limit, got %d", len(posts))
	}

	// Test pagination with offset
	posts, err = service.ListPosts(ctx, 10, 2)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(posts) != 3 {
		t.Errorf("expected 3 posts with offset, got %d", len(posts))
	}
}

// MockPostService implements the PostService interface for testing
type MockPostService struct {
	repo post.Repository
}

func NewMockPostService(repo post.Repository) *MockPostService {
	return &MockPostService{repo: repo}
}

func (s *MockPostService) CreatePost(ctx context.Context, userID int, title, content string) (*post.Post, error) {
	p, err := post.NewPost(title, content, userID)
	if err != nil {
		return nil, err
	}

	err = s.repo.Create(ctx, p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (s *MockPostService) GetPost(ctx context.Context, id int) (*post.Post, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *MockPostService) GetPostsByAuthor(ctx context.Context, authorID int, limit, offset int) ([]*post.Post, error) {
	// Validate pagination parameters
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetByAuthorID(ctx, authorID, limit, offset)
}

func (s *MockPostService) ListPosts(ctx context.Context, limit, offset int) ([]*post.Post, error) {
	// Validate pagination parameters
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(ctx, limit, offset)
}

func (s *MockPostService) UpdatePost(ctx context.Context, userID, postID int, title, content string) (*post.Post, error) {
	// Get the existing post
	existingPost, err := s.repo.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	// Check authorization
	if !existingPost.IsAuthor(userID) {
		return nil, post.ErrUnauthorized
	}

	// Update the post
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

func (s *MockPostService) DeletePost(ctx context.Context, userID, postID int) error {
	// Get the existing post
	existingPost, err := s.repo.GetByID(ctx, postID)
	if err != nil {
		return err
	}

	// Check authorization
	if !existingPost.IsAuthor(userID) {
		return post.ErrUnauthorized
	}

	// Delete the post
	return s.repo.Delete(ctx, postID)
}
