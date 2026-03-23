package authentication

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/MetaDandy/go-fiber-skeleton/constant"
	"github.com/MetaDandy/go-fiber-skeleton/src/enum"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/MetaDandy/go-fiber-skeleton/src/service/mail"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	UserAuthProviders(email string) ([]string, error)
	SignUpPassword(input SignUpPassword) error
	SendTestEmail(email, name string) error
	VerifyEmail(token string) error
	ResendVerificationEmail(email string) error
}

type uRepo interface {
	FindByEmail(email string) (model.User, error)
	ExistsByEmail(email string) error
}

type service struct {
	repo        Repo
	uRepo       uRepo
	mailService mail.EmailService
}

func NewService(repo Repo, uRepo uRepo, mailService mail.EmailService) Service {
	return &service{
		repo:        repo,
		uRepo:       uRepo,
		mailService: mailService,
	}
}

func (s *service) UserAuthProviders(email string) ([]string, error) {
	user, err := s.uRepo.FindByEmail(email)
	if err != nil {
		return []string{}, err
	}

	providers := s.repo.UserAuthProviders(user.ID.String())
	if user.Password != "" {
		providers = append(providers, "password")
	}

	return providers, nil
}

func (s *service) SignUpPassword(input SignUpPassword) error {
	if err := s.uRepo.ExistsByEmail(input.Email); err == nil {
		return fmt.Errorf("%s already exist", input.Email)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	u := model.User{
		ID:            uuid.New(),
		Email:         input.Email,
		Password:      string(hash),
		EmailVerified: false,
		RoleID:        constant.GenericID,
	}

	al := model.AuthLog{
		ID:        uuid.New(),
		Event:     enum.SignUpSuccess,
		UserID:    u.ID,
		Ip:        input.Ip,
		UserAgent: input.UserAgent,
	}

	if err := s.repo.Create(u, al, nil); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Generar token de verificación de email
	token, err := mail.GenerateVerificationToken()
	if err != nil {
		log.Printf("failed to generate verification token: %v", err)
		// No retornar error, el signup ya fue exitoso
		// El usuario puede solicitar resend después
		return nil
	}

	// Hashear y guardar el token en BD
	tokenHash := mail.HashToken(token)
	evt := model.EmailVerificationToken{
		ID:        uuid.New(),
		TokenHash: tokenHash,
		UserID:    u.ID,
	}

	if err := s.repo.SaveEmailVerificationToken(evt); err != nil {
		log.Printf("failed to save verification token: %v", err)
		// No retornar error, el signup ya fue exitoso
		return nil
	}

	// Enviar email de verificación (agnóstico - funciona con Mailpit o Resend)
	ctx := context.Background()
	if err := s.mailService.SendVerificationEmail(ctx, u.Email, u.Name, token); err != nil {
		log.Printf("failed to send verification email to %s: %v", u.Email, err)
		// No retornar error, el signup ya fue exitoso
		// El usuario puede solicitar resend después
		return nil
	}

	return nil
}

// SendTestEmail envía un email de prueba (para testing durante desarrollo)
func (s *service) SendTestEmail(email, name string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	if name == "" {
		name = "User"
	}

	ctx := context.Background()

	// Enviar email de bienvenida como prueba
	if err := s.mailService.SendWelcome(ctx, email, name); err != nil {
		return fmt.Errorf("failed to send test email: %w", err)
	}

	log.Printf("Test email sent to %s", email)
	return nil
}

// VerifyEmail verifica el email del usuario usando el token recibido
func (s *service) VerifyEmail(token string) error {
	if token == "" {
		return api_error.BadRequest("token is required")
	}

	// Hashear el token recibido
	tokenHash := mail.HashToken(token)

	// Buscar en BD verificando que no esté usado y no esté expirado
	evt, err := s.repo.GetEmailVerificationTokenByHash(tokenHash)
	if err != nil {
		return api_error.Unauthorized("Invalid or expired token")
	}

	// Validar expiración (24 horas)
	if time.Since(evt.CreatedAt) > 24*time.Hour {
		return api_error.Unauthorized("Token has expired")
	}

	// Marcar usuario como verificado
	if err := s.repo.MarkEmailAsVerified(evt.UserID); err != nil {
		log.Printf("failed to mark email as verified: %v", err)
		return api_error.InternalServerError("Could not verify email")
	}

	// Marcar token como usado
	if err := s.repo.InvalidateOldEmailTokens(evt.UserID); err != nil {
		log.Printf("failed to invalidate tokens: %v", err)
		// No retornar error, el email ya está verificado
	}

	return nil
}

// ResendVerificationEmail reenvía un email de verificación
func (s *service) ResendVerificationEmail(email string) error {
	if email == "" {
		return api_error.BadRequest("email is required")
	}

	// Buscar el usuario
	user, err := s.uRepo.FindByEmail(email)
	if err != nil {
		return api_error.Unauthorized("User not found")
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
		log.Printf("failed to generate verification token: %v", err)
		return api_error.InternalServerError("Could not generate token")
	}

	// Guardar el token
	tokenHash := mail.HashToken(token)
	evt := model.EmailVerificationToken{
		ID:        uuid.New(),
		TokenHash: tokenHash,
		UserID:    user.ID,
	}

	if err := s.repo.SaveEmailVerificationToken(evt); err != nil {
		log.Printf("failed to save verification token: %v", err)
		return api_error.InternalServerError("Could not save token")
	}

	// Enviar email
	ctx := context.Background()
	if err := s.mailService.SendVerificationEmail(ctx, user.Email, user.Name, token); err != nil {
		log.Printf("failed to send verification email: %v", err)
		return api_error.InternalServerError("Could not send email")
	}

	return nil
}
