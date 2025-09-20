package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"blog-platform/internal/domain/auth"
	infraAuth "blog-platform/internal/infrastructure/auth"
)

func TestPasswordValidator_ValidatePassword(t *testing.T) {
	validator := infraAuth.NewPasswordValidator()

	tests := []struct {
		name        string
		password    string
		expectError bool
		expectedErr error
	}{
		{
			name:        "valid strong password",
			password:    "StrongPass123!",
			expectError: false,
		},
		{
			name:        "valid password with special chars",
			password:    "MySecure@Pass2024",
			expectError: false,
		},
		{
			name:        "empty password should fail",
			password:    "",
			expectError: true,
			expectedErr: auth.ErrPasswordTooShort,
		},
		{
			name:        "password too short should fail",
			password:    "Short1!",
			expectError: true,
			expectedErr: auth.ErrPasswordTooShort,
		},
		{
			name:        "password too long should fail",
			password:    "ThisPasswordIsWayTooLongAndExceedsTheMaximumAllowedLengthForSecurityReasons123!",
			expectError: true,
			expectedErr: auth.ErrPasswordTooLong,
		},
		{
			name:        "password without uppercase should fail",
			password:    "lowercase123!",
			expectError: true,
			expectedErr: auth.ErrPasswordMissingUppercase,
		},
		{
			name:        "password without lowercase should fail",
			password:    "UPPERCASE123!",
			expectError: true,
			expectedErr: auth.ErrPasswordMissingLowercase,
		},
		{
			name:        "password without numbers should fail",
			password:    "NoNumbers!",
			expectError: true,
			expectedErr: auth.ErrPasswordMissingNumber,
		},
		{
			name:        "password without special chars should fail",
			password:    "NoSpecialChars123",
			expectError: true,
			expectedErr: auth.ErrPasswordMissingSpecial,
		},
		{
			name:        "password with only spaces should fail",
			password:    "        ",
			expectError: true,
			expectedErr: auth.ErrPasswordTooShort,
		},
		{
			name:        "sequential characters should fail",
			password:    "Abcde12345!",
			expectError: true,
			expectedErr: auth.ErrPasswordTooWeak,
		},
		{
			name:        "repeated characters should fail",
			password:    "Aaaa1111!",
			expectError: true,
			expectedErr: auth.ErrPasswordTooWeak,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePassword(tt.password)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}



func TestPasswordValidator_HasSequentialChars(t *testing.T) {
	validator := infraAuth.NewPasswordValidator()

	tests := []struct {
		name       string
		password   string
		hasSequence bool
	}{
		{
			name:       "ascending sequence abcd",
			password:   "Myabcdpass123!",
			hasSequence: true,
		},
		{
			name:       "ascending sequence 1234",
			password:   "Pass1234word!",
			hasSequence: true,
		},
		{
			name:       "descending sequence dcba",
			password:   "Mydcbapass456!",
			hasSequence: true,
		},
		{
			name:       "descending sequence 4321",
			password:   "Pass4321word!",
			hasSequence: true,
		},
		{
			name:       "no sequence",
			password:   "RandomPass2024!",
			hasSequence: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasSequence := validator.HasSequentialChars(tt.password)
			assert.Equal(t, tt.hasSequence, hasSequence)
		})
	}
}

func TestPasswordValidator_HasRepeatedChars(t *testing.T) {
	validator := infraAuth.NewPasswordValidator()

	tests := []struct {
		name        string
		password    string
		hasRepeated bool
	}{
		{
			name:        "repeated letters aaa",
			password:    "Myaaapass123!",
			hasRepeated: true,
		},
		{
			name:        "repeated numbers 111",
			password:    "Pass111word!",
			hasRepeated: true,
		},
		{
			name:        "repeated special chars !!!",
			password:    "Password123!!!",
			hasRepeated: true,
		},
		{
			name:        "no repeated chars",
			password:    "RandomPass2024!",
			hasRepeated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasRepeated := validator.HasRepeatedChars(tt.password)
			assert.Equal(t, tt.hasRepeated, hasRepeated)
		})
	}
}

func TestPasswordValidator_Interface(t *testing.T) {
	validator := infraAuth.NewPasswordValidator()
	
	// Verify that PasswordValidator implements the expected interface
	assert.NotNil(t, validator)
	assert.NotNil(t, validator.ValidatePassword)
	assert.NotNil(t, validator.HasSequentialChars)
	assert.NotNil(t, validator.HasRepeatedChars)
}
