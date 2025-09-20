package post_test

import (
	"testing"
	"time"

	"blog-platform/internal/domain/post"
)

func TestNewPost(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		content     string
		authorID    int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid post creation",
			title:       "Test Post Title",
			content:     "This is a test post content with sufficient length.",
			authorID:    1,
			expectError: false,
		},
		{
			name:        "empty title should fail",
			title:       "",
			content:     "This is a test post content with sufficient length.",
			authorID:    1,
			expectError: true,
			errorMsg:    "title cannot be empty",
		},
		{
			name:        "empty content should fail",
			title:       "Test Post Title",
			content:     "",
			authorID:    1,
			expectError: true,
			errorMsg:    "content cannot be empty",
		},
		{
			name:        "invalid author ID should fail",
			title:       "Test Post Title",
			content:     "This is a test post content with sufficient length.",
			authorID:    0,
			expectError: true,
			errorMsg:    "author ID must be positive",
		},
		{
			name:        "title too long should fail",
			title:       string(make([]byte, 501)), // 501 characters
			content:     "This is a test post content with sufficient length.",
			authorID:    1,
			expectError: true,
			errorMsg:    "title cannot exceed 500 characters",
		},
		{
			name:        "content too short should fail",
			title:       "Test Post Title",
			content:     "Short",
			authorID:    1,
			expectError: true,
			errorMsg:    "content must be at least 10 characters",
		},
		{
			name:        "content too long should fail",
			title:       "Test Post Title",
			content:     string(make([]byte, 10001)), // 10001 characters
			authorID:    1,
			expectError: true,
			errorMsg:    "content cannot exceed 10000 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := post.NewPost(tt.title, tt.content, tt.authorID)

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

			if p.AuthorID != tt.authorID {
				t.Errorf("expected author ID %d, got %d", tt.authorID, p.AuthorID)
			}

			if p.ID != 0 {
				t.Errorf("expected ID to be 0 for new post, got %d", p.ID)
			}

			if p.CreatedAt.IsZero() {
				t.Error("expected CreatedAt to be set")
			}

			if p.UpdatedAt.IsZero() {
				t.Error("expected UpdatedAt to be set")
			}
		})
	}
}

func TestPost_Update(t *testing.T) {
	// Create a valid post first
	p, err := post.NewPost("Original Title", "Original content with sufficient length.", 1)
	if err != nil {
		t.Fatalf("failed to create post: %v", err)
	}

	originalUpdatedAt := p.UpdatedAt
	time.Sleep(time.Millisecond) // Ensure time difference

	tests := []struct {
		name        string
		title       string
		content     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid update",
			title:       "Updated Title",
			content:     "Updated content with sufficient length for validation.",
			expectError: false,
		},
		{
			name:        "empty title should fail",
			title:       "",
			content:     "Updated content with sufficient length for validation.",
			expectError: true,
			errorMsg:    "title cannot be empty",
		},
		{
			name:        "empty content should fail",
			title:       "Updated Title",
			content:     "",
			expectError: true,
			errorMsg:    "content cannot be empty",
		},
		{
			name:        "title too long should fail",
			title:       string(make([]byte, 501)),
			content:     "Updated content with sufficient length for validation.",
			expectError: true,
			errorMsg:    "title cannot exceed 500 characters",
		},
		{
			name:        "content too short should fail",
			title:       "Updated Title",
			content:     "Short",
			expectError: true,
			errorMsg:    "content must be at least 10 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh post for each test
			testPost, _ := post.NewPost("Original Title", "Original content with sufficient length.", 1)
			
			err := testPost.Update(tt.title, tt.content)

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

			if testPost.Title != tt.title {
				t.Errorf("expected title '%s', got '%s'", tt.title, testPost.Title)
			}

			if testPost.Content != tt.content {
				t.Errorf("expected content '%s', got '%s'", tt.content, testPost.Content)
			}

			if !testPost.UpdatedAt.After(originalUpdatedAt) {
				t.Error("expected UpdatedAt to be updated")
			}
		})
	}
}

func TestPost_IsAuthor(t *testing.T) {
	p, err := post.NewPost("Test Title", "Test content with sufficient length.", 123)
	if err != nil {
		t.Fatalf("failed to create post: %v", err)
	}

	tests := []struct {
		name     string
		userID   int
		expected bool
	}{
		{
			name:     "correct author",
			userID:   123,
			expected: true,
		},
		{
			name:     "incorrect author",
			userID:   456,
			expected: false,
		},
		{
			name:     "zero user ID",
			userID:   0,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.IsAuthor(tt.userID)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPost_Validation(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		content string
		valid   bool
	}{
		{
			name:    "valid post",
			title:   "Valid Title",
			content: "Valid content with sufficient length for validation.",
			valid:   true,
		},
		{
			name:    "title with special characters",
			title:   "Title with @#$%^&*() special chars!",
			content: "Valid content with sufficient length for validation.",
			valid:   true,
		},
		{
			name:    "content with newlines and formatting",
			title:   "Valid Title",
			content: "Content with\nnewlines and\ttabs and proper length.",
			valid:   true,
		},
		{
			name:    "maximum length title",
			title:   string(make([]byte, 500)),
			content: "Valid content with sufficient length for validation.",
			valid:   true,
		},
		{
			name:    "maximum length content",
			title:   "Valid Title",
			content: string(make([]byte, 10000)),
			valid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := post.NewPost(tt.title, tt.content, 1)

			if tt.valid {
				if err != nil {
					t.Errorf("expected valid post, got error: %v", err)
				}
				if p == nil {
					t.Error("expected post to be created")
				}
			} else {
				if err == nil {
					t.Error("expected error for invalid post")
				}
			}
		})
	}
}
