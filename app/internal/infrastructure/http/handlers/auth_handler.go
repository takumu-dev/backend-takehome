package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"blog-platform/internal/application/service"
	"blog-platform/internal/domain/auth"
	"blog-platform/internal/domain/user"
	"blog-platform/internal/infrastructure/http/errors"
	"blog-platform/internal/infrastructure/http/middleware"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	userService user.Service
	authService auth.AuthService
	logger      service.Logger
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(userService user.Service, authService auth.AuthService, logger service.Logger) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		authService: authService,
		logger:      logger,
	}
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=1,max=255,no_html,safe_string"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,strong_password"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

// UserResponse represents the user data in responses
type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Register handles user registration
// @Summary Register a new user
// @Description Register a new user account with name, email, and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "User registration data"
// @Success 201 {object} AuthResponse "User successfully registered"
// @Failure 400 {object} ErrorResponse "Invalid request data or validation error"
// @Failure 409 {object} ErrorResponse "User already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c echo.Context) error {
	ctx := c.Request().Context()
	h.logger.Info(ctx, "registration request received")

	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error(ctx, "failed to bind registration request", "error", err.Error())
		return errors.HandleError(c, errors.ErrInvalidRequest)
	}
	
	// Sanitize input
	req.Name = middleware.SanitizeInput(req.Name)
	req.Email = middleware.SanitizeInput(req.Email)

	if err := c.Validate(&req); err != nil {
		h.logger.Error(ctx, "registration request validation failed", "error", err.Error())
		return errors.HandleError(c, err)
	}

	// Register user
	registeredUser, token, err := h.authService.Register(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		h.logger.Error(ctx, "failed to register user", "error", err.Error(), "email", req.Email)
		return errors.HandleError(c, err)
	}

	response := AuthResponse{
		User: UserResponse{
			ID:    registeredUser.ID,
			Name:  registeredUser.Name,
			Email: registeredUser.Email,
		},
		Token: token,
	}

	h.logger.Info(ctx, "user registered successfully", "userID", registeredUser.ID, "email", registeredUser.Email)
	return c.JSON(http.StatusCreated, response)
}

// Login handles user authentication
// @Summary Login user
// @Description Authenticate user with email and password, returns JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "User login credentials"
// @Success 200 {object} AuthResponse "User successfully authenticated"
// @Failure 400 {object} ErrorResponse "Invalid request data or validation error"
// @Failure 401 {object} ErrorResponse "Invalid credentials"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()
	h.logger.Info(ctx, "login request received")

	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error(ctx, "failed to bind login request", "error", err.Error())
		return errors.HandleError(c, errors.ErrInvalidRequest)
	}
	
	// Sanitize input
	req.Email = middleware.SanitizeInput(req.Email)

	if err := c.Validate(&req); err != nil {
		h.logger.Error(ctx, "login request validation failed", "error", err.Error())
		return errors.HandleError(c, err)
	}

	// Authenticate user
	authenticatedUser, token, err := h.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		h.logger.Error(ctx, "user login failed", "email", req.Email, "error", err.Error())
		return errors.HandleError(c, err)
	}

	response := AuthResponse{
		User: UserResponse{
			ID:    authenticatedUser.ID,
			Name:  authenticatedUser.Name,
			Email: authenticatedUser.Email,
		},
		Token: token,
	}

	h.logger.Info(ctx, "user logged in successfully", "userID", authenticatedUser.ID, "email", authenticatedUser.Email)
	return c.JSON(http.StatusOK, response)
}
