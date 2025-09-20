package comment

import (
	"errors"
	"strings"
	"time"
)

// Comment represents a comment entity in the domain
type Comment struct {
	ID         int       `json:"id" db:"id"`
	PostID     int       `json:"post_id" db:"post_id"`
	AuthorName string    `json:"author_name" db:"author_name"`
	Content    string    `json:"content" db:"content"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// NewComment creates a new comment instance with validation
func NewComment(postID int, authorName, content string) (*Comment, error) {
	// Validate post ID
	if postID <= 0 {
		return nil, errors.New("post ID must be positive")
	}

	// Validate and sanitize author name
	authorName = strings.TrimSpace(authorName)
	if authorName == "" {
		return nil, errors.New("author name cannot be empty")
	}
	if len(authorName) > 255 {
		return nil, errors.New("author name cannot exceed 255 characters")
	}

	// Validate and sanitize content
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("content cannot be empty")
	}
	if len(content) < 3 {
		return nil, errors.New("content must be at least 3 characters")
	}
	if len(content) > 1000 {
		return nil, errors.New("content cannot exceed 1000 characters")
	}

	return &Comment{
		PostID:     postID,
		AuthorName: authorName,
		Content:    content,
		CreatedAt:  time.Now(),
	}, nil
}

// Update updates the comment content with validation
func (c *Comment) Update(content string) error {
	// Validate and sanitize content
	content = strings.TrimSpace(content)
	if content == "" {
		return errors.New("content cannot be empty")
	}
	if len(content) < 3 {
		return errors.New("content must be at least 3 characters")
	}
	if len(content) > 1000 {
		return errors.New("content cannot exceed 1000 characters")
	}

	c.Content = content
	return nil
}

// IsAuthor checks if the given author name matches the comment's author
func (c *Comment) IsAuthor(authorName string) bool {
	return c.AuthorName == authorName
}

// BelongsToPost checks if the comment belongs to the specified post
func (c *Comment) BelongsToPost(postID int) bool {
	return c.PostID == postID
}
