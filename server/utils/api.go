package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/pranaovs/qashare/models"

	"github.com/gin-gonic/gin"
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
		return "", errors.New("empty password provided")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
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
	expiryStr := GetEnv("JWT_EXPIRY", "24")
	expiryHours, err := strconv.Atoi(expiryStr)
	if err != nil || expiryHours <= 0 {
		return "", fmt.Errorf("invalid JWT_EXPIRY value: %q, must be a positive integer", expiryStr)
	}
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Duration(expiryHours) * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ExtractClaims(authHeader string) (jwt.MapClaims, error) {
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, errors.New("authorization header missing or malformed")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, errors.New("invalid token")
	}
	if !token.Valid {
		return nil, errors.New("expired token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
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
		return "", errors.New("invalid token claims")
	}

	return userID, nil
}

// AbortWithStatusJSON is a unified helper function that aborts the request
// and sends a JSON response with the specified HTTP status code and error message.
// This replaces the pattern of calling c.JSON() followed by c.Abort() separately.
func AbortWithStatusJSON(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, gin.H{"error": message})
}

// AbortWithError aborts the request and sends a structured error response
func AbortWithError(c *gin.Context, statusCode int, errResp models.ErrorResponse) {
	LogError(c.Request.Context(), "Request aborted with error",
		fmt.Errorf("%s", errResp.Error),
		"code", errResp.Code,
		"message", errResp.Message,
		"status", statusCode,
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
	)
	c.AbortWithStatusJSON(statusCode, errResp)
}

// SendJSON is a helper function that sends a JSON response with the specified
// HTTP status code and data.
func SendJSON(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

// SendError is a helper function that sends a JSON error response without aborting.
// Use AbortWithStatusJSON when you need to abort the request chain.
func SendError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{"error": message})
}

// SendErrorWithCode sends a structured error response with error code
func SendErrorWithCode(c *gin.Context, statusCode int, errResp models.ErrorResponse) {
	LogError(c.Request.Context(), "Error response sent",
		fmt.Errorf("%s", errResp.Error),
		"code", errResp.Code,
		"message", errResp.Message,
		"status", statusCode,
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
	)
	c.JSON(statusCode, errResp)
}
