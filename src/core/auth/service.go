package authentication

import (
	"context"
	"fmt"
	"log"
	"os"
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
	ForgotPassword(input ForgotPassword) error
	ResetPassword(input ResetPassword) error
	ChangePassword(userID uuid.UUID, input ChangePassword, ip string, userAgent string) error
}

type uRepo interface {
	FindByEmail(email string) (model.User, error)
	ExistsByEmail(email string) error
	FindByID(id string) (model.User, error)
	UpdatePassword(userID string, passwordHash string) error
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

// ForgotPassword genera un token de reset y lo envía por mail
func (s *service) ForgotPassword(input ForgotPassword) error {
	// Buscar el usuario
	user, err := s.uRepo.FindByEmail(input.Email)
	if err != nil {
		// No revelar si el email existe por seguridad
		return api_error.InternalServerError("Error")
	}

	// Generar token
	token, err := mail.GeneratePasswordResetToken()
	if err != nil {
		log.Printf("failed to generate password reset token: %v", err)
		return api_error.InternalServerError("Error")
	}

	// Construir estructuras para guardar
	tokenHash := mail.HashToken(token)
	prt := model.PasswordResetToken{
		ID:        uuid.New(),
		TokenHash: tokenHash,
		UserID:    user.ID,
	}

	al := model.AuthLog{
		ID:     uuid.New(),
		Event:  enum.PasswordReset,
		UserID: user.ID,
		Ip:     input.Ip,
	}

	// EL REPO MANEJA LA TRANSACCIÓN: guarda token + log + invalida tokens anteriores
	if err := s.repo.SavePasswordResetTokenWithLog(prt, al); err != nil {
		log.Printf("failed to save password reset token and log: %v", err)
		return api_error.InternalServerError("Error")
	}

	// Construir el link de reset
	// El token sin hash se envía en el email, el usuario lo usará en reset-password
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:3000"
	}
	resetLink := fmt.Sprintf("%s/reset-password/%s", appURL, token)

	// Enviar email con el link
	ctx := context.Background()
	if err := s.mailService.SendPasswordReset(ctx, user.Email, user.Name, resetLink); err != nil {
		log.Printf("failed to send password reset email to %s: %v", user.Email, err)
		return api_error.InternalServerError("Error")
	}

	return nil
}

// ResetPassword cambia la contraseña usando un token válido
// Se llama cuando el usuario hace click en el link del email
func (s *service) ResetPassword(input ResetPassword) error {
	tokenHash := mail.HashToken(input.Token)

	// Buscar el token en BD
	prt, err := s.repo.GetPasswordResetTokenByHash(tokenHash)
	if err != nil {
		return api_error.Unauthorized("Invalid or expired password reset token")
	}

	// Validar expiración (1 hora)
	if time.Since(prt.CreatedAt) > 1*time.Hour {
		return api_error.Unauthorized("Password reset token has expired")
	}

	// Hash la nueva contraseña
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("failed to hash password: %v", err)
		return api_error.InternalServerError("Could not process password reset")
	}

	// Construir auth log
	al := model.AuthLog{
		ID:        uuid.New(),
		Event:     enum.PasswordResetSuccess,
		UserID:    prt.UserID,
		Ip:        input.Ip,
		UserAgent: input.UserAgent,
	}

	// EL REPO MANEJA LA TRANSACCIÓN: actualizar contraseña en user + invalidar token + guardar log
	userIDStr := prt.UserID.String()
	if err := s.repo.CompletePasswordReset(userIDStr, string(passwordHash), al); err != nil {
		log.Printf("failed to complete password reset: %v", err)
		return api_error.InternalServerError("Could not reset password")
	}

	return nil
}

// ChangePassword cambia la contraseña del usuario autenticado
// Se llama cuando el usuario pide cambiar su contraseña (no olvido)
func (s *service) ChangePassword(userID uuid.UUID, input ChangePassword, ip string, userAgent string) error {
	// Obtener el usuario (por user repo)
	userIDStr := userID.String()
	user, err := s.uRepo.FindByID(userIDStr)
	if err != nil {
		return api_error.Unauthorized("User not found")
	}

	// Validar que la contraseña actual sea correcta
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.CurrentPassword)); err != nil {
		// Registrar intento fallido en auth logs
		al := model.AuthLog{
			ID:        uuid.New(),
			Event:     enum.PasswordChangeFailure,
			UserID:    userID,
			Ip:        ip,
			UserAgent: userAgent,
		}
		s.repo.CreateAuthLog(al)

		return api_error.Unauthorized("Current password is incorrect")
	}

	// Hash la nueva contraseña
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("failed to hash password: %v", err)
		return api_error.InternalServerError("Could not process password change")
	}

	// Construir auth log exitoso
	al := model.AuthLog{
		ID:        uuid.New(),
		Event:     enum.PasswordChangeSuccess,
		UserID:    userID,
		Ip:        ip,
		UserAgent: userAgent,
	}

	// EL USER REPO MANEJA: actualizar contraseña + guardar log con transacción
	if err := s.uRepo.UpdatePassword(userIDStr, string(passwordHash)); err != nil {
		log.Printf("failed to update user password: %v", err)
		return api_error.InternalServerError("Could not change password")
	}

	// Guardar log de cambio exitoso (en auth repo - es log)
	if err := s.repo.CreateAuthLog(al); err != nil {
		log.Printf("failed to create auth log: %v", err)
		// Si falla el log, no rollback el cambio de contraseña (es crítico el cambio)
	}

	return nil
}
