package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pranaovs/qashare/config"
	"github.com/pranaovs/qashare/models"
	"golang.org/x/crypto/bcrypt"
)

// Passwords

// HashPassword hashes a plaintext password using bcrypt.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", ErrInvalidPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", ErrHashingFailed.WithError(err)
	}
	return string(hash), nil
}

// CheckPassword compares a plaintext password with its hashed version.
func CheckPassword(password, hashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	return err == nil
}

func randB64() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		slog.Error("Failed to generate random bytes for JWT secret", "error", err)
		os.Exit(1)
	}

	return base64.StdEncoding.EncodeToString(b)
}

func generateToken(userID uuid.UUID, tokenType models.TokenType, expiry time.Duration, jwtConfig config.JWTConfig) (string, uuid.UUID, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(expiry)
	tokenID := uuid.New()
	claims := models.TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jwtConfig.Issuer,
			Subject:   userID.String(),
			Audience:  jwt.ClaimStrings{jwtConfig.Audience},
			ID:        tokenID.String(),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
		TokenType: tokenType,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(jwtConfig.Secret))
	if err != nil {
		return "", uuid.UUID{}, time.Time{}, err
	}
	return signed, tokenID, expiresAt, nil
}

func GenerateRefreshToken(userID uuid.UUID, jwtConfig config.JWTConfig) (string, uuid.UUID, time.Time, error) {
	return generateToken(userID, models.TokenTypeRefresh, jwtConfig.RefreshExpiry, jwtConfig)
}

func GenerateAccessToken(userID uuid.UUID, jwtConfig config.JWTConfig) (string, error) {
	signed, _, _, err := generateToken(userID, models.TokenTypeAccess, jwtConfig.AccessExpiry, jwtConfig)
	return signed, err
}

func extractClaims(tokenString string, jwtConfig config.JWTConfig) (*models.TokenClaims, error) {
	claims := &models.TokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(jwtConfig.Secret), nil
	},
		jwt.WithIssuer(jwtConfig.Issuer),
		jwt.WithAudience(jwtConfig.Audience),
	)
	if err != nil {
		return nil, ErrInvalidToken.Msg("failed to parse token")
	}
	if !token.Valid {
		return nil, ErrInvalidToken.Msg("expired token")
	}

	return claims, nil
}

func extractBearerToken(authHeader string) (string, error) {
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return "", ErrInvalidToken.Msg("authorization header missing or malformed")
	}
	return strings.TrimPrefix(authHeader, "Bearer "), nil
}

func ExtractAccessClaims(authHeader string, jwtConfig config.JWTConfig) (*models.TokenClaims, error) {
	tokenString, err := extractBearerToken(authHeader)
	if err != nil {
		return nil, err
	}

	claims, err := extractClaims(tokenString, jwtConfig)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != models.TokenTypeAccess {
		return nil, ErrInvalidToken.Msg("expected access token")
	}

	return claims, nil
}

func ExtractRefreshClaims(refreshToken string, jwtConfig config.JWTConfig) (*models.TokenClaims, error) {
	claims, err := extractClaims(refreshToken, jwtConfig)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != models.TokenTypeRefresh {
		return nil, ErrInvalidToken.Msg("expected refresh token")
	}

	return claims, nil
}

func ExtractUserID(authHeader string, jwtConfig config.JWTConfig) (uuid.UUID, error) {
	claims, err := ExtractAccessClaims(authHeader, jwtConfig)
	if err != nil {
		return uuid.UUID{}, err
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.UUID{}, ErrInvalidToken.Msg("invalid subject in token")
	}

	return userID, nil
}
