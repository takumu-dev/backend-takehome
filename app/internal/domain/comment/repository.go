package comment

import (
	"context"
)

// Repository defines the interface for comment data access
type Repository interface {
	Create(ctx context.Context, comment *Comment) error
	GetByPostID(ctx context.Context, postID int) ([]*Comment, error)
	Delete(ctx context.Context, id int) error
}
