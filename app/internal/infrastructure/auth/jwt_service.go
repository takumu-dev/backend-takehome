package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"blog-platform/internal/domain/auth"
)

// JWTService implements the auth.TokenService interface using JWT
type JWTService struct {
	secretKey []byte
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string) *JWTService {
	return &JWTService{
		secretKey: []byte(secretKey),
	}
}

// Claims represents the JWT claims structure
type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken creates a JWT token with the given claims
func (j *JWTService) GenerateToken(userID int, email string, duration time.Duration) (string, error) {
	// Validate inputs
	if userID <= 0 {
		return "", auth.ErrInvalidUserID
	}
	if email == "" {
		return "", auth.ErrInvalidEmail
	}
	if duration <= 0 {
		return "", auth.ErrInvalidDuration
	}
	if len(j.secretKey) == 0 {
		return "", auth.ErrInvalidSecretKey
	}

	now := time.Now()
	expirationTime := now.Add(duration)

	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "blog-platform",
			Subject:   email,
			Audience:  []string{"blog-platform-api"},
		},
	}

	// Use HS256 for security (HMAC with SHA-256)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns claims
func (j *JWTService) ValidateToken(tokenString string) (*auth.TokenClaims, error) {
	if tokenString == "" {
		return nil, auth.ErrEmptyToken
	}
	if len(j.secretKey) == 0 {
		return nil, auth.ErrInvalidSecretKey
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, auth.ErrInvalidToken
		}
		return j.secretKey, nil
	})

	if err != nil {
		// Check if token is expired by examining the error message
		errMsg := err.Error()
		if errMsg == "token is expired" || errMsg == "Token is expired" {
			return nil, auth.ErrTokenExpired
		}
		return nil, auth.ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, auth.ErrInvalidToken
	}

	return &auth.TokenClaims{
		UserID:    claims.UserID,
		Email:     claims.Email,
		IssuedAt:  claims.IssuedAt.Unix(),
		ExpiresAt: claims.ExpiresAt.Unix(),
	}, nil
}

// RefreshToken generates a new token from an existing valid token
func (j *JWTService) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Generate new token with same user info but extended expiry
	// Add a small delay to ensure different issued at time
	time.Sleep(1 * time.Millisecond)
	return j.GenerateToken(claims.UserID, claims.Email, 24*time.Hour)
}
