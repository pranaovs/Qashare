package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	Loadenv()
}

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
		log.Fatal("failed to generate random bytes for JWT secret:", err)
	}

	return base64.StdEncoding.EncodeToString(b)
}

var jwtSecret = []byte(GetEnv("JWT_SECRET", randB64()))

func GenerateJWT(userID string) (string, error) {
	expiryHours := GetEnvDuration("JWT_EXPIRY", 60*60*24) // Default to 24 hours
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(expiryHours).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ExtractClaims(authHeader string) (jwt.MapClaims, error) {
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, ErrInvalidToken.Msg("authorization header missing or malformed")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(jwtSecret), nil
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

func ExtractUserID(authHeader string) (string, error) {
	claims, err := ExtractClaims(authHeader)
	if err != nil {
		return "", err
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", ErrInvalidToken.Msg("invalid token claims")
	}

	return userID, nil
}
