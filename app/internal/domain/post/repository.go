package post

import (
	"context"
	"errors"
)

// Repository errors
var (
	ErrPostNotFound     = errors.New("post not found")
	ErrInvalidPostData  = errors.New("invalid post data")
	ErrUnauthorized     = errors.New("unauthorized access to post")
)

// Repository defines the interface for post data access
type Repository interface {
	Create(ctx context.Context, post *Post) error
	GetByID(ctx context.Context, id int) (*Post, error)
	GetByAuthorID(ctx context.Context, authorID int, limit, offset int) ([]*Post, error)
	List(ctx context.Context, limit, offset int) ([]*Post, error)
	Update(ctx context.Context, post *Post) error
	Delete(ctx context.Context, id int) error
}
