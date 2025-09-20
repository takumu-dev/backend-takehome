package auth

import (
	"strings"
	"unicode"

	"blog-platform/internal/domain/auth"
)

// passwordValidator provides password validation and strength checking
type passwordValidator struct {}

// NewPasswordValidator creates a new password validator
func NewPasswordValidator() auth.PasswordValidator {
	return &passwordValidator{}
}

// ValidatePassword validates a password against security requirements
func (pv *passwordValidator) ValidatePassword(password string) error {
	// Trim spaces for length check
	trimmed := strings.TrimSpace(password)
	
	// Check length requirements
	if len(trimmed) < 8 {
		return auth.ErrPasswordTooShort
	}
	if len(password) > 72 {
		return auth.ErrPasswordTooLong
	}

	// Check for required character types
	var hasUpper, hasLower, hasNumber, hasSpecial bool
	
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Check for missing character types
	if !hasUpper {
		return auth.ErrPasswordMissingUppercase
	}
	if !hasLower {
		return auth.ErrPasswordMissingLowercase
	}
	if !hasNumber {
		return auth.ErrPasswordMissingNumber
	}
	if !hasSpecial {
		return auth.ErrPasswordMissingSpecial
	}


	// Check for weak patterns
	if pv.HasSequentialChars(password) || pv.HasRepeatedChars(password) {
		return auth.ErrPasswordTooWeak
	}

	return nil
}



// HasSequentialChars checks for sequential characters in the password
func (pv *passwordValidator) HasSequentialChars(password string) bool {
	password = strings.ToLower(password)
	
	// Check for sequences of 4 or more characters (like "abcd" or "1234")
	for i := 0; i < len(password)-3; i++ {
		// Check ascending sequence
		if password[i]+1 == password[i+1] && password[i+1]+1 == password[i+2] && password[i+2]+1 == password[i+3] {
			return true
		}
		// Check descending sequence
		if password[i]-1 == password[i+1] && password[i+1]-1 == password[i+2] && password[i+2]-1 == password[i+3] {
			return true
		}
	}

	// Check for common keyboard sequences (longer sequences only)
	keyboardSequences := []string{
		"qwerty", "asdf", "zxcv", "123456", "abcdef",
		"qwertyuiop", "asdfghjkl", "zxcvbnm",
	}

	for _, seq := range keyboardSequences {
		if strings.Contains(password, seq) || strings.Contains(password, reverse(seq)) {
			return true
		}
	}

	return false
}

// HasRepeatedChars checks for repeated characters in the password
func (pv *passwordValidator) HasRepeatedChars(password string) bool {
	// Check for 3 or more consecutive identical characters
	for i := 0; i < len(password)-2; i++ {
		if password[i] == password[i+1] && password[i+1] == password[i+2] {
			return true
		}
	}

	// Check for patterns like "aaa", "111", "!!!"
	charCount := make(map[rune]int)
	for _, char := range password {
		charCount[char]++
		if charCount[char] >= 4 {
			return true
		}
	}

	return false
}

// reverse reverses a string
func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
