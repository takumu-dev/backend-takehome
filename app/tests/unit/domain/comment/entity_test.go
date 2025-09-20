package comment_test

import (
	"testing"

	"blog-platform/internal/domain/comment"
)

func TestNewComment(t *testing.T) {
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
			content:     "This is a test comment with sufficient length.",
			expectError: false,
		},
		{
			name:        "invalid post ID should fail",
			postID:      0,
			authorName:  "John Doe",
			content:     "This is a test comment with sufficient length.",
			expectError: true,
			errorMsg:    "post ID must be positive",
		},
		{
			name:        "empty author name should fail",
			postID:      1,
			authorName:  "",
			content:     "This is a test comment with sufficient length.",
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
			name:        "author name too long should fail",
			postID:      1,
			authorName:  string(make([]byte, 256)), // 256 characters
			content:     "This is a test comment with sufficient length.",
			expectError: true,
			errorMsg:    "author name cannot exceed 255 characters",
		},
		{
			name:        "content too short should fail",
			postID:      1,
			authorName:  "John Doe",
			content:     "Hi",
			expectError: true,
			errorMsg:    "content must be at least 3 characters",
		},
		{
			name:        "content too long should fail",
			postID:      1,
			authorName:  "John Doe",
			content:     string(make([]byte, 1001)), // 1001 characters
			expectError: true,
			errorMsg:    "content cannot exceed 1000 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := comment.NewComment(tt.postID, tt.authorName, tt.content)

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

			if c.ID != 0 {
				t.Errorf("expected ID to be 0 for new comment, got %d", c.ID)
			}

			if c.CreatedAt.IsZero() {
				t.Error("expected CreatedAt to be set")
			}
		})
	}
}

func TestComment_Update(t *testing.T) {
	// Create a valid comment first
	_, err := comment.NewComment(1, "John Doe", "Original content with sufficient length.")
	if err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}

	tests := []struct {
		name        string
		content     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid update",
			content:     "Updated content with sufficient length for validation.",
			expectError: false,
		},
		{
			name:        "empty content should fail",
			content:     "",
			expectError: true,
			errorMsg:    "content cannot be empty",
		},
		{
			name:        "content too short should fail",
			content:     "Hi",
			expectError: true,
			errorMsg:    "content must be at least 3 characters",
		},
		{
			name:        "content too long should fail",
			content:     string(make([]byte, 1001)),
			expectError: true,
			errorMsg:    "content cannot exceed 1000 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh comment for each test
			testComment, _ := comment.NewComment(1, "John Doe", "Original content with sufficient length.")
			
			err := testComment.Update(tt.content)

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

			if testComment.Content != tt.content {
				t.Errorf("expected content '%s', got '%s'", tt.content, testComment.Content)
			}
		})
	}
}

func TestComment_IsAuthor(t *testing.T) {
	c, err := comment.NewComment(1, "John Doe", "Test content with sufficient length.")
	if err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}

	tests := []struct {
		name       string
		authorName string
		expected   bool
	}{
		{
			name:       "correct author",
			authorName: "John Doe",
			expected:   true,
		},
		{
			name:       "incorrect author",
			authorName: "Jane Smith",
			expected:   false,
		},
		{
			name:       "empty author name",
			authorName: "",
			expected:   false,
		},
		{
			name:       "case sensitive check",
			authorName: "john doe",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.IsAuthor(tt.authorName)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestComment_Validation(t *testing.T) {
	tests := []struct {
		name       string
		postID     int
		authorName string
		content    string
		valid      bool
	}{
		{
			name:       "valid comment",
			postID:     1,
			authorName: "John Doe",
			content:    "Valid comment content with sufficient length.",
			valid:      true,
		},
		{
			name:       "author name with special characters",
			postID:     1,
			authorName: "John O'Connor-Smith Jr.",
			content:    "Valid comment content with sufficient length.",
			valid:      true,
		},
		{
			name:       "content with newlines and formatting",
			postID:     1,
			authorName: "John Doe",
			content:    "Comment with\nnewlines and\ttabs and proper length.",
			valid:      true,
		},
		{
			name:       "maximum length author name",
			postID:     1,
			authorName: string(make([]byte, 255)),
			content:    "Valid comment content with sufficient length.",
			valid:      true,
		},
		{
			name:       "maximum length content",
			postID:     1,
			authorName: "John Doe",
			content:    string(make([]byte, 1000)),
			valid:      true,
		},
		{
			name:       "minimum length content",
			postID:     1,
			authorName: "John Doe",
			content:    "Hi!",
			valid:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := comment.NewComment(tt.postID, tt.authorName, tt.content)

			if tt.valid {
				if err != nil {
					t.Errorf("expected valid comment, got error: %v", err)
				}
				if c == nil {
					t.Error("expected comment to be created")
				}
			} else {
				if err == nil {
					t.Error("expected error for invalid comment")
				}
			}
		})
	}
}

func TestComment_BelongsToPost(t *testing.T) {
	c, err := comment.NewComment(123, "John Doe", "Test content with sufficient length.")
	if err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}

	tests := []struct {
		name     string
		postID   int
		expected bool
	}{
		{
			name:     "correct post",
			postID:   123,
			expected: true,
		},
		{
			name:     "incorrect post",
			postID:   456,
			expected: false,
		},
		{
			name:     "zero post ID",
			postID:   0,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.BelongsToPost(tt.postID)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
