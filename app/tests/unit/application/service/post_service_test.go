package service_test

import (
	"context"
	"testing"

	"blog-platform/internal/application/service"
	"blog-platform/internal/domain/post"
)

// MockLogger implements the logging.Logger interface for testing
type MockLogger struct{}

func (m *MockLogger) Info(ctx context.Context, msg string, args ...any)  {}
func (m *MockLogger) Error(ctx context.Context, msg string, args ...any) {}
func (m *MockLogger) Warn(ctx context.Context, msg string, args ...any)  {}
func (m *MockLogger) Debug(ctx context.Context, msg string, args ...any) {}

func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

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

func TestPostService_Implementation(t *testing.T) {
	// Test that our concrete service implements the interface
	repo := NewMockPostRepository()
	var _ post.Service = service.NewPostService(repo, NewMockLogger())
}

func TestPostService_CreatePost_Integration(t *testing.T) {
	repo := NewMockPostRepository()
	postService := service.NewPostService(repo, NewMockLogger())
	ctx := context.Background()

	// Test successful post creation
	p, err := postService.CreatePost(ctx, 1, "Test Post", "Test content with sufficient length.")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if p == nil {
		t.Error("expected post to be created, got nil")
		return
	}

	if p.ID == 0 {
		t.Error("expected post ID to be set")
	}

	if p.Title != "Test Post" {
		t.Errorf("expected title 'Test Post', got '%s'", p.Title)
	}

	if p.AuthorID != 1 {
		t.Errorf("expected author ID 1, got %d", p.AuthorID)
	}

	// Test creation with invalid data
	_, err = postService.CreatePost(ctx, 0, "Test Post", "Test content with sufficient length.")
	if err == nil {
		t.Error("expected error for invalid author ID")
	}
}

func TestPostService_GetPost_Integration(t *testing.T) {
	repo := NewMockPostRepository()
	postService := service.NewPostService(repo, NewMockLogger())
	ctx := context.Background()

	// Create a post first
	createdPost, err := postService.CreatePost(ctx, 1, "Test Post", "Test content with sufficient length.")
	if err != nil {
		t.Fatalf("failed to create post: %v", err)
	}

	// Test successful retrieval
	retrievedPost, err := postService.GetPost(ctx, createdPost.ID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if retrievedPost.ID != createdPost.ID {
		t.Errorf("expected post ID %d, got %d", createdPost.ID, retrievedPost.ID)
	}

	// Test retrieval of non-existent post
	_, err = postService.GetPost(ctx, 999)
	if err != post.ErrPostNotFound {
		t.Errorf("expected ErrPostNotFound, got %v", err)
	}
}

func TestPostService_UpdatePost_Integration(t *testing.T) {
	repo := NewMockPostRepository()
	postService := service.NewPostService(repo, NewMockLogger())
	ctx := context.Background()

	// Create a post first
	createdPost, err := postService.CreatePost(ctx, 1, "Original Title", "Original content with sufficient length.")
	if err != nil {
		t.Fatalf("failed to create post: %v", err)
	}

	// Test successful update by author
	updatedPost, err := postService.UpdatePost(ctx, 1, createdPost.ID, "Updated Title", "Updated content with sufficient length.")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if updatedPost.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got '%s'", updatedPost.Title)
	}

	// Test unauthorized update
	_, err = postService.UpdatePost(ctx, 2, createdPost.ID, "Unauthorized Update", "Unauthorized content with sufficient length.")
	if err != post.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}

	// Test update of non-existent post
	_, err = postService.UpdatePost(ctx, 1, 999, "Non-existent", "Non-existent content with sufficient length.")
	if err != post.ErrPostNotFound {
		t.Errorf("expected ErrPostNotFound, got %v", err)
	}
}

func TestPostService_DeletePost_Integration(t *testing.T) {
	repo := NewMockPostRepository()
	postService := service.NewPostService(repo, NewMockLogger())
	ctx := context.Background()

	// Create a post first
	createdPost, err := postService.CreatePost(ctx, 1, "Test Post", "Test content with sufficient length.")
	if err != nil {
		t.Fatalf("failed to create post: %v", err)
	}

	// Test unauthorized deletion
	err = postService.DeletePost(ctx, 2, createdPost.ID)
	if err != post.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}

	// Test successful deletion by author
	err = postService.DeletePost(ctx, 1, createdPost.ID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify the post is deleted
	_, err = postService.GetPost(ctx, createdPost.ID)
	if err != post.ErrPostNotFound {
		t.Error("expected post to be deleted")
	}

	// Test deletion of non-existent post
	err = postService.DeletePost(ctx, 1, 999)
	if err != post.ErrPostNotFound {
		t.Errorf("expected ErrPostNotFound, got %v", err)
	}
}

func TestPostService_GetPostsByAuthor_Integration(t *testing.T) {
	repo := NewMockPostRepository()
	postService := service.NewPostService(repo, NewMockLogger())
	ctx := context.Background()

	// Create posts by different authors
	postService.CreatePost(ctx, 1, "Post 1", "Content for post 1 with sufficient length.")
	postService.CreatePost(ctx, 1, "Post 2", "Content for post 2 with sufficient length.")
	postService.CreatePost(ctx, 2, "Post 3", "Content for post 3 with sufficient length.")

	// Test getting posts by author ID 1
	posts, err := postService.GetPostsByAuthor(ctx, 1, 10, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(posts) != 2 {
		t.Errorf("expected 2 posts, got %d", len(posts))
	}

	// Test pagination
	posts, err = postService.GetPostsByAuthor(ctx, 1, 1, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(posts) != 1 {
		t.Errorf("expected 1 post with limit, got %d", len(posts))
	}

	// Test limit validation (should cap at 100)
	posts, err = postService.GetPostsByAuthor(ctx, 1, 200, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Should still return posts since we only have 2
	if len(posts) != 2 {
		t.Errorf("expected 2 posts, got %d", len(posts))
	}
}

func TestPostService_ListPosts_Integration(t *testing.T) {
	repo := NewMockPostRepository()
	postService := service.NewPostService(repo, NewMockLogger())
	ctx := context.Background()

	// Create multiple posts
	for i := 1; i <= 5; i++ {
		postService.CreatePost(ctx, i, "Post Title", "Post content with sufficient length for validation.")
	}

	// Test listing all posts
	posts, err := postService.ListPosts(ctx, 10, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(posts) != 5 {
		t.Errorf("expected 5 posts, got %d", len(posts))
	}

	// Test pagination with limit
	posts, err = postService.ListPosts(ctx, 3, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(posts) != 3 {
		t.Errorf("expected 3 posts with limit, got %d", len(posts))
	}

	// Test pagination with offset
	posts, err = postService.ListPosts(ctx, 10, 2)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(posts) != 3 {
		t.Errorf("expected 3 posts with offset, got %d", len(posts))
	}
}
