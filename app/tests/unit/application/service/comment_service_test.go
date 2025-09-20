package service_test

import (
	"context"
	"testing"

	"blog-platform/internal/application/service"
	"blog-platform/internal/domain/comment"
)

// MockCommentRepository implements the comment.Repository interface for testing
type MockCommentRepository struct {
	comments map[int]*comment.Comment
	nextID   int
}

// NewMockCommentRepository creates a new mock repository
func NewMockCommentRepository() *MockCommentRepository {
	return &MockCommentRepository{
		comments: make(map[int]*comment.Comment),
		nextID:   1,
	}
}

// Create adds a new comment to the mock repository
func (m *MockCommentRepository) Create(ctx context.Context, c *comment.Comment) error {
	c.ID = m.nextID
	m.comments[c.ID] = c
	m.nextID++
	return nil
}

// GetByID retrieves a comment by its ID
func (m *MockCommentRepository) GetByID(ctx context.Context, id int) (*comment.Comment, error) {
	c, exists := m.comments[id]
	if !exists {
		return nil, comment.ErrCommentNotFound
	}
	return c, nil
}

// GetByPostID retrieves comments for a specific post with pagination
func (m *MockCommentRepository) GetByPostID(ctx context.Context, postID int, limit, offset int) ([]*comment.Comment, error) {
	var result []*comment.Comment
	count := 0
	
	for _, c := range m.comments {
		if c.PostID == postID {
			if count >= offset {
				result = append(result, c)
				if len(result) >= limit {
					break
				}
			}
			count++
		}
	}
	
	return result, nil
}

// Update modifies an existing comment
func (m *MockCommentRepository) Update(ctx context.Context, c *comment.Comment) error {
	if _, exists := m.comments[c.ID]; !exists {
		return comment.ErrCommentNotFound
	}
	m.comments[c.ID] = c
	return nil
}

// Delete removes a comment from the repository
func (m *MockCommentRepository) Delete(ctx context.Context, id int) error {
	if _, exists := m.comments[id]; !exists {
		return comment.ErrCommentNotFound
	}
	delete(m.comments, id)
	return nil
}

func TestCommentService_Implementation(t *testing.T) {
	repo := NewMockCommentRepository()
	
	// Verify that CommentService implements the Service interface
	var _ comment.Service = service.NewCommentService(repo)
}

func TestCommentService_AddComment_Integration(t *testing.T) {
	repo := NewMockCommentRepository()
	commentService := service.NewCommentService(repo)
	ctx := context.Background()

	// Test successful comment creation
	c, err := commentService.AddComment(ctx, 1, "John Doe", "Test comment content")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if c == nil {
		t.Fatal("expected comment to be created, got nil")
	}

	if c.ID == 0 {
		t.Error("expected comment ID to be set")
	}

	if c.PostID != 1 {
		t.Errorf("expected post ID 1, got %d", c.PostID)
	}

	if c.AuthorName != "John Doe" {
		t.Errorf("expected author name 'John Doe', got '%s'", c.AuthorName)
	}

	if c.Content != "Test comment content" {
		t.Errorf("expected content 'Test comment content', got '%s'", c.Content)
	}

	// Test validation error
	_, err = commentService.AddComment(ctx, 0, "John Doe", "Test comment content")
	if err == nil {
		t.Error("expected error for invalid post ID")
	}
}

func TestCommentService_GetComment_Integration(t *testing.T) {
	repo := NewMockCommentRepository()
	commentService := service.NewCommentService(repo)
	ctx := context.Background()

	// Test getting non-existent comment
	_, err := commentService.GetComment(ctx, 999)
	if err != comment.ErrCommentNotFound {
		t.Errorf("expected ErrCommentNotFound, got %v", err)
	}

	// Test invalid ID
	_, err = commentService.GetComment(ctx, 0)
	if err == nil {
		t.Error("expected error for invalid comment ID")
	}

	// Create and retrieve comment
	created, err := commentService.AddComment(ctx, 1, "John Doe", "Test comment content")
	if err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}

	retrieved, err := commentService.GetComment(ctx, created.ID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, retrieved.ID)
	}
}

func TestCommentService_GetCommentsByPost_Integration(t *testing.T) {
	repo := NewMockCommentRepository()
	commentService := service.NewCommentService(repo)
	ctx := context.Background()

	// Create comments for different posts
	for i := 1; i <= 5; i++ {
		postID := 1
		if i > 3 {
			postID = 2
		}
		
		_, err := commentService.AddComment(ctx, postID, "Author", "Comment content for testing")
		if err != nil {
			t.Fatalf("failed to create comment %d: %v", i, err)
		}
	}

	// Test successful retrieval
	comments, err := commentService.GetCommentsByPost(ctx, 1, 10, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(comments) != 3 {
		t.Errorf("expected 3 comments for post 1, got %d", len(comments))
	}

	// Test validation errors
	_, err = commentService.GetCommentsByPost(ctx, 0, 10, 0)
	if err == nil {
		t.Error("expected error for invalid post ID")
	}

	_, err = commentService.GetCommentsByPost(ctx, 1, 0, 0)
	if err == nil {
		t.Error("expected error for invalid limit")
	}

	_, err = commentService.GetCommentsByPost(ctx, 1, 10, -1)
	if err == nil {
		t.Error("expected error for invalid offset")
	}
}

func TestCommentService_UpdateComment_Integration(t *testing.T) {
	repo := NewMockCommentRepository()
	commentService := service.NewCommentService(repo)
	ctx := context.Background()

	// Create a comment
	created, err := commentService.AddComment(ctx, 1, "John Doe", "Original content")
	if err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}

	// Test successful update
	updated, err := commentService.UpdateComment(ctx, created.ID, "John Doe", "Updated content")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if updated.Content != "Updated content" {
		t.Errorf("expected content 'Updated content', got '%s'", updated.Content)
	}

	// Test authorization failure
	_, err = commentService.UpdateComment(ctx, created.ID, "Jane Smith", "Unauthorized update")
	if err == nil {
		t.Error("expected error for unauthorized update")
	}

	// Test validation errors
	_, err = commentService.UpdateComment(ctx, 0, "John Doe", "Updated content")
	if err == nil {
		t.Error("expected error for invalid comment ID")
	}

	_, err = commentService.UpdateComment(ctx, 999, "John Doe", "Updated content")
	if err != comment.ErrCommentNotFound {
		t.Errorf("expected ErrCommentNotFound, got %v", err)
	}
}

func TestCommentService_DeleteComment_Integration(t *testing.T) {
	repo := NewMockCommentRepository()
	commentService := service.NewCommentService(repo)
	ctx := context.Background()

	// Create a comment
	created, err := commentService.AddComment(ctx, 1, "John Doe", "Test comment content")
	if err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}

	// Test authorization failure
	err = commentService.DeleteComment(ctx, created.ID, "Jane Smith")
	if err == nil {
		t.Error("expected error for unauthorized deletion")
	}

	// Test validation errors
	err = commentService.DeleteComment(ctx, 0, "John Doe")
	if err == nil {
		t.Error("expected error for invalid comment ID")
	}

	err = commentService.DeleteComment(ctx, 999, "John Doe")
	if err != comment.ErrCommentNotFound {
		t.Errorf("expected ErrCommentNotFound, got %v", err)
	}

	// Test successful deletion
	err = commentService.DeleteComment(ctx, created.ID, "John Doe")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify deletion
	_, err = commentService.GetComment(ctx, created.ID)
	if err != comment.ErrCommentNotFound {
		t.Errorf("expected ErrCommentNotFound after deletion, got %v", err)
	}
}
