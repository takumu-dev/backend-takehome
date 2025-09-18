package post

import (
	"time"
	"blog-platform/internal/domain/user"
)

// Post represents a blog post entity in the domain
type Post struct {
	ID        int        `json:"id" db:"id"`
	Title     string     `json:"title" db:"title"`
	Content   string     `json:"content" db:"content"`
	AuthorID  int        `json:"author_id" db:"author_id"`
	Author    *user.User `json:"author,omitempty"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// NewPost creates a new post instance
func NewPost(title, content string, authorID int) *Post {
	now := time.Now()
	return &Post{
		Title:     title,
		Content:   content,
		AuthorID:  authorID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Update updates the post's title and content
func (p *Post) Update(title, content string) {
	p.Title = title
	p.Content = content
	p.UpdatedAt = time.Now()
}

// IsOwnedBy checks if the post is owned by the given user ID
func (p *Post) IsOwnedBy(userID int) bool {
	return p.AuthorID == userID
}
