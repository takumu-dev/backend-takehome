package post_test

import (
	"context"
	"testing"

	"blog-platform/internal/domain/post"
)

// MockPostRepository implements the PostRepository interface for testing
type MockPostRepository struct {
	posts  map[int]*post.Post
	nextID int
}

func NewMockPostRepository() *MockPostRepository {
	return &MockPostRepository{
		posts:  make(map[int]*post.Post),
		nextID: 1,
	}
}

func (m *MockPostRepository) Create(ctx context.Context, p *post.Post) error {
	if p == nil {
		return post.ErrInvalidPostData
	}
	
	p.ID = m.nextID
	m.nextID++
	m.posts[p.ID] = p
	return nil
}

func (m *MockPostRepository) GetByID(ctx context.Context, id int) (*post.Post, error) {
	if p, exists := m.posts[id]; exists {
		return p, nil
	}
	return nil, post.ErrPostNotFound
}

func (m *MockPostRepository) GetByAuthorID(ctx context.Context, authorID int, limit, offset int) ([]*post.Post, error) {
	var posts []*post.Post
	count := 0
	skipped := 0
	
	for _, p := range m.posts {
		if p.AuthorID == authorID {
			if skipped < offset {
				skipped++
				continue
			}
			if count >= limit {
				break
			}
			posts = append(posts, p)
			count++
		}
	}
	
	return posts, nil
}

func (m *MockPostRepository) List(ctx context.Context, limit, offset int) ([]*post.Post, error) {
	var posts []*post.Post
	count := 0
	skipped := 0
	
	for _, p := range m.posts {
		if skipped < offset {
			skipped++
			continue
		}
		if count >= limit {
			break
		}
		posts = append(posts, p)
		count++
	}
	
	return posts, nil
}

func (m *MockPostRepository) Update(ctx context.Context, p *post.Post) error {
	if p == nil {
		return post.ErrInvalidPostData
	}
	
	if _, exists := m.posts[p.ID]; !exists {
		return post.ErrPostNotFound
	}
	
	m.posts[p.ID] = p
	return nil
}

func (m *MockPostRepository) Delete(ctx context.Context, id int) error {
	if _, exists := m.posts[id]; !exists {
		return post.ErrPostNotFound
	}
	
	delete(m.posts, id)
	return nil
}

func TestPostRepository_Create(t *testing.T) {
	repo := NewMockPostRepository()
	ctx := context.Background()

	// Create a valid post
	p, err := post.NewPost("Test Title", "Test content with sufficient length.", 1)
	if err != nil {
		t.Fatalf("failed to create post: %v", err)
	}

	// Test successful creation
	err = repo.Create(ctx, p)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if p.ID == 0 {
		t.Error("expected post ID to be set after creation")
	}

	// Test creation with nil post
	err = repo.Create(ctx, nil)
	if err != post.ErrInvalidPostData {
		t.Errorf("expected ErrInvalidPostData, got %v", err)
	}
}

func TestPostRepository_GetByID(t *testing.T) {
	repo := NewMockPostRepository()
	ctx := context.Background()

	// Create and store a post
	p, _ := post.NewPost("Test Title", "Test content with sufficient length.", 1)
	repo.Create(ctx, p)

	// Test successful retrieval
	retrieved, err := repo.GetByID(ctx, p.ID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if retrieved == nil {
		t.Error("expected post to be retrieved")
		return
	}

	if retrieved.ID != p.ID {
		t.Errorf("expected post ID %d, got %d", p.ID, retrieved.ID)
	}

	if retrieved.Title != p.Title {
		t.Errorf("expected title '%s', got '%s'", p.Title, retrieved.Title)
	}

	// Test retrieval of non-existent post
	_, err = repo.GetByID(ctx, 999)
	if err != post.ErrPostNotFound {
		t.Errorf("expected ErrPostNotFound, got %v", err)
	}
}

func TestPostRepository_GetByAuthorID(t *testing.T) {
	repo := NewMockPostRepository()
	ctx := context.Background()

	// Create posts by different authors
	p1, _ := post.NewPost("Post 1", "Content for post 1 with sufficient length.", 1)
	p2, _ := post.NewPost("Post 2", "Content for post 2 with sufficient length.", 1)
	p3, _ := post.NewPost("Post 3", "Content for post 3 with sufficient length.", 2)

	repo.Create(ctx, p1)
	repo.Create(ctx, p2)
	repo.Create(ctx, p3)

	// Test getting posts by author ID 1
	posts, err := repo.GetByAuthorID(ctx, 1, 10, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(posts) != 2 {
		t.Errorf("expected 2 posts, got %d", len(posts))
	}

	// Test getting posts by author ID 2
	posts, err = repo.GetByAuthorID(ctx, 2, 10, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(posts) != 1 {
		t.Errorf("expected 1 post, got %d", len(posts))
	}

	// Test getting posts by non-existent author
	posts, err = repo.GetByAuthorID(ctx, 999, 10, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(posts) != 0 {
		t.Errorf("expected 0 posts, got %d", len(posts))
	}

	// Test pagination
	posts, err = repo.GetByAuthorID(ctx, 1, 1, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(posts) != 1 {
		t.Errorf("expected 1 post with limit, got %d", len(posts))
	}
}

func TestPostRepository_List(t *testing.T) {
	repo := NewMockPostRepository()
	ctx := context.Background()

	// Create multiple posts
	for i := 1; i <= 5; i++ {
		p, _ := post.NewPost("Post Title", "Post content with sufficient length for validation.", i)
		repo.Create(ctx, p)
	}

	// Test listing all posts
	posts, err := repo.List(ctx, 10, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(posts) != 5 {
		t.Errorf("expected 5 posts, got %d", len(posts))
	}

	// Test pagination with limit
	posts, err = repo.List(ctx, 3, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(posts) != 3 {
		t.Errorf("expected 3 posts with limit, got %d", len(posts))
	}

	// Test pagination with offset
	posts, err = repo.List(ctx, 10, 2)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(posts) != 3 {
		t.Errorf("expected 3 posts with offset, got %d", len(posts))
	}
}

func TestPostRepository_Update(t *testing.T) {
	repo := NewMockPostRepository()
	ctx := context.Background()

	// Create and store a post
	p, _ := post.NewPost("Original Title", "Original content with sufficient length.", 1)
	repo.Create(ctx, p)

	// Update the post
	p.Update("Updated Title", "Updated content with sufficient length for validation.")

	// Test successful update
	err := repo.Update(ctx, p)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify the update
	updated, err := repo.GetByID(ctx, p.ID)
	if err != nil {
		t.Errorf("failed to retrieve updated post: %v", err)
	}

	if updated.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got '%s'", updated.Title)
	}

	// Test update with nil post
	err = repo.Update(ctx, nil)
	if err != post.ErrInvalidPostData {
		t.Errorf("expected ErrInvalidPostData, got %v", err)
	}

	// Test update of non-existent post
	nonExistent, _ := post.NewPost("Non-existent", "Non-existent content with sufficient length.", 1)
	nonExistent.ID = 999
	err = repo.Update(ctx, nonExistent)
	if err != post.ErrPostNotFound {
		t.Errorf("expected ErrPostNotFound, got %v", err)
	}
}

func TestPostRepository_Delete(t *testing.T) {
	repo := NewMockPostRepository()
	ctx := context.Background()

	// Create and store a post
	p, _ := post.NewPost("Test Title", "Test content with sufficient length.", 1)
	repo.Create(ctx, p)

	// Test successful deletion
	err := repo.Delete(ctx, p.ID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify the post is deleted
	_, err = repo.GetByID(ctx, p.ID)
	if err != post.ErrPostNotFound {
		t.Errorf("expected post to be deleted, but it still exists")
	}

	// Test deletion of non-existent post
	err = repo.Delete(ctx, 999)
	if err != post.ErrPostNotFound {
		t.Errorf("expected ErrPostNotFound, got %v", err)
	}
}

func TestPostRepository_Interface(t *testing.T) {
	// Test that our mock implements the interface
	var _ post.Repository = NewMockPostRepository()
}
