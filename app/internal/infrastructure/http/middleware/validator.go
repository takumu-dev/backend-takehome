package middleware

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// CustomValidator wraps the validator instance with custom validation rules
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator creates a new validator instance with custom validation rules
func NewValidator() *CustomValidator {
	v := validator.New()
	
	// Register custom validation functions
	v.RegisterValidation("strong_password", validateStrongPassword)
	v.RegisterValidation("no_html", validateNoHTML)
	v.RegisterValidation("safe_string", validateSafeString)
	
	return &CustomValidator{validator: v}
}

// Validate validates the struct
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// validateStrongPassword validates password strength
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	
	// At least 8 characters
	if len(password) < 8 {
		return false
	}
	
	// Must contain at least one letter and one number
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	
	return hasLetter && hasNumber
}

// validateNoHTML validates that the field contains no HTML tags
func validateNoHTML(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	
	// Simple HTML tag detection
	htmlPattern := regexp.MustCompile(`<[^>]*>`)
	return !htmlPattern.MatchString(value)
}

// validateSafeString validates that the field contains only safe characters
func validateSafeString(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	
	// Allow alphanumeric, spaces, and common punctuation
	safePattern := regexp.MustCompile(`^[a-zA-Z0-9\s\.,!?\-_@#$%&*()+=\[\]{}|\\:";'<>?/~` + "`" + `]*$`)
	return safePattern.MatchString(value)
}

// SanitizeInput removes potentially dangerous characters from input
func SanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")
	
	// Trim whitespace
	input = strings.TrimSpace(input)
	
	return input
}
