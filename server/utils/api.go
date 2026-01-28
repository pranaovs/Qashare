package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pranaovs/qashare/routes/apierrors"
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
		return nil, errors.New("authorization header missing or malformed")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
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

// SendJSON is a helper function that sends a JSON response with the specified
// HTTP status code and data.
func SendJSON(c *gin.Context, statusCode int, data any) {
	c.JSON(statusCode, data)
}

// SendOK sends a standard OK response with a message.
func SendOK(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{"message": message})
}

// SendData sends a standard OK response with arbitrary data.
func SendData(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}

// SendError inspects the provided error and sends an appropriate JSON response.
// This function differentiates between known application errors and unexpected errors.
// Application errors are sent with their specific HTTP status codes and messages,
// Generic errors result in a 500 Internal Server Error response.
func SendError(c *gin.Context, err error) {
	// Check if the error is our custom AppError
	if appErr, ok := err.(*apierrors.AppError); ok {

		LogDebug(c, fmt.Sprintf("Error: %s | Code: %s | Internal: %v",
			appErr.Message, appErr.MachineCode, appErr.Err))

		// Send the encapsulated response and return
		c.JSON(appErr.HTTPCode, gin.H{
			"code":    appErr.MachineCode,
			"message": appErr.Message,
		})
		return
	}

	// Handle unexpected/unknown errors (Panic recovery or generic errors)
	LogError(c, "[ERROR] Internal Server Error: %v", err)

	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    "INTERNAL_ERROR",
		"message": "Something went wrong on our end. Please report this.",
	})
}
