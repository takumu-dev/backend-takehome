package comment_test

import (
	"context"
	"errors"
	"testing"

	"blog-platform/internal/domain/comment"
)

// MockCommentService implements the comment.Service interface for testing
type MockCommentService struct {
	repo comment.Repository
}

// NewMockCommentService creates a new mock service
func NewMockCommentService(repo comment.Repository) *MockCommentService {
	return &MockCommentService{
		repo: repo,
	}
}

// AddComment creates a new comment
func (s *MockCommentService) AddComment(ctx context.Context, postID int, authorName, content string) (*comment.Comment, error) {
	c, err := comment.NewComment(postID, authorName, content)
	if err != nil {
		return nil, err
	}

	err = s.repo.Create(ctx, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// GetComment retrieves a comment by ID
func (s *MockCommentService) GetComment(ctx context.Context, id int) (*comment.Comment, error) {
	return s.repo.GetByID(ctx, id)
}

// GetCommentsByPost retrieves comments for a post with pagination
func (s *MockCommentService) GetCommentsByPost(ctx context.Context, postID int, limit, offset int) ([]*comment.Comment, error) {
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
func (s *MockCommentService) UpdateComment(ctx context.Context, id int, authorName, content string) (*comment.Comment, error) {
	// Get existing comment
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check authorization
	if !c.IsAuthor(authorName) {
		return nil, errors.New("unauthorized: only the author can update this comment")
	}

	// Update content
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
func (s *MockCommentService) DeleteComment(ctx context.Context, id int, authorName string) error {
	// Get existing comment
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check authorization
	if !c.IsAuthor(authorName) {
		return errors.New("unauthorized: only the author can delete this comment")
	}

	return s.repo.Delete(ctx, id)
}

func TestCommentService_AddComment(t *testing.T) {
	repo := NewMockCommentRepository()
	service := NewMockCommentService(repo)
	ctx := context.Background()

	tests := []struct {
		name        string
		postID      int
		authorName  string
		content     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid comment creation",
			postID:      1,
			authorName:  "John Doe",
			content:     "This is a valid comment with sufficient length.",
			expectError: false,
		},
		{
			name:        "invalid post ID should fail",
			postID:      0,
			authorName:  "John Doe",
			content:     "This is a valid comment with sufficient length.",
			expectError: true,
			errorMsg:    "post ID must be positive",
		},
		{
			name:        "empty author name should fail",
			postID:      1,
			authorName:  "",
			content:     "This is a valid comment with sufficient length.",
			expectError: true,
			errorMsg:    "author name cannot be empty",
		},
		{
			name:        "empty content should fail",
			postID:      1,
			authorName:  "John Doe",
			content:     "",
			expectError: true,
			errorMsg:    "content cannot be empty",
		},
		{
			name:        "content too short should fail",
			postID:      1,
			authorName:  "John Doe",
			content:     "Hi",
			expectError: true,
			errorMsg:    "content must be at least 3 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := service.AddComment(ctx, tt.postID, tt.authorName, tt.content)

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

			if c == nil {
				t.Error("expected comment to be created, got nil")
				return
			}

			if c.PostID != tt.postID {
				t.Errorf("expected post ID %d, got %d", tt.postID, c.PostID)
			}
			if c.AuthorName != tt.authorName {
				t.Errorf("expected author name '%s', got '%s'", tt.authorName, c.AuthorName)
			}
			if c.Content != tt.content {
				t.Errorf("expected content '%s', got '%s'", tt.content, c.Content)
			}
		})
	}
}

func TestCommentService_GetComment(t *testing.T) {
	repo := NewMockCommentRepository()
	service := NewMockCommentService(repo)
	ctx := context.Background()

	// Test getting non-existent comment
	_, err := service.GetComment(ctx, 999)
	if err != comment.ErrCommentNotFound {
		t.Errorf("expected ErrCommentNotFound, got %v", err)
	}

	// Create and retrieve comment
	c, err := service.AddComment(ctx, 1, "John Doe", "Test comment content")
	if err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}

	retrieved, err := service.GetComment(ctx, c.ID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if retrieved.ID != c.ID {
		t.Errorf("expected ID %d, got %d", c.ID, retrieved.ID)
	}
	if retrieved.Content != c.Content {
		t.Errorf("expected content '%s', got '%s'", c.Content, retrieved.Content)
	}
}

func TestCommentService_GetCommentsByPost(t *testing.T) {
	repo := NewMockCommentRepository()
	service := NewMockCommentService(repo)
	ctx := context.Background()

	// Create comments for different posts
	for i := 1; i <= 5; i++ {
		postID := 1
		if i > 3 {
			postID = 2
		}
		
		_, err := service.AddComment(ctx, postID, "Author", "Comment content for testing")
		if err != nil {
			t.Fatalf("failed to create comment %d: %v", i, err)
		}
	}

	tests := []struct {
		name           string
		postID         int
		limit          int
		offset         int
		expectedCount  int
		expectError    bool
		errorMsg       string
	}{
		{
			name:          "valid pagination",
			postID:        1,
			limit:         10,
			offset:        0,
			expectedCount: 3,
			expectError:   false,
		},
		{
			name:          "limit too high should fail",
			postID:        1,
			limit:         101,
			offset:        0,
			expectedCount: 0,
			expectError:   true,
			errorMsg:      "limit must be between 1 and 100",
		},
		{
			name:          "limit zero should fail",
			postID:        1,
			limit:         0,
			offset:        0,
			expectedCount: 0,
			expectError:   true,
			errorMsg:      "limit must be between 1 and 100",
		},
		{
			name:          "negative offset should fail",
			postID:        1,
			limit:         10,
			offset:        -1,
			expectedCount: 0,
			expectError:   true,
			errorMsg:      "offset must be non-negative",
		},
		{
			name:          "pagination with limit",
			postID:        1,
			limit:         2,
			offset:        0,
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "pagination with offset",
			postID:        1,
			limit:         2,
			offset:        1,
			expectedCount: 2,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comments, err := service.GetCommentsByPost(ctx, tt.postID, tt.limit, tt.offset)

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

			if len(comments) != tt.expectedCount {
				t.Errorf("expected %d comments, got %d", tt.expectedCount, len(comments))
			}

			for _, c := range comments {
				if c.PostID != tt.postID {
					t.Errorf("expected post ID %d, got %d", tt.postID, c.PostID)
				}
			}
		})
	}
}

func TestCommentService_UpdateComment(t *testing.T) {
	repo := NewMockCommentRepository()
	service := NewMockCommentService(repo)
	ctx := context.Background()

	// Create a comment
	c, err := service.AddComment(ctx, 1, "John Doe", "Original content")
	if err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}

	tests := []struct {
		name        string
		id          int
		authorName  string
		content     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid update by author",
			id:          c.ID,
			authorName:  "John Doe",
			content:     "Updated content with sufficient length",
			expectError: false,
		},
		{
			name:        "unauthorized update should fail",
			id:          c.ID,
			authorName:  "Jane Smith",
			content:     "Updated content with sufficient length",
			expectError: true,
			errorMsg:    "unauthorized: only the author can update this comment",
		},
		{
			name:        "update non-existent comment should fail",
			id:          999,
			authorName:  "John Doe",
			content:     "Updated content with sufficient length",
			expectError: true,
			errorMsg:    "comment not found",
		},
		{
			name:        "update with empty content should fail",
			id:          c.ID,
			authorName:  "John Doe",
			content:     "",
			expectError: true,
			errorMsg:    "content cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, err := service.UpdateComment(ctx, tt.id, tt.authorName, tt.content)

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

			if updated == nil {
				t.Error("expected updated comment, got nil")
				return
			}

			if updated.Content != tt.content {
				t.Errorf("expected content '%s', got '%s'", tt.content, updated.Content)
			}
		})
	}
}

func TestCommentService_DeleteComment(t *testing.T) {
	repo := NewMockCommentRepository()
	service := NewMockCommentService(repo)
	ctx := context.Background()

	// Create a comment
	c, err := service.AddComment(ctx, 1, "John Doe", "Test comment content")
	if err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}

	tests := []struct {
		name        string
		id          int
		authorName  string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "unauthorized deletion should fail",
			id:          c.ID,
			authorName:  "Jane Smith",
			expectError: true,
			errorMsg:    "unauthorized: only the author can delete this comment",
		},
		{
			name:        "delete non-existent comment should fail",
			id:          999,
			authorName:  "John Doe",
			expectError: true,
			errorMsg:    "comment not found",
		},
		{
			name:        "valid deletion by author",
			id:          c.ID,
			authorName:  "John Doe",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DeleteComment(ctx, tt.id, tt.authorName)

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

			// Verify deletion
			_, err = service.GetComment(ctx, tt.id)
			if err != comment.ErrCommentNotFound {
				t.Errorf("expected ErrCommentNotFound after deletion, got %v", err)
			}
		})
	}
}
