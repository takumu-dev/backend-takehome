package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"blog-platform/internal/domain/comment"
	"blog-platform/internal/infrastructure/http/handlers"
	"blog-platform/internal/infrastructure/http/middleware"
)

// MockCommentService implements the comment.Service interface for testing
type MockCommentService struct {
	comments map[int]*comment.Comment
	nextID   int
}

func NewMockCommentService() *MockCommentService {
	return &MockCommentService{
		comments: make(map[int]*comment.Comment),
		nextID:   1,
	}
}

func (m *MockCommentService) AddComment(ctx context.Context, postID int, authorName, content string) (*comment.Comment, error) {
	newComment, err := comment.NewComment(postID, authorName, content)
	if err != nil {
		return nil, err
	}
	
	newComment.ID = m.nextID
	m.comments[m.nextID] = newComment
	m.nextID++
	
	return newComment, nil
}

func (m *MockCommentService) GetComment(ctx context.Context, id int) (*comment.Comment, error) {
	if comment, exists := m.comments[id]; exists {
		return comment, nil
	}
	return nil, fmt.Errorf("comment not found")
}

func (m *MockCommentService) GetCommentsByPost(ctx context.Context, postID int, limit, offset int) ([]*comment.Comment, error) {
	var result []*comment.Comment
	count := 0
	
	for _, c := range m.comments {
		if c.PostID == postID {
			if count >= offset && len(result) < limit {
				result = append(result, c)
			}
			count++
		}
	}
	
	return result, nil
}

func (m *MockCommentService) UpdateComment(ctx context.Context, id int, authorName, content string) (*comment.Comment, error) {
	if c, exists := m.comments[id]; exists {
		if c.AuthorName != authorName {
			return nil, fmt.Errorf("unauthorized")
		}
		err := c.Update(content)
		if err != nil {
			return nil, err
		}
		return c, nil
	}
	return nil, fmt.Errorf("comment not found")
}

func (m *MockCommentService) DeleteComment(ctx context.Context, id int, authorName string) error {
	if c, exists := m.comments[id]; exists {
		if c.AuthorName != authorName {
			return fmt.Errorf("unauthorized")
		}
		delete(m.comments, id)
		return nil
	}
	return fmt.Errorf("comment not found")
}

func setupCommentTestServer() (*echo.Echo, *handlers.CommentHandler) {
	e := echo.New()
	e.Validator = middleware.NewValidator()
	
	commentService := NewMockCommentService()
	logger := NewMockLogger()
	
	commentHandler := handlers.NewCommentHandler(commentService, logger)
	
	return e, commentHandler
}

func TestCommentHandler_CreateComment_Success(t *testing.T) {
	e, commentHandler := setupCommentTestServer()
	
	createReq := handlers.CreateCommentRequest{
		AuthorName: "John Doe",
		Content:    "This is a test comment with sufficient content.",
	}
	
	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)
	
	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts/1/comments", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/posts/:id/comments")
	c.SetParamNames("id")
	c.SetParamValues("1")
	
	err = commentHandler.CreateComment(c)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusCreated, rec.Code)
	
	var response handlers.CommentResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, 1, response.ID)
	assert.Equal(t, 1, response.PostID)
	assert.Equal(t, "John Doe", response.AuthorName)
	assert.Equal(t, "This is a test comment with sufficient content.", response.Content)
	assert.NotEmpty(t, response.CreatedAt)
}

func TestCommentHandler_CreateComment_ValidationError(t *testing.T) {
	e, commentHandler := setupCommentTestServer()
	
	testCases := []struct {
		name    string
		request handlers.CreateCommentRequest
	}{
		{
			name: "empty author name",
			request: handlers.CreateCommentRequest{
				AuthorName: "",
				Content:    "Valid content here",
			},
		},
		{
			name: "empty content",
			request: handlers.CreateCommentRequest{
				AuthorName: "John Doe",
				Content:    "",
			},
		},
		{
			name: "content too short",
			request: handlers.CreateCommentRequest{
				AuthorName: "John Doe",
				Content:    "Hi",
			},
		},
		{
			name: "author name too long",
			request: handlers.CreateCommentRequest{
				AuthorName: string(make([]byte, 256)), // 256 characters, exceeds 255 limit
				Content:    "Valid content here",
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tc.request)
			require.NoError(t, err)
			
			req := httptest.NewRequest(http.MethodPost, "/api/v1/posts/1/comments", bytes.NewReader(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/api/v1/posts/:id/comments")
			c.SetParamNames("id")
			c.SetParamValues("1")
			
			err = commentHandler.CreateComment(c)
			require.NoError(t, err)
			
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})
	}
}

func TestCommentHandler_CreateComment_InvalidPostID(t *testing.T) {
	e, commentHandler := setupCommentTestServer()
	
	createReq := handlers.CreateCommentRequest{
		AuthorName: "John Doe",
		Content:    "This is a test comment.",
	}
	
	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)
	
	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts/invalid/comments", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/posts/:id/comments")
	c.SetParamNames("id")
	c.SetParamValues("invalid")
	
	err = commentHandler.CreateComment(c)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCommentHandler_GetCommentsByPost_Success(t *testing.T) {
	e, commentHandler := setupCommentTestServer()
	
	// Create a few comments first
	for i := 1; i <= 3; i++ {
		createReq := handlers.CreateCommentRequest{
			AuthorName: fmt.Sprintf("Author %d", i),
			Content:    fmt.Sprintf("This is test comment %d with sufficient content.", i),
		}
		
		reqBody, err := json.Marshal(createReq)
		require.NoError(t, err)
		
		req := httptest.NewRequest(http.MethodPost, "/api/v1/posts/1/comments", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/posts/:id/comments")
		c.SetParamNames("id")
		c.SetParamValues("1")
		
		err = commentHandler.CreateComment(c)
		require.NoError(t, err)
	}
	
	// Now get comments
	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts/1/comments?limit=10&offset=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/posts/:id/comments")
	c.SetParamNames("id")
	c.SetParamValues("1")
	
	err := commentHandler.GetCommentsByPost(c)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusOK, rec.Code)
	
	var response handlers.CommentListResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, 3, response.Total)
	assert.Equal(t, 10, response.Limit)
	assert.Equal(t, 0, response.Offset)
	assert.Len(t, response.Comments, 3)
	
	// Verify comment details
	for i, comment := range response.Comments {
		assert.Equal(t, 1, comment.PostID)
		assert.Equal(t, fmt.Sprintf("Author %d", i+1), comment.AuthorName)
		assert.Contains(t, comment.Content, fmt.Sprintf("test comment %d", i+1))
		assert.NotEmpty(t, comment.CreatedAt)
	}
}

func TestCommentHandler_GetCommentsByPost_EmptyResult(t *testing.T) {
	e, commentHandler := setupCommentTestServer()
	
	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts/999/comments", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/posts/:id/comments")
	c.SetParamNames("id")
	c.SetParamValues("999")
	
	err := commentHandler.GetCommentsByPost(c)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusOK, rec.Code)
	
	var response handlers.CommentListResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, 0, response.Total)
	assert.Equal(t, 10, response.Limit)
	assert.Equal(t, 0, response.Offset)
	assert.Len(t, response.Comments, 0)
}

func TestCommentHandler_GetCommentsByPost_InvalidPostID(t *testing.T) {
	e, commentHandler := setupCommentTestServer()
	
	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts/invalid/comments", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/posts/:id/comments")
	c.SetParamNames("id")
	c.SetParamValues("invalid")
	
	err := commentHandler.GetCommentsByPost(c)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCommentHandler_GetCommentsByPost_Pagination(t *testing.T) {
	e, commentHandler := setupCommentTestServer()
	
	// Create 5 comments
	for i := 1; i <= 5; i++ {
		createReq := handlers.CreateCommentRequest{
			AuthorName: fmt.Sprintf("Author %d", i),
			Content:    fmt.Sprintf("This is test comment %d with sufficient content.", i),
		}
		
		reqBody, err := json.Marshal(createReq)
		require.NoError(t, err)
		
		req := httptest.NewRequest(http.MethodPost, "/api/v1/posts/1/comments", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/posts/:id/comments")
		c.SetParamNames("id")
		c.SetParamValues("1")
		
		err = commentHandler.CreateComment(c)
		require.NoError(t, err)
	}
	
	// Test pagination: limit=2, offset=1
	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts/1/comments?limit=2&offset=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/posts/:id/comments")
	c.SetParamNames("id")
	c.SetParamValues("1")
	
	err := commentHandler.GetCommentsByPost(c)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusOK, rec.Code)
	
	var response handlers.CommentListResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, 2, response.Total) // Should return 2 comments due to limit
	assert.Equal(t, 2, response.Limit)
	assert.Equal(t, 1, response.Offset)
	assert.Len(t, response.Comments, 2)
}
