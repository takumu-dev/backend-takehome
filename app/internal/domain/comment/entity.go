package comment

import (
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
func NewComment(postID int, authorName, content string) *Comment {
	return &Comment{
		PostID:     postID,
		AuthorName: authorName,
		Content:    content,
		CreatedAt:  time.Now(),
	}
}
