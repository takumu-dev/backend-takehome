package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"blog-platform/internal/application/service"
	"blog-platform/internal/domain/post"
	"blog-platform/internal/infrastructure/http/errors"
	"blog-platform/internal/infrastructure/http/middleware"
)

// PostHandler handles post-related HTTP requests
type PostHandler struct {
	postService post.Service
	logger      service.Logger
}

// NewPostHandler creates a new post handler
func NewPostHandler(postService post.Service, logger service.Logger) *PostHandler {
	return &PostHandler{
		postService: postService,
		logger:      logger,
	}
}

// CreatePostRequest represents the create post request payload
type CreatePostRequest struct {
	Title   string `json:"title" validate:"required,min=1,max=500,no_html,safe_string"`
	Content string `json:"content" validate:"required,min=10,max=10000,no_html"`
}

// UpdatePostRequest represents the update post request payload
type UpdatePostRequest struct {
	Title   string `json:"title" validate:"required,min=1,max=500,no_html,safe_string"`
	Content string `json:"content" validate:"required,min=10,max=10000,no_html"`
}

// PostResponse represents the post data in responses
type PostResponse struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	AuthorID  int    `json:"author_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// PostListResponse represents the paginated post list response
type PostListResponse struct {
	Posts  []PostResponse `json:"posts"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// CreatePost handles POST /api/v1/posts
// @Summary Create a new post
// @Description Create a new blog post
// @Tags posts
// @Accept json
// @Produce json
// @Param request body CreatePostRequest true "Post creation data"
// @Success 201 {object} PostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/posts [post]
func (h *PostHandler) CreatePost(c echo.Context) error {
	ctx := c.Request().Context()
	
	// Get user ID from context (set by auth middleware)
	userID, ok := c.Get("user_id").(int)
	if !ok {
		h.logger.Error(ctx, "user_id not found in context")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User authentication required",
		})
	}

	// Parse and validate request
	var req CreatePostRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error(ctx, "failed to bind create post request", "error", err.Error())
		return errors.HandleError(c, errors.ErrInvalidRequest)
	}

	// Sanitize input
	req.Title = middleware.SanitizeInput(req.Title)
	req.Content = middleware.SanitizeInput(req.Content)

	if err := c.Validate(req); err != nil {
		h.logger.Error(ctx, "create post request validation failed", "error", err.Error())
		return errors.HandleError(c, err)
	}

	// Create post
	createdPost, err := h.postService.CreatePost(ctx, userID, req.Title, req.Content)
	if err != nil {
		h.logger.Error(ctx, "failed to create post", "userID", userID, "error", err.Error())
		return errors.HandleError(c, err)
	}

	// Convert to response format
	response := PostResponse{
		ID:        createdPost.ID,
		Title:     createdPost.Title,
		Content:   createdPost.Content,
		AuthorID:  createdPost.AuthorID,
		CreatedAt: createdPost.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: createdPost.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	h.logger.Info(ctx, "post created successfully", "postID", createdPost.ID, "userID", userID)
	return c.JSON(http.StatusCreated, response)
}

// GetPost handles GET /api/v1/posts/{id}
// @Summary Get a post by ID
// @Description Retrieve a specific blog post by its ID
// @Tags posts
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} PostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/posts/{id} [get]
func (h *PostHandler) GetPost(c echo.Context) error {
	ctx := c.Request().Context()
	
	// Parse post ID
	postIDStr := c.Param("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		h.logger.Error(ctx, "invalid post ID", "postID", postIDStr)
		return errors.HandleError(c, errors.ErrInvalidRequest)
	}

	// Get post
	retrievedPost, err := h.postService.GetPost(ctx, postID)
	if err != nil {
		h.logger.Error(ctx, "failed to get post", "postID", postID, "error", err.Error())
		return errors.HandleError(c, err)
	}

	// Convert to response format
	response := PostResponse{
		ID:        retrievedPost.ID,
		Title:     retrievedPost.Title,
		Content:   retrievedPost.Content,
		AuthorID:  retrievedPost.AuthorID,
		CreatedAt: retrievedPost.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: retrievedPost.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return c.JSON(http.StatusOK, response)
}

// ListPosts handles GET /api/v1/posts
// @Summary List posts with pagination
// @Description Retrieve a paginated list of blog posts
// @Tags posts
// @Produce json
// @Param limit query int false "Number of posts to return (default: 10, max: 100)"
// @Param offset query int false "Number of posts to skip (default: 0)"
// @Success 200 {object} PostListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/posts [get]
func (h *PostHandler) ListPosts(c echo.Context) error {
	ctx := c.Request().Context()
	
	// Parse pagination parameters
	limit := 10 // default
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0 // default
	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get posts
	posts, err := h.postService.ListPosts(ctx, limit, offset)
	if err != nil {
		h.logger.Error(ctx, "failed to list posts", "limit", limit, "offset", offset, "error", err.Error())
		return errors.HandleError(c, err)
	}

	// Convert to response format
	postResponses := make([]PostResponse, len(posts))
	for i, p := range posts {
		postResponses[i] = PostResponse{
			ID:        p.ID,
			Title:     p.Title,
			Content:   p.Content,
			AuthorID:  p.AuthorID,
			CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	response := PostListResponse{
		Posts:  postResponses,
		Total:  len(postResponses), // Note: This is simplified - in production you'd want actual total count
		Limit:  limit,
		Offset: offset,
	}

	return c.JSON(http.StatusOK, response)
}

// UpdatePost handles PUT /api/v1/posts/{id}
// @Summary Update a post
// @Description Update an existing blog post (only by author)
// @Tags posts
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Param request body UpdatePostRequest true "Post update data"
// @Success 200 {object} PostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/posts/{id} [put]
func (h *PostHandler) UpdatePost(c echo.Context) error {
	ctx := c.Request().Context()
	
	// Get user ID from context (set by auth middleware)
	userID, ok := c.Get("user_id").(int)
	if !ok {
		h.logger.Error(ctx, "user_id not found in context")
		return errors.HandleError(c, errors.ErrUnauthorized)
	}

	// Parse post ID
	postIDStr := c.Param("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		h.logger.Error(ctx, "invalid post ID", "postID", postIDStr)
		return errors.HandleError(c, errors.ErrInvalidRequest)
	}

	// Parse and validate request
	var req UpdatePostRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error(ctx, "failed to bind update post request", "error", err.Error())
		return errors.HandleError(c, errors.ErrInvalidRequest)
	}

	// Sanitize input
	req.Title = middleware.SanitizeInput(req.Title)
	req.Content = middleware.SanitizeInput(req.Content)

	if err := c.Validate(req); err != nil {
		h.logger.Error(ctx, "update post request validation failed", "error", err.Error())
		return errors.HandleError(c, err)
	}

	// Update post
	updatedPost, err := h.postService.UpdatePost(ctx, userID, postID, req.Title, req.Content)
	if err != nil {
		h.logger.Error(ctx, "failed to update post", "userID", userID, "postID", postID, "error", err.Error())
		return errors.HandleError(c, err)
	}

	// Convert to response format
	response := PostResponse{
		ID:        updatedPost.ID,
		Title:     updatedPost.Title,
		Content:   updatedPost.Content,
		AuthorID:  updatedPost.AuthorID,
		CreatedAt: updatedPost.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: updatedPost.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	h.logger.Info(ctx, "post updated successfully", "postID", postID, "userID", userID)
	return c.JSON(http.StatusOK, response)
}

// DeletePost handles DELETE /api/v1/posts/{id}
// @Summary Delete a post
// @Description Delete an existing blog post (only by author)
// @Tags posts
// @Param id path int true "Post ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/posts/{id} [delete]
func (h *PostHandler) DeletePost(c echo.Context) error {
	ctx := c.Request().Context()
	
	// Get user ID from context (set by auth middleware)
	userID, ok := c.Get("user_id").(int)
	if !ok {
		h.logger.Error(ctx, "user_id not found in context")
		return errors.HandleError(c, errors.ErrUnauthorized)
	}

	// Parse post ID
	postIDStr := c.Param("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		h.logger.Error(ctx, "invalid post ID", "postID", postIDStr)
		return errors.HandleError(c, errors.ErrInvalidRequest)
	}

	// Delete post
	err = h.postService.DeletePost(ctx, userID, postID)
	if err != nil {
		h.logger.Error(ctx, "failed to delete post", "userID", userID, "postID", postID, "error", err.Error())
		return errors.HandleError(c, err)
	}

	h.logger.Info(ctx, "post deleted successfully", "postID", postID, "userID", userID)
	return c.NoContent(http.StatusNoContent)
}
