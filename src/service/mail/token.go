package mail

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// GenerateVerificationToken genera un token aleatorio de 32 bytes en hex
// Retorna un token que se envía al usuario por email
func GenerateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GeneratePasswordResetToken genera un token aleatorio de 32 bytes en hex
// Similar a GenerateVerificationToken pero con nombre más descriptivo
func GeneratePasswordResetToken() (string, error) {
	return GenerateVerificationToken()
}

// HashToken hashea un token usando SHA256
// Se almacena en BD para comparar después (nunca almacenar el token sin hash)
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
