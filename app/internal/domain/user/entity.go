package user

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user entity in the domain
type User struct {
	ID           int       `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// NewUser creates a new user instance with validation and password hashing
func NewUser(name, email, password string) (*User, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}

	if err := validateEmail(email); err != nil {
		return nil, err
	}

	if err := validatePassword(password); err != nil {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	now := time.Now()
	return &User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// UpdateProfile updates the user's profile information with validation
func (u *User) UpdateProfile(name, email string) error {
	if err := validateName(name); err != nil {
		return err
	}

	if err := validateEmail(email); err != nil {
		return err
	}

	u.Name = name
	u.Email = email
	u.UpdatedAt = time.Now()
	return nil
}

// ValidatePassword checks if the provided password matches the user's hashed password
func (u *User) ValidatePassword(password string) bool {
	if password == "" {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// UpdatePassword updates the user's password with validation and hashing
func (u *User) UpdatePassword(password string) error {
	if err := validatePassword(password); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	u.PasswordHash = string(hashedPassword)
	u.UpdatedAt = time.Now()
	return nil
}

// validateName validates the user's name
func validateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("name cannot be empty")
	}
	if len(name) > 255 {
		return errors.New("name must be less than 255 characters")
	}
	return nil
}

// validateEmail validates the user's email format
func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("email cannot be empty")
	}
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	return nil
}

// validatePassword validates the password requirements
func validatePassword(password string) error {
	if password == "" {
		return errors.New("password cannot be empty")
	}
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}
	return nil
}
