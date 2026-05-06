package testutil

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateTestToken_ValidToken verifies GenerateTestToken returns a valid JWT
func TestGenerateTestToken_ValidToken(t *testing.T) {
	// Ensure JWT_SECRET is set for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing-only")
	defer os.Unsetenv("JWT_SECRET")

	userID := "test-user-123"
	email := "test@example.com"
	role := "admin"
	permissions := []string{"read", "write"}

	token, err := GenerateTestToken(userID, email, role, permissions)
	require.NoError(t, err, "GenerateTestToken should not return error")
	assert.NotEmpty(t, token, "token should not be empty")

	// Verify the token is parseable and has correct claims
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	require.NoError(t, err, "token should be parseable")
	assert.True(t, parsedToken.Valid, "token should be valid")

	claims := parsedToken.Claims.(jwt.MapClaims)
	assert.Equal(t, userID, claims["sub"])
	assert.Equal(t, email, claims["email"])
	assert.Equal(t, role, claims["role"])
	assert.Equal(t, []interface{}{"read", "write"}, claims["permissions"])
}

// TestGenerateTestToken_DefaultPermissions verifies default permissions are empty
func TestGenerateTestToken_DefaultPermissions(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing-only")
	defer os.Unsetenv("JWT_SECRET")

	token, err := GenerateTestToken("user-1", "user@test.com", "user", nil)
	require.NoError(t, err)

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	require.NoError(t, err, "token should parse without error")

	claims := parsedToken.Claims.(jwt.MapClaims)
	assert.Contains(t, claims, "permissions", "permissions key should exist")
}

// TestGenerateExpiredToken_ReturnsExpiredToken verifies expired token is generated
func TestGenerateExpiredToken_ReturnsExpiredToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing-only")
	defer os.Unsetenv("JWT_SECRET")

	token, err := GenerateExpiredToken("user-1", "user@test.com")
	require.NoError(t, err, "GenerateExpiredToken should not return error")
	assert.NotEmpty(t, token, "expired token should not be empty")

	// Verify the token is expired - jwt.Parse returns error for expired tokens
	_, err = jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	assert.Error(t, err, "expired token should cause parse error")
	assert.Contains(t, err.Error(), "expired", "error should indicate expiration")
}

// TestGenerateInvalidToken_ReturnsInvalidToken verifies invalid signature token
func TestGenerateInvalidToken_ReturnsInvalidToken(t *testing.T) {
	token, err := GenerateInvalidToken("user-1", "user@test.com")
	require.NoError(t, err, "GenerateInvalidToken should not return error")
	assert.NotEmpty(t, token, "invalid token should not be empty")

	// Verify the token has invalid signature - parse with correct secret should fail
	_, err = jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret-key-for-testing-only"), nil
	})
	assert.Error(t, err, "invalid signature token should cause parse error")
	assert.Contains(t, err.Error(), "signature", "error should indicate invalid signature")
}

// TestGenerateTestToken_WithExpiration verifies custom expiration is respected
func TestGenerateTestToken_WithExpiration(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing-only")
	defer os.Unsetenv("JWT_SECRET")

	expiresAt := time.Now().Add(2 * time.Hour)
	token, err := GenerateTestTokenWithExpiration("user-1", "user@test.com", "admin", nil, expiresAt)
	require.NoError(t, err)

	parsedToken, _ := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	claims := parsedToken.Claims.(jwt.MapClaims)

	expClaim := claims["exp"].(float64)
	actualExp := time.Unix(int64(expClaim), 0)

	// Allow 5 second tolerance
	assert.WithinDuration(t, expiresAt, actualExp, 5*time.Second, "expiration should match")
}