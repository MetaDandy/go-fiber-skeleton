package authentication

import (
	"time"

	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/MetaDandy/go-fiber-skeleton/src/service/mail"
	"github.com/google/uuid"
)

// sessionURepo defines user repository methods needed by sessionService
type sessionURepo interface {
	FindByID(id string) (model.User, error)
}

// SessionService interface
type SessionService interface {
	RefreshToken(refreshToken string, ip string, userAgent string) (string, string, error)
	Logout(refreshToken string) error
}

type sessionService struct {
	repo  Repo
	uRepo sessionURepo
}

func NewSessionService(repo Repo, uRepo sessionURepo) SessionService {
	return &sessionService{
		repo:  repo,
		uRepo: uRepo,
	}
}

// RefreshToken valida un refresh token y emite un nuevo par (Token Rotation)
func (s *sessionService) RefreshToken(refreshToken string, ip string, userAgent string) (string, string, error) {
	// 1. Hashear el token recibido
	tokenHash := mail.HashToken(refreshToken)

	// 2. Buscar la sesión por el hash
	session, err := s.repo.GetSessionByHash(tokenHash)
	if err != nil {
		return "", "", api_error.Unauthorized("Invalid or expired session")
	}

	// 3. Validar expiración
	if time.Now().After(session.ExpiresAt) {
		_ = s.repo.RevokeSession(session.ID)
		return "", "", api_error.Unauthorized("Session expired")
	}

	// 4. Obtener usuario
	user, err := s.uRepo.FindByID(session.UserID.String())
	if err != nil {
		return "", "", api_error.Unauthorized("User not found")
	}

	permissions, err := s.repo.GetUserPermissions(user.ID.String())
	if err != nil {
		return "", "", api_error.InternalServerError("Failed to get user permissions")
	}

	var roleIDStr string
	if user.RoleID != uuid.Nil {
		roleIDStr = user.RoleID.String()
	}

	// 5. ROTACIÓN: Generar nuevos tokens y reemplazar la sesión
	accessToken, err := helper.GenerateJwt(user.ID.String(), user.Email, roleIDStr, permissions)
	if err != nil {
		return "", "", api_error.InternalServerError("Failed to generate access token")
	}

	newRefreshToken, err := helper.GenerateRefreshToken()
	if err != nil {
		return "", "", api_error.InternalServerError("Failed to generate refresh token")
	}

	// Actualizar sesión actual
	newSession := model.Session{
		ID:               uuid.New(),
		Provider:         session.Provider,
		RefreshTokenHash: mail.HashToken(newRefreshToken),
		ExpiresAt:        time.Now().Add(7 * 24 * time.Hour),
		Ip:               ip,
		UserAgent:        userAgent,
		UserID:           user.ID,
	}

	// Revocar la sesión vieja y crear la nueva
	_ = s.repo.RevokeSession(session.ID)
	if err := s.repo.CreateSession(newSession); err != nil {
		return "", "", api_error.InternalServerError("Failed to rotate session")
	}

	return accessToken, newRefreshToken, nil
}

// Logout invalida la sesión de forma inmediata
func (s *sessionService) Logout(refreshToken string) error {
	if refreshToken == "" {
		return nil
	}

	tokenHash := mail.HashToken(refreshToken)
	session, err := s.repo.GetSessionByHash(tokenHash)
	if err != nil {
		return nil // Ya no existe o ya está revocada
	}

	return s.repo.RevokeSession(session.ID)
}
