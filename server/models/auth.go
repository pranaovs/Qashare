package models

import "github.com/golang-jwt/jwt/v5"

// TokenType represents the type of JWT token (access or refresh).
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// TokenClaims are the JWT claims used for both access and refresh tokens.
type TokenClaims struct {
	jwt.RegisteredClaims
	TokenType TokenType `json:"typ" example:"access"`
	SessionID string    `json:"sid,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// TokenResponse is the JSON body returned on login and token refresh.
type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken string `json:"refresh_token,omitempty" example:"eyJhbGciOiJIUzI1NiIs..."`
	TokenType    string `json:"token_type" example:"Bearer"`
}
