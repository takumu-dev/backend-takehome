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

// NewComment creates a new comment instance
func NewComment(postID int, authorName, content string) (*Comment, error) {
	// Validate post ID
	if postID <= 0 {
		return nil, errors.New("post ID must be positive")
	}

	return &Comment{
		PostID:     postID,
		AuthorName: strings.TrimSpace(authorName),
		Content:    strings.TrimSpace(content),
		CreatedAt:  time.Now(),
	}, nil
}

// Update updates the comment content
func (c *Comment) Update(content string) error {
	c.Content = strings.TrimSpace(content)
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
