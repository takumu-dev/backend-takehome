package auth

import "errors"

var (
	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound      = errors.New("user not found")
	ErrUserExists        = errors.New("user already exists")
	ErrInvalidEmail      = errors.New("invalid email format")
	ErrWeakPassword      = errors.New("password is too weak")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

// Token errors
var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrTokenGeneration  = errors.New("failed to generate token")
	ErrTokenValidation  = errors.New("failed to validate token")
	ErrInvalidSecretKey = errors.New("invalid secret key")
	ErrMissingToken     = errors.New("missing token")
	ErrInvalidUserID    = errors.New("invalid user ID")
	ErrInvalidDuration  = errors.New("invalid duration")
	ErrEmptyToken       = errors.New("empty token")
	ErrTokenExpired     = errors.New("token expired")
)

// Password validation errors
var (
	ErrPasswordTooShort          = errors.New("password must be at least 8 characters long")
	ErrPasswordTooLong           = errors.New("password must be no more than 72 characters long")
	ErrPasswordMissingUppercase  = errors.New("password must contain at least one uppercase letter")
	ErrPasswordMissingLowercase  = errors.New("password must contain at least one lowercase letter")
	ErrPasswordMissingNumber     = errors.New("password must contain at least one number")
	ErrPasswordMissingSpecial    = errors.New("password must contain at least one special character")
	ErrPasswordTooCommon         = errors.New("password is too common")
	ErrPasswordTooWeak           = errors.New("password is too weak")
)
