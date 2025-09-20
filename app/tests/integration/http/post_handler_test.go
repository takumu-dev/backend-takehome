package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"blog-platform/internal/domain/auth"
	"blog-platform/internal/domain/post"
	"blog-platform/internal/domain/user"
	"blog-platform/internal/infrastructure/http/handlers"
	"blog-platform/internal/infrastructure/http/middleware"
)

// MockPostService implements the post.Service interface for testing
type MockPostService struct {
	posts  map[int]*post.Post
	nextID int
}

func NewMockPostService() *MockPostService {
	return &MockPostService{
		posts:  make(map[int]*post.Post),
		nextID: 1,
	}
}

func (m *MockPostService) CreatePost(ctx context.Context, userID int, title, content string) (*post.Post, error) {
	p, err := post.NewPost(title, content, userID)
	if err != nil {
		return nil, err
	}
	p.ID = m.nextID
	m.nextID++
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	m.posts[p.ID] = p
	return p, nil
}

func (m *MockPostService) GetPost(ctx context.Context, id int) (*post.Post, error) {
	if p, exists := m.posts[id]; exists {
		return p, nil
	}
	return nil, post.ErrPostNotFound
}

func (m *MockPostService) GetPostsByAuthor(ctx context.Context, authorID int, limit, offset int) ([]*post.Post, error) {
	var result []*post.Post
	count := 0
	for _, p := range m.posts {
		if p.AuthorID == authorID {
			if count >= offset && len(result) < limit {
				result = append(result, p)
			}
			count++
		}
	}
	return result, nil
}

func (m *MockPostService) ListPosts(ctx context.Context, limit, offset int) ([]*post.Post, error) {
	var result []*post.Post
	count := 0
	for _, p := range m.posts {
		if count >= offset && len(result) < limit {
			result = append(result, p)
		}
		count++
		if len(result) >= limit {
			break
		}
	}
	return result, nil
}

func (m *MockPostService) UpdatePost(ctx context.Context, userID, postID int, title, content string) (*post.Post, error) {
	p, exists := m.posts[postID]
	if !exists {
		return nil, post.ErrPostNotFound
	}
	if p.AuthorID != userID {
		return nil, post.ErrUnauthorized
	}
	err := p.Update(title, content)
	if err != nil {
		return nil, err
	}
	p.UpdatedAt = time.Now()
	return p, nil
}

func (m *MockPostService) DeletePost(ctx context.Context, userID, postID int) error {
	p, exists := m.posts[postID]
	if !exists {
		return post.ErrPostNotFound
	}
	if p.AuthorID != userID {
		return post.ErrUnauthorized
	}
	delete(m.posts, postID)
	return nil
}

// MockUserService for authentication context
type MockUserService struct {
	users map[string]*user.User
}

func NewMockUserService() *MockUserService {
	return &MockUserService{
		users: make(map[string]*user.User),
	}
}

func (m *MockUserService) Register(ctx context.Context, name, email, password string) (*user.User, error) {
	u, err := user.NewUser(name, email, password)
	if err != nil {
		return nil, err
	}
	u.ID = len(m.users) + 1
	m.users[email] = u
	return u, nil
}

func (m *MockUserService) Login(ctx context.Context, email, password string) (*user.User, error) {
	u, exists := m.users[email]
	if !exists {
		return nil, user.ErrUserNotFound
	}
	if !u.ValidatePassword(password) {
		return nil, user.ErrInvalidCredentials
	}
	return u, nil
}

func (m *MockUserService) GetByID(ctx context.Context, id int) (*user.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, user.ErrUserNotFound
}

func (m *MockUserService) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	u, exists := m.users[email]
	if !exists {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

func (m *MockUserService) UpdateProfile(ctx context.Context, userID int, name, email string) (*user.User, error) {
	return nil, nil // Not needed for post tests
}

func (m *MockUserService) UpdatePassword(ctx context.Context, userID int, currentPassword, newPassword string) error {
	return nil // Not needed for post tests
}

func (m *MockUserService) List(ctx context.Context, limit, offset int) ([]*user.User, error) {
	return nil, nil // Not needed for post tests
}

// MockAuthService for authentication
type MockAuthService struct {
	userService *MockUserService
}

func NewMockAuthService(userService *MockUserService) *MockAuthService {
	return &MockAuthService{
		userService: userService,
	}
}

func (m *MockAuthService) GenerateToken(ctx context.Context, user *user.User) (string, error) {
	return "mock-jwt-token", nil
}

func (m *MockAuthService) ValidateToken(ctx context.Context, token string) (*auth.TokenClaims, error) {
	if token == "mock-jwt-token" {
		return &auth.TokenClaims{
			UserID: 1,
			Email:  "test@example.com",
		}, nil
	}
	return nil, auth.ErrInvalidToken
}

func (m *MockAuthService) Login(ctx context.Context, email, password string) (*user.User, string, error) {
	u, err := m.userService.Login(ctx, email, password)
	if err != nil {
		return nil, "", err
	}
	token, _ := m.GenerateToken(ctx, u)
	return u, token, nil
}

func (m *MockAuthService) Register(ctx context.Context, name, email, password string) (*user.User, string, error) {
	u, err := m.userService.Register(ctx, name, email, password)
	if err != nil {
		return nil, "", err
	}
	token, _ := m.GenerateToken(ctx, u)
	return u, token, nil
}

func (m *MockAuthService) RefreshToken(ctx context.Context, token string) (string, error) {
	return "mock-refreshed-token", nil
}

// MockLogger implements the service.Logger interface for testing
type MockLogger struct{}

func (m *MockLogger) Info(ctx context.Context, msg string, args ...any)  {}
func (m *MockLogger) Error(ctx context.Context, msg string, args ...any) {}
func (m *MockLogger) Warn(ctx context.Context, msg string, args ...any)  {}
func (m *MockLogger) Debug(ctx context.Context, msg string, args ...any) {}

func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func setupTestServer() (*echo.Echo, *handlers.PostHandler) {
	e := echo.New()
	e.Validator = middleware.NewValidator()
	
	postService := NewMockPostService()
	logger := NewMockLogger()
	
	postHandler := handlers.NewPostHandler(postService, logger)
	
	return e, postHandler
}

func setupAuthenticatedRequest(e *echo.Echo, method, path string, body []byte) (*httptest.ResponseRecorder, echo.Context) {
	req := httptest.NewRequest(method, path, bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	// Set authenticated user context
	c.Set("user_id", 1)
	
	return rec, c
}

func TestPostHandler_CreatePost_Success(t *testing.T) {
	e, postHandler := setupTestServer()
	
	createReq := handlers.CreatePostRequest{
		Title:   "Test Post",
		Content: "This is a test post content with more than 10 characters.",
	}
	
	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)
	
	rec, c := setupAuthenticatedRequest(e, http.MethodPost, "/api/v1/posts", reqBody)
	
	err = postHandler.CreatePost(c)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusCreated, rec.Code)
	
	var response handlers.PostResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "Test Post", response.Title)
	assert.Equal(t, "This is a test post content with more than 10 characters.", response.Content)
	assert.Equal(t, 1, response.AuthorID)
	assert.NotEmpty(t, response.CreatedAt)
	assert.NotEmpty(t, response.UpdatedAt)
}

func TestPostHandler_CreatePost_ValidationError(t *testing.T) {
	e, postHandler := setupTestServer()
	
	createReq := handlers.CreatePostRequest{
		Title:   "", // Invalid: empty title
		Content: "Short", // Invalid: too short content
	}
	
	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)
	
	rec, c := setupAuthenticatedRequest(e, http.MethodPost, "/api/v1/posts", reqBody)
	
	err = postHandler.CreatePost(c)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	
	var response handlers.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "validation_error", response.Error)
}

func TestPostHandler_CreatePost_Unauthorized(t *testing.T) {
	e, postHandler := setupTestServer()
	
	createReq := handlers.CreatePostRequest{
		Title:   "Test Post",
		Content: "This is a test post content with more than 10 characters.",
	}
	
	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)
	
	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts", bytes.NewBuffer(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// Don't set user_id in context to simulate unauthorized request
	
	err = postHandler.CreatePost(c)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestPostHandler_GetPost_Success(t *testing.T) {
	e, postHandler := setupTestServer()
	
	// First create a post
	createReq := handlers.CreatePostRequest{
		Title:   "Test Post",
		Content: "This is a test post content with more than 10 characters.",
	}
	
	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)
	
	rec, c := setupAuthenticatedRequest(e, http.MethodPost, "/api/v1/posts", reqBody)
	err = postHandler.CreatePost(c)
	require.NoError(t, err)
	
	var createResponse handlers.PostResponse
	err = json.Unmarshal(rec.Body.Bytes(), &createResponse)
	require.NoError(t, err)
	
	// Now get the post
	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts/"+strconv.Itoa(createResponse.ID), nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(createResponse.ID))
	
	err = postHandler.GetPost(c)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusOK, rec.Code)
	
	var response handlers.PostResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, createResponse.ID, response.ID)
	assert.Equal(t, "Test Post", response.Title)
	assert.Equal(t, "This is a test post content with more than 10 characters.", response.Content)
}

func TestPostHandler_GetPost_NotFound(t *testing.T) {
	e, postHandler := setupTestServer()
	
	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts/999", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("999")
	
	err := postHandler.GetPost(c)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestPostHandler_GetPost_InvalidID(t *testing.T) {
	e, postHandler := setupTestServer()
	
	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts/invalid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("invalid")
	
	err := postHandler.GetPost(c)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPostHandler_ListPosts_Success(t *testing.T) {
	e, postHandler := setupTestServer()
	
	// Create a few posts first
	for i := 1; i <= 3; i++ {
		createReq := handlers.CreatePostRequest{
			Title:   fmt.Sprintf("Test Post %d", i),
			Content: fmt.Sprintf("This is test post content %d with more than 10 characters.", i),
		}
		
		reqBody, err := json.Marshal(createReq)
		require.NoError(t, err)
		
		_, c := setupAuthenticatedRequest(e, http.MethodPost, "/api/v1/posts", reqBody)
		err = postHandler.CreatePost(c)
		require.NoError(t, err)
	}
	
	// Now list posts
	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts?limit=10&offset=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	err := postHandler.ListPosts(c)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusOK, rec.Code)
	
	var response handlers.PostListResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, 3, len(response.Posts))
	assert.Equal(t, 10, response.Limit)
	assert.Equal(t, 0, response.Offset)
}

func TestPostHandler_UpdatePost_Success(t *testing.T) {
	e, postHandler := setupTestServer()
	
	// First create a post
	createReq := handlers.CreatePostRequest{
		Title:   "Original Title",
		Content: "Original content with more than 10 characters.",
	}
	
	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)
	
	rec, c := setupAuthenticatedRequest(e, http.MethodPost, "/api/v1/posts", reqBody)
	err = postHandler.CreatePost(c)
	require.NoError(t, err)
	
	var createResponse handlers.PostResponse
	err = json.Unmarshal(rec.Body.Bytes(), &createResponse)
	require.NoError(t, err)
	
	// Now update the post
	updateReq := handlers.UpdatePostRequest{
		Title:   "Updated Title",
		Content: "Updated content with more than 10 characters.",
	}
	
	reqBody, err = json.Marshal(updateReq)
	require.NoError(t, err)
	
	req := httptest.NewRequest(http.MethodPut, "/api/v1/posts/"+strconv.Itoa(createResponse.ID), bytes.NewBuffer(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("user_id", 1) // Same user who created the post
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(createResponse.ID))
	
	err = postHandler.UpdatePost(c)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusOK, rec.Code)
	
	var response handlers.PostResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "Updated Title", response.Title)
	assert.Equal(t, "Updated content with more than 10 characters.", response.Content)
}

func TestPostHandler_UpdatePost_Forbidden(t *testing.T) {
	e, postHandler := setupTestServer()
	
	// First create a post
	createReq := handlers.CreatePostRequest{
		Title:   "Original Title",
		Content: "Original content with more than 10 characters.",
	}
	
	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)
	
	rec, c := setupAuthenticatedRequest(e, http.MethodPost, "/api/v1/posts", reqBody)
	err = postHandler.CreatePost(c)
	require.NoError(t, err)
	
	var createResponse handlers.PostResponse
	err = json.Unmarshal(rec.Body.Bytes(), &createResponse)
	require.NoError(t, err)
	
	// Try to update with different user
	updateReq := handlers.UpdatePostRequest{
		Title:   "Updated Title",
		Content: "Updated content with more than 10 characters.",
	}
	
	reqBody, err = json.Marshal(updateReq)
	require.NoError(t, err)
	
	req := httptest.NewRequest(http.MethodPut, "/api/v1/posts/"+strconv.Itoa(createResponse.ID), bytes.NewBuffer(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("user_id", 999) // Different user
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(createResponse.ID))
	
	err = postHandler.UpdatePost(c)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestPostHandler_DeletePost_Success(t *testing.T) {
	e, postHandler := setupTestServer()
	
	// First create a post
	createReq := handlers.CreatePostRequest{
		Title:   "Test Post",
		Content: "This is a test post content with more than 10 characters.",
	}
	
	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)
	
	rec, c := setupAuthenticatedRequest(e, http.MethodPost, "/api/v1/posts", reqBody)
	err = postHandler.CreatePost(c)
	require.NoError(t, err)
	
	var createResponse handlers.PostResponse
	err = json.Unmarshal(rec.Body.Bytes(), &createResponse)
	require.NoError(t, err)
	
	// Now delete the post
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/posts/"+strconv.Itoa(createResponse.ID), nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("user_id", 1) // Same user who created the post
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(createResponse.ID))
	
	err = postHandler.DeletePost(c)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestPostHandler_DeletePost_Forbidden(t *testing.T) {
	e, postHandler := setupTestServer()
	
	// First create a post
	createReq := handlers.CreatePostRequest{
		Title:   "Test Post",
		Content: "This is a test post content with more than 10 characters.",
	}
	
	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)
	
	rec, c := setupAuthenticatedRequest(e, http.MethodPost, "/api/v1/posts", reqBody)
	err = postHandler.CreatePost(c)
	require.NoError(t, err)
	
	var createResponse handlers.PostResponse
	err = json.Unmarshal(rec.Body.Bytes(), &createResponse)
	require.NoError(t, err)
	
	// Try to delete with different user
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/posts/"+strconv.Itoa(createResponse.ID), nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("user_id", 999) // Different user
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(createResponse.ID))
	
	err = postHandler.DeletePost(c)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestPostHandler_DeletePost_NotFound(t *testing.T) {
	e, postHandler := setupTestServer()
	
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/posts/999", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", 1)
	c.SetParamNames("id")
	c.SetParamValues("999")
	
	err := postHandler.DeletePost(c)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusNotFound, rec.Code)
}
