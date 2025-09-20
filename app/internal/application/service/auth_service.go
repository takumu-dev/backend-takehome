package service

import (
	"context"
	"fmt"
	"time"

	"blog-platform/internal/domain/auth"
	"blog-platform/internal/domain/user"
)

// AuthService implements the auth.AuthService interface
type AuthService struct {
	userService  user.Service
	tokenService auth.TokenService
	logger       Logger
}

// NewAuthService creates a new authentication service
func NewAuthService(userService user.Service, tokenService auth.TokenService, logger Logger) auth.AuthService {
	return &AuthService{
		userService:  userService,
		tokenService: tokenService,
		logger:       logger,
	}
}

// GenerateToken generates a JWT token for the given user
func (a *AuthService) GenerateToken(ctx context.Context, user *user.User) (string, error) {
	a.logger.Debug(ctx, "Generating token for user", "user_id", user.ID, "email", user.Email)
	
	if user == nil {
		a.logger.Error(ctx, "Cannot generate token for nil user")
		return "", fmt.Errorf("user cannot be nil")
	}
	
	// Generate token with 24 hour duration
	token, err := a.tokenService.GenerateToken(user.ID, user.Email, 24*time.Hour)
	if err != nil {
		a.logger.Error(ctx, "Failed to generate token", "user_id", user.ID, "error", err)
		return "", err
	}
	
	a.logger.Info(ctx, "Token generated successfully", "user_id", user.ID)
	return token, nil
}

// ValidateToken validates a JWT token and returns the claims
func (a *AuthService) ValidateToken(ctx context.Context, token string) (*auth.TokenClaims, error) {
	a.logger.Debug(ctx, "Validating token")
	
	claims, err := a.tokenService.ValidateToken(token)
	if err != nil {
		a.logger.Warn(ctx, "Token validation failed", "error", err)
		return nil, err
	}
	
	a.logger.Debug(ctx, "Token validated successfully", "user_id", claims.UserID)
	return claims, nil
}

// Login authenticates a user and returns user data with token
func (a *AuthService) Login(ctx context.Context, email, password string) (*user.User, string, error) {
	a.logger.Info(ctx, "User login attempt", "email", email)
	
	// Use the user service to authenticate
	u, err := a.userService.Login(ctx, email, password)
	if err != nil {
		a.logger.Warn(ctx, "Login failed", "email", email, "error", err)
		return nil, "", err
	}
	
	// Generate token for the authenticated user
	token, err := a.GenerateToken(ctx, u)
	if err != nil {
		a.logger.Error(ctx, "Failed to generate token after login", "error", err)
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}
	
	a.logger.Info(ctx, "User logged in successfully", "user_id", u.ID, "email", email)
	return u, token, nil
}

// Register creates a new user and returns user data with token
func (a *AuthService) Register(ctx context.Context, name, email, password string) (*user.User, string, error) {
	a.logger.Info(ctx, "User registration attempt", "name", name, "email", email)
	
	// Use the user service to register
	u, err := a.userService.Register(ctx, name, email, password)
	if err != nil {
		a.logger.Warn(ctx, "Registration failed", "name", name, "email", email, "error", err)
		return nil, "", err
	}
	
	// Generate token for the new user
	token, err := a.GenerateToken(ctx, u)
	if err != nil {
		a.logger.Error(ctx, "Failed to generate token after registration", "error", err)
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}
	
	a.logger.Info(ctx, "User registered successfully", "user_id", u.ID, "name", name, "email", email)
	return u, token, nil
}

// RefreshToken refreshes an existing token
func (a *AuthService) RefreshToken(ctx context.Context, token string) (string, error) {
	a.logger.Debug(ctx, "Refreshing token")
	
	newToken, err := a.tokenService.RefreshToken(token)
	if err != nil {
		a.logger.Warn(ctx, "Token refresh failed", "error", err)
		return "", err
	}
	
	a.logger.Info(ctx, "Token refreshed successfully")
	return newToken, nil
}
