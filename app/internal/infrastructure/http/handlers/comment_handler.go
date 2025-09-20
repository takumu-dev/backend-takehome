package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"blog-platform/internal/application/service"
	"blog-platform/internal/domain/comment"
	"blog-platform/internal/infrastructure/http/errors"
	"blog-platform/internal/infrastructure/http/middleware"
)

// CommentHandler handles HTTP requests for comment operations
type CommentHandler struct {
	commentService comment.Service
	logger         service.Logger
}

// NewCommentHandler creates a new comment handler
func NewCommentHandler(commentService comment.Service, logger service.Logger) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
		logger:         logger,
	}
}

// CreateCommentRequest represents the request body for creating a comment
type CreateCommentRequest struct {
	AuthorName string `json:"author_name" validate:"required,max=255"`
	Content    string `json:"content" validate:"required,min=3,max=1000"`
}

// CommentResponse represents a comment in API responses
type CommentResponse struct {
	ID         int    `json:"id"`
	PostID     int    `json:"post_id"`
	AuthorName string `json:"author_name"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
}

// CommentListResponse represents the response for listing comments
type CommentListResponse struct {
	Comments []CommentResponse `json:"comments"`
	Total    int               `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

// CreateComment handles POST /api/v1/posts/{id}/comments
// @Summary Create a new comment
// @Description Create a new comment for a specific post
// @Tags comments
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Param comment body CreateCommentRequest true "Comment data"
// @Success 201 {object} CommentResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/posts/{id}/comments [post]
func (h *CommentHandler) CreateComment(c echo.Context) error {
	ctx := c.Request().Context()
	
	// Get post ID from path parameter
	postIDStr := c.Param("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		h.logger.Warn(ctx, "Invalid post ID in path", "post_id", postIDStr, "error", err.Error())
		return errors.HandleError(c, errors.ErrInvalidRequest)
	}
	
	// Parse and validate request
	var req CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn(ctx, "Failed to bind comment request", "error", err.Error())
		return errors.HandleError(c, errors.ErrInvalidRequest)
	}
	
	// Sanitize input
	req.AuthorName = middleware.SanitizeInput(req.AuthorName)
	req.Content = middleware.SanitizeInput(req.Content)
	
	if err := c.Validate(&req); err != nil {
		h.logger.Warn(ctx, "Comment validation failed", "error", err.Error())
		return errors.HandleError(c, err)
	}
	
	h.logger.Info(ctx, "Creating comment", "post_id", postID, "author_name", req.AuthorName)
	
	// Create comment
	createdComment, err := h.commentService.AddComment(ctx, postID, req.AuthorName, req.Content)
	if err != nil {
		h.logger.Error(ctx, "Failed to create comment", "error", err.Error(), "post_id", postID)
		return errors.HandleError(c, err)
	}
	
	response := CommentResponse{
		ID:         createdComment.ID,
		PostID:     createdComment.PostID,
		AuthorName: createdComment.AuthorName,
		Content:    createdComment.Content,
		CreatedAt:  createdComment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	
	h.logger.Info(ctx, "Comment created successfully", "comment_id", createdComment.ID, "post_id", postID)
	return c.JSON(http.StatusCreated, response)
}

// GetCommentsByPost handles GET /api/v1/posts/{id}/comments
// @Summary Get comments for a post
// @Description Get all comments for a specific post with pagination
// @Tags comments
// @Produce json
// @Param id path int true "Post ID"
// @Param limit query int false "Number of comments to return (default: 10, max: 100)"
// @Param offset query int false "Number of comments to skip (default: 0)"
// @Success 200 {object} CommentListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/posts/{id}/comments [get]
func (h *CommentHandler) GetCommentsByPost(c echo.Context) error {
	ctx := c.Request().Context()
	
	// Get post ID from path parameter
	postIDStr := c.Param("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		h.logger.Warn(ctx, "Invalid post ID in path", "post_id", postIDStr, "error", err.Error())
		return errors.HandleError(c, errors.ErrInvalidRequest)
	}
	
	// Parse pagination parameters
	limit := 10 // default
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}
	
	offset := 0 // default
	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}
	
	h.logger.Info(ctx, "Getting comments for post", "post_id", postID, "limit", limit, "offset", offset)
	
	// Get comments
	comments, err := h.commentService.GetCommentsByPost(ctx, postID, limit, offset)
	if err != nil {
		h.logger.Error(ctx, "Failed to get comments", "error", err.Error(), "post_id", postID)
		return errors.HandleError(c, err)
	}
	
	// Convert to response format
	commentResponses := make([]CommentResponse, len(comments))
	for i, comment := range comments {
		commentResponses[i] = CommentResponse{
			ID:         comment.ID,
			PostID:     comment.PostID,
			AuthorName: comment.AuthorName,
			Content:    comment.Content,
			CreatedAt:  comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	
	response := CommentListResponse{
		Comments: commentResponses,
		Total:    len(commentResponses),
		Limit:    limit,
		Offset:   offset,
	}
	
	h.logger.Info(ctx, "Comments retrieved successfully", "post_id", postID, "count", len(comments))
	return c.JSON(http.StatusOK, response)
}
