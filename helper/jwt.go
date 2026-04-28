package helper

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJwt(user_id, email, role string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = user_id
	claims["email"] = email
	claims["role"] = role
	claims["exp"] = time.Now().Add(time.Hour).Unix() // Token expira en 1 hora

	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// GenerateRefreshToken genera un token opaco aleatorio de 32 bytes
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
