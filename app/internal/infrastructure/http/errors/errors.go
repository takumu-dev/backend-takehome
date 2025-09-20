package errors

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	// Client errors (4xx)
	ErrCodeValidation     ErrorCode = "validation_error"
	ErrCodeInvalidRequest ErrorCode = "invalid_request"
	ErrCodeUnauthorized   ErrorCode = "unauthorized"
	ErrCodeForbidden      ErrorCode = "forbidden"
	ErrCodeNotFound       ErrorCode = "not_found"
	ErrCodeConflict       ErrorCode = "conflict"
	ErrCodeUserExists     ErrorCode = "user_exists"
	ErrCodeInvalidCredentials ErrorCode = "invalid_credentials"
	ErrCodeRateLimitExceeded ErrorCode = "rate_limit_exceeded"
	
	// Server errors (5xx)
	ErrCodeInternal       ErrorCode = "internal_error"
	ErrCodeDatabase       ErrorCode = "database_error"
	ErrCodeService        ErrorCode = "service_error"
)

// APIError represents a standardized API error
type APIError struct {
	Code       ErrorCode `json:"error"`
	Message    string    `json:"message"`
	Details    []string  `json:"details,omitempty"`
	StatusCode int       `json:"-"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ErrorResponse represents the standard error response format
type ErrorResponse struct {
	Error   string   `json:"error"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

// NewAPIError creates a new API error
func NewAPIError(code ErrorCode, message string, statusCode int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// NewValidationError creates a validation error from validator errors
func NewValidationError(err error) *APIError {
	var details []string
	
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			detail := formatValidationError(fieldError)
			details = append(details, detail)
		}
	} else {
		details = append(details, err.Error())
	}
	
	return &APIError{
		Code:       ErrCodeValidation,
		Message:    "Request validation failed",
		Details:    details,
		StatusCode: http.StatusBadRequest,
	}
}

// NewDomainError creates an error from domain validation
func NewDomainError(err error) *APIError {
	message := err.Error()
	
	// Map common domain errors to appropriate HTTP status codes
	switch {
	case strings.Contains(message, "not found"):
		return NewAPIError(ErrCodeNotFound, message, http.StatusNotFound)
	case strings.Contains(message, "unauthorized") || strings.Contains(message, "forbidden"):
		return NewAPIError(ErrCodeForbidden, message, http.StatusForbidden)
	case strings.Contains(message, "already exists"):
		return NewAPIError(ErrCodeConflict, message, http.StatusConflict)
	case strings.Contains(message, "invalid credentials"):
		return NewAPIError(ErrCodeInvalidCredentials, message, http.StatusUnauthorized)
	default:
		return NewAPIError(ErrCodeValidation, message, http.StatusBadRequest)
	}
}

// Common pre-defined errors
var (
	ErrUnauthorized = NewAPIError(
		ErrCodeUnauthorized,
		"Missing or invalid authorization token",
		http.StatusUnauthorized,
	)
	
	ErrForbidden = NewAPIError(
		ErrCodeForbidden,
		"You don't have permission to access this resource",
		http.StatusForbidden,
	)
	
	ErrNotFound = NewAPIError(
		ErrCodeNotFound,
		"The requested resource was not found",
		http.StatusNotFound,
	)
	
	ErrInternal = NewAPIError(
		ErrCodeInternal,
		"An internal server error occurred",
		http.StatusInternalServerError,
	)
	
	ErrInvalidRequest = NewAPIError(
		ErrCodeInvalidRequest,
		"Invalid request format",
		http.StatusBadRequest,
	)
	
	ErrTooManyRequests = NewAPIError(
		ErrCodeRateLimitExceeded,
		"Rate limit exceeded. Please try again later",
		http.StatusTooManyRequests,
	)
)

// HandleError handles different types of errors and returns appropriate HTTP responses
func HandleError(c echo.Context, err error) error {
	var apiErr *APIError
	
	switch e := err.(type) {
	case *APIError:
		apiErr = e
	case validator.ValidationErrors:
		apiErr = NewValidationError(e)
	default:
		// Check if it's a domain error
		if isDomainError(err) {
			apiErr = NewDomainError(err)
		} else {
			apiErr = ErrInternal
		}
	}
	
	response := ErrorResponse{
		Error:   string(apiErr.Code),
		Message: apiErr.Message,
		Details: apiErr.Details,
	}
	
	return c.JSON(apiErr.StatusCode, response)
}

// formatValidationError formats a single validation error into a human-readable message
func formatValidationError(fieldError validator.FieldError) string {
	field := fieldError.Field()
	tag := fieldError.Tag()
	param := fieldError.Param()
	
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, param)
	case "max":
		return fmt.Sprintf("%s cannot exceed %s characters", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", field, param)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	default:
		return fmt.Sprintf("%s validation failed for tag '%s'", field, tag)
	}
}

// isDomainError checks if an error is from the domain layer
func isDomainError(err error) bool {
	message := err.Error()
	
	// Common domain error patterns
	domainPatterns := []string{
		"cannot be empty",
		"must be at least",
		"cannot exceed",
		"must be positive",
		"already exists",
		"not found",
		"unauthorized",
		"forbidden",
		"invalid",
	}
	
	for _, pattern := range domainPatterns {
		if strings.Contains(strings.ToLower(message), pattern) {
			return true
		}
	}
	
	return false
}
