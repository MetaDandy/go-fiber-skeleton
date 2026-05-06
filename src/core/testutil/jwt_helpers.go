package testutil

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateTestToken creates a valid JWT token for testing purposes.
// It uses the JWT_SECRET environment variable for signing.
func GenerateTestToken(userID, email, role string, permissions []string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "test-secret-key-for-testing-only"
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = userID
	claims["email"] = email
	claims["role"] = role
	claims["permissions"] = permissions
	claims["exp"] = time.Now().Add(time.Hour).Unix()

	return token.SignedString([]byte(secret))
}

// GenerateExpiredToken creates an expired JWT token for testing error handling.
func GenerateExpiredToken(userID, email string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "test-secret-key-for-testing-only"
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = userID
	claims["email"] = email
	claims["role"] = "user"
	claims["permissions"] = []string{}
	claims["exp"] = time.Now().Add(-1 * time.Hour).Unix() // Expired 1 hour ago

	return token.SignedString([]byte(secret))
}

// GenerateInvalidToken creates a JWT token signed with a different secret
// to test invalid signature handling.
func GenerateInvalidToken(userID, email string) (string, error) {
	// Sign with a different secret to create an invalid token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = userID
	claims["email"] = email
	claims["role"] = "user"
	claims["permissions"] = []string{}
	claims["exp"] = time.Now().Add(time.Hour).Unix()

	return token.SignedString([]byte("invalid-signing-secret"))
}

// GenerateTestTokenWithExpiration creates a JWT token with a custom expiration time.
func GenerateTestTokenWithExpiration(userID, email, role string, permissions []string, expiresAt time.Time) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "test-secret-key-for-testing-only"
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = userID
	claims["email"] = email
	claims["role"] = role
	if permissions != nil {
		claims["permissions"] = permissions
	} else {
		claims["permissions"] = []string{}
	}
	claims["exp"] = expiresAt.Unix()

	return token.SignedString([]byte(secret))
}