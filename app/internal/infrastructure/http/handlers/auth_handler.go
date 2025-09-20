package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"blog-platform/internal/application/service"
	"blog-platform/internal/domain/auth"
	"blog-platform/internal/domain/user"
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
	Name     string `json:"name" validate:"required,min=2,max=255"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
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
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration request"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c echo.Context) error {
	ctx := c.Request().Context()
	h.logger.Info(ctx, "registration request received")

	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error(ctx, "failed to bind registration request", "error", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request format",
		})
	}

	if err := c.Validate(&req); err != nil {
		h.logger.Error(ctx, "registration request validation failed", "error", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_failed",
			Message: err.Error(),
		})
	}

	// Register user
	registeredUser, token, err := h.authService.Register(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		if err == user.ErrUserExists {
			h.logger.Warn(ctx, "registration attempt with existing email", "email", req.Email)
			return c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "user_exists",
				Message: "User with this email already exists",
			})
		}
		h.logger.Error(ctx, "user registration failed", "email", req.Email, "error", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "registration_failed",
			Message: "Failed to register user",
		})
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
// @Description Authenticate user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()
	h.logger.Info(ctx, "login request received")

	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error(ctx, "failed to bind login request", "error", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request format",
		})
	}

	if err := c.Validate(&req); err != nil {
		h.logger.Error(ctx, "login request validation failed", "error", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_failed",
			Message: err.Error(),
		})
	}

	// Authenticate user
	authenticatedUser, token, err := h.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		if err == user.ErrInvalidCredentials {
			h.logger.Warn(ctx, "login attempt with invalid credentials", "email", req.Email)
			return c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "invalid_credentials",
				Message: "Invalid email or password",
			})
		}
		h.logger.Error(ctx, "user login failed", "email", req.Email, "error", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "login_failed",
			Message: "Failed to authenticate user",
		})
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
