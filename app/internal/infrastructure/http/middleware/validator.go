package middleware

import (
	"github.com/go-playground/validator/v10"
)

// CustomValidator wraps the validator instance
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator creates a new validator instance
func NewValidator() *CustomValidator {
	return &CustomValidator{validator: validator.New()}
}

// Validate validates the struct
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}
