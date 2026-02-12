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
	"github.com/pranaovs/qashare/config"
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

func GenerateJWT(userID string, jwtConfig config.JWTConfig) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(jwtConfig.Expiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtConfig.Secret))
}

func ExtractClaims(authHeader string, jwtConfig config.JWTConfig) (jwt.MapClaims, error) {
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, ErrInvalidToken.Msg("authorization header missing or malformed")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(jwtConfig.Secret), nil
	})
	if err != nil {
		return nil, ErrInvalidToken.Msg("failed to parse token")
	}
	if !token.Valid {
		return nil, ErrInvalidToken.Msg("expired token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken.Msg("invalid token claims")
	}

	return claims, nil
}

func ExtractUserID(authHeader string, jwtConfig config.JWTConfig) (string, error) {
	claims, err := ExtractClaims(authHeader, jwtConfig)
	if err != nil {
		return "", err
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", ErrInvalidToken.Msg("invalid token claims")
	}

	return userID, nil
}
