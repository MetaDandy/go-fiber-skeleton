package authentication

import (
	"context"
	"log"
	"time"

	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/MetaDandy/go-fiber-skeleton/src/service/mail"
	"github.com/google/uuid"
)

// emailURepo defines user repository methods needed by emailService
type emailURepo interface {
	FindByEmail(email string) (model.User, error)
}

// EmailService interface
type EmailService interface {
	VerifyEmail(token string) *api_error.Error
	ResendVerificationEmail(email string) *api_error.Error
}

type emailService struct {
	repo        Repo
	uRepo       emailURepo
	mailService mail.EmailService
}

func NewEmailService(repo Repo, uRepo emailURepo, mailService mail.EmailService) EmailService {
	return &emailService{
		repo:        repo,
		uRepo:       uRepo,
		mailService: mailService,
	}
}

// VerifyEmail verifica el email del usuario usando el token recibido
func (s *emailService) VerifyEmail(token string) *api_error.Error {
	if token == "" {
		return api_error.BadRequest("token is required")
	}

	// Hashear el token recibido
	tokenHash := mail.HashToken(token)

	// Buscar en BD verificando que no esté usado y no esté expirado
	evt, err := s.repo.GetEmailVerificationTokenByHash(tokenHash)
	if err != nil {
		return api_error.Unauthorized("Invalid or expired token").WithErr(err)
	}

	// Validar expiración (24 horas)
	if time.Since(evt.CreatedAt) > 24*time.Hour {
		return api_error.Unauthorized("Token has expired")
	}

	// Marcar usuario como verificado
	if err := s.repo.MarkEmailAsVerified(evt.UserID); err != nil {
		return api_error.InternalServerError("Could not verify email").WithErr(err)
	}

	// Marcar token como usado
	if err := s.repo.InvalidateOldEmailTokens(evt.UserID); err != nil {
		log.Printf("failed to invalidate tokens: %v", err)
	}

	return nil
}

// ResendVerificationEmail reenvía un email de verificación
func (s *emailService) ResendVerificationEmail(email string) *api_error.Error {
	if email == "" {
		return api_error.BadRequest("email is required")
	}

	// Buscar el usuario
	user, err := s.uRepo.FindByEmail(email)
	if err != nil {
		return api_error.NotFound("User not found").WithErr(err)
	}

	// Validar que el email no esté ya verificado
	if user.EmailVerified {
		return api_error.BadRequest("Email is already verified")
	}

	// Invalidar tokens anteriores
	if err := s.repo.InvalidateOldEmailTokens(user.ID); err != nil {
		log.Printf("failed to invalidate old tokens: %v", err)
	}

	// Generar nuevo token
	token, err := mail.GenerateVerificationToken()
	if err != nil {
		return api_error.InternalServerError("Could not generate token").WithErr(err)
	}

	// Guardar el token
	tokenHash := mail.HashToken(token)
	evt := model.EmailVerificationToken{
		ID:        uuid.New(),
		TokenHash: tokenHash,
		UserID:    user.ID,
	}

	if err := s.repo.SaveEmailVerificationToken(evt); err != nil {
		return api_error.InternalServerError("Could not save token").WithErr(err)
	}

	// Enviar email
	ctx := context.Background()
	if err := s.mailService.SendVerificationEmail(ctx, user.Email, user.Name, token); err != nil {
		return api_error.InternalServerError("Could not send email").WithErr(err)
	}

	return nil
}
