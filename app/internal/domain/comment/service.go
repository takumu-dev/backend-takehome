package comment

import (
	"context"
)

// Service defines the interface for comment business logic
type Service interface {
	AddComment(ctx context.Context, postID int, authorName, content string) (*Comment, error)
	GetCommentsByPost(ctx context.Context, postID int) ([]*Comment, error)
}
