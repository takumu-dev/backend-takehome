package auth

// TokenClaims represents the JWT token claims
type TokenClaims struct {
	UserID    int    `json:"user_id"`
	Email     string `json:"email"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}
