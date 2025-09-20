package comment_test

import (
	"context"
	"testing"

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

func TestCommentRepository_Create(t *testing.T) {
	repo := NewMockCommentRepository()
	ctx := context.Background()

	c, err := comment.NewComment(1, "John Doe", "Test comment content")
	if err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}

	err = repo.Create(ctx, c)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if c.ID == 0 {
		t.Error("expected comment ID to be set after creation")
	}

	// Verify comment was stored
	stored, err := repo.GetByID(ctx, c.ID)
	if err != nil {
		t.Errorf("failed to retrieve created comment: %v", err)
	}

	if stored.PostID != c.PostID {
		t.Errorf("expected post ID %d, got %d", c.PostID, stored.PostID)
	}
	if stored.AuthorName != c.AuthorName {
		t.Errorf("expected author name '%s', got '%s'", c.AuthorName, stored.AuthorName)
	}
	if stored.Content != c.Content {
		t.Errorf("expected content '%s', got '%s'", c.Content, stored.Content)
	}
}

func TestCommentRepository_GetByID(t *testing.T) {
	repo := NewMockCommentRepository()
	ctx := context.Background()

	// Test getting non-existent comment
	_, err := repo.GetByID(ctx, 999)
	if err != comment.ErrCommentNotFound {
		t.Errorf("expected ErrCommentNotFound, got %v", err)
	}

	// Create and retrieve comment
	c, err := comment.NewComment(1, "John Doe", "Test comment content")
	if err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}

	err = repo.Create(ctx, c)
	if err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}

	retrieved, err := repo.GetByID(ctx, c.ID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if retrieved.ID != c.ID {
		t.Errorf("expected ID %d, got %d", c.ID, retrieved.ID)
	}
	if retrieved.PostID != c.PostID {
		t.Errorf("expected post ID %d, got %d", c.PostID, retrieved.PostID)
	}
	if retrieved.AuthorName != c.AuthorName {
		t.Errorf("expected author name '%s', got '%s'", c.AuthorName, retrieved.AuthorName)
	}
	if retrieved.Content != c.Content {
		t.Errorf("expected content '%s', got '%s'", c.Content, retrieved.Content)
	}
}

func TestCommentRepository_GetByPostID(t *testing.T) {
	repo := NewMockCommentRepository()
	ctx := context.Background()

	// Create comments for different posts
	comments := []*comment.Comment{}
	for i := 1; i <= 5; i++ {
		postID := 1
		if i > 3 {
			postID = 2
		}
		
		c, err := comment.NewComment(postID, "Author", "Comment content for testing")
		if err != nil {
			t.Fatalf("failed to create comment %d: %v", i, err)
		}
		
		err = repo.Create(ctx, c)
		if err != nil {
			t.Fatalf("failed to store comment %d: %v", i, err)
		}
		comments = append(comments, c)
	}

	// Test getting comments for post 1 (should have 3 comments)
	result, err := repo.GetByPostID(ctx, 1, 10, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 comments for post 1, got %d", len(result))
	}

	for _, c := range result {
		if c.PostID != 1 {
			t.Errorf("expected post ID 1, got %d", c.PostID)
		}
	}

	// Test getting comments for post 2 (should have 2 comments)
	result, err = repo.GetByPostID(ctx, 2, 10, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 comments for post 2, got %d", len(result))
	}

	// Test pagination
	result, err = repo.GetByPostID(ctx, 1, 2, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 comments with limit 2, got %d", len(result))
	}

	// Test offset
	result, err = repo.GetByPostID(ctx, 1, 2, 1)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 comments with offset 1, got %d", len(result))
	}

	// Test non-existent post
	result, err = repo.GetByPostID(ctx, 999, 10, 0)
	if err != nil {
		t.Errorf("expected no error for non-existent post, got %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 comments for non-existent post, got %d", len(result))
	}
}

func TestCommentRepository_Update(t *testing.T) {
	repo := NewMockCommentRepository()
	ctx := context.Background()

	// Test updating non-existent comment
	c, err := comment.NewComment(1, "John Doe", "Test comment content")
	if err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}
	c.ID = 999

	err = repo.Update(ctx, c)
	if err != comment.ErrCommentNotFound {
		t.Errorf("expected ErrCommentNotFound, got %v", err)
	}

	// Create and update comment
	c, err = comment.NewComment(1, "John Doe", "Original content")
	if err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}

	err = repo.Create(ctx, c)
	if err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}

	// Update the comment
	err = c.Update("Updated content for testing")
	if err != nil {
		t.Fatalf("failed to update comment content: %v", err)
	}

	err = repo.Update(ctx, c)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify update
	updated, err := repo.GetByID(ctx, c.ID)
	if err != nil {
		t.Errorf("failed to retrieve updated comment: %v", err)
	}

	if updated.Content != "Updated content for testing" {
		t.Errorf("expected updated content, got '%s'", updated.Content)
	}
}

func TestCommentRepository_Delete(t *testing.T) {
	repo := NewMockCommentRepository()
	ctx := context.Background()

	// Test deleting non-existent comment
	err := repo.Delete(ctx, 999)
	if err != comment.ErrCommentNotFound {
		t.Errorf("expected ErrCommentNotFound, got %v", err)
	}

	// Create and delete comment
	c, err := comment.NewComment(1, "John Doe", "Test comment content")
	if err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}

	err = repo.Create(ctx, c)
	if err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}

	err = repo.Delete(ctx, c.ID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(ctx, c.ID)
	if err != comment.ErrCommentNotFound {
		t.Errorf("expected ErrCommentNotFound after deletion, got %v", err)
	}
}

func TestCommentRepository_Interface(t *testing.T) {
	// Verify that MockCommentRepository implements the Repository interface
	var _ comment.Repository = (*MockCommentRepository)(nil)
}
