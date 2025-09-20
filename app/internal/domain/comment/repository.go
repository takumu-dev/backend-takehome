package comment

import (
	"context"
	"errors"
)

var (
	// ErrCommentNotFound is returned when a comment is not found
	ErrCommentNotFound = errors.New("comment not found")
)

// Repository defines the interface for comment data access
type Repository interface {
	Create(ctx context.Context, comment *Comment) error
	GetByID(ctx context.Context, id int) (*Comment, error)
	GetByPostID(ctx context.Context, postID int, limit, offset int) ([]*Comment, error)
	Update(ctx context.Context, comment *Comment) error
	Delete(ctx context.Context, id int) error
}
