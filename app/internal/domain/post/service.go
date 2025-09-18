package post

import (
	"context"
)

// Service defines the interface for post business logic
type Service interface {
	CreatePost(ctx context.Context, userID int, title, content string) (*Post, error)
	GetPost(ctx context.Context, id int) (*Post, error)
	GetAllPosts(ctx context.Context, limit, offset int) ([]*Post, error)
	UpdatePost(ctx context.Context, userID, postID int, title, content string) (*Post, error)
	DeletePost(ctx context.Context, userID, postID int) error
}
