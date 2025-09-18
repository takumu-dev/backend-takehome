package post

import (
	"context"
)

// Repository defines the interface for post data access
type Repository interface {
	Create(ctx context.Context, post *Post) error
	GetByID(ctx context.Context, id int) (*Post, error)
	GetAll(ctx context.Context, limit, offset int) ([]*Post, error)
	Update(ctx context.Context, post *Post) error
	Delete(ctx context.Context, id int) error
	GetByAuthorID(ctx context.Context, authorID int) ([]*Post, error)
}
