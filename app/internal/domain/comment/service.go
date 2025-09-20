package comment

import (
	"context"
)

// Service defines the interface for comment business logic
type Service interface {
	AddComment(ctx context.Context, postID int, authorName, content string) (*Comment, error)
	GetComment(ctx context.Context, id int) (*Comment, error)
	GetCommentsByPost(ctx context.Context, postID int, limit, offset int) ([]*Comment, error)
	UpdateComment(ctx context.Context, id int, authorName, content string) (*Comment, error)
	DeleteComment(ctx context.Context, id int, authorName string) error
}
