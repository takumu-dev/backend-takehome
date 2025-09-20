package post

import (
	"errors"
	"strings"
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

// NewPost creates a new post instance with validation
func NewPost(title, content string, authorID int) (*Post, error) {
	if err := validateTitle(title); err != nil {
		return nil, err
	}

	if err := validateContent(content); err != nil {
		return nil, err
	}

	if authorID <= 0 {
		return nil, errors.New("author ID must be positive")
	}

	now := time.Now()
	return &Post{
		Title:     strings.TrimSpace(title),
		Content:   strings.TrimSpace(content),
		AuthorID:  authorID,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Update updates the post's title and content with validation
func (p *Post) Update(title, content string) error {
	if err := validateTitle(title); err != nil {
		return err
	}

	if err := validateContent(content); err != nil {
		return err
	}

	p.Title = strings.TrimSpace(title)
	p.Content = strings.TrimSpace(content)
	p.UpdatedAt = time.Now()
	return nil
}

// IsAuthor checks if the post is authored by the given user ID
func (p *Post) IsAuthor(userID int) bool {
	return p.AuthorID == userID
}

// IsOwnedBy checks if the post is owned by the given user ID (alias for IsAuthor)
func (p *Post) IsOwnedBy(userID int) bool {
	return p.IsAuthor(userID)
}

// validateTitle validates the post title
func validateTitle(title string) error {
	title = strings.TrimSpace(title)
	
	if title == "" {
		return errors.New("title cannot be empty")
	}

	if len(title) > 500 {
		return errors.New("title cannot exceed 500 characters")
	}

	return nil
}

// validateContent validates the post content
func validateContent(content string) error {
	content = strings.TrimSpace(content)
	
	if content == "" {
		return errors.New("content cannot be empty")
	}

	if len(content) < 10 {
		return errors.New("content must be at least 10 characters")
	}

	if len(content) > 10000 {
		return errors.New("content cannot exceed 10000 characters")
	}

	return nil
}
