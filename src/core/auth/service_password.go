package authentication

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/MetaDandy/go-fiber-skeleton/constant"
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/enum"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/MetaDandy/go-fiber-skeleton/src/service/mail"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// passwordURepo defines user repository methods needed by passwordService
type passwordURepo interface {
	FindByEmail(email string) (model.User, error)
	ExistsByEmail(email string) error
	FindByID(id string) (model.User, error)
	UpdatePassword(userID string, passwordHash string) error
}

// PasswordService interface
type PasswordService interface {
	UserAuthProviders(email string) ([]string, *api_error.Error)
	SignUpPassword(input SignUpPassword) *api_error.Error
	LoginPassword(input LoginPassword) (string, string, *api_error.Error)
	ForgotPassword(input ForgotPassword) *api_error.Error
	ResetPassword(input ResetPassword) *api_error.Error
	ChangePassword(userID string, input ChangePassword, ip string, userAgent string) *api_error.Error
}

type passwordService struct {
	repo        Repo
	uRepo       passwordURepo
	mailService mail.EmailService
	appURL      string
}

func NewPasswordService(repo Repo, uRepo passwordURepo, mailService mail.EmailService, appURL string) PasswordService {
	return &passwordService{
		repo:        repo,
		uRepo:       uRepo,
		mailService: mailService,
		appURL:      appURL,
	}
}

func (s *passwordService) UserAuthProviders(email string) ([]string, *api_error.Error) {
	user, err := s.uRepo.FindByEmail(email)
	if err != nil {
		return []string{}, api_error.InternalServerError("User not found").WithErr(err)
	}

	providers := s.repo.UserAuthProviders(user.ID)
	if user.Password != nil {
		providers = append(providers, "password")
	}

	return providers, nil
}

func (s *passwordService) SignUpPassword(input SignUpPassword) *api_error.Error {
	if err := s.uRepo.ExistsByEmail(input.Email); err == nil {
		return api_error.Conflict("User already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return api_error.InternalServerError("Failed to hash password").WithErr(err)
	}

	hashed := string(hash)

	u := model.User{
		ID:            uuid.New(),
		Email:         input.Email,
		Password:      &hashed,
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
		return api_error.InternalServerError("Failed to create user").WithErr(err)
	}

	// Generar token de verificación de email
	token, err := mail.GenerateVerificationToken()
	if err != nil {
		log.Printf("failed to generate verification token: %v", err)
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
		return nil
	}

	// Enviar email de verificación
	ctx := context.Background()
	if err := s.mailService.SendVerificationEmail(ctx, u.Email, u.Name, token); err != nil {
		log.Printf("failed to send verification email to %s: %v", u.Email, err)
		return nil
	}

	return nil
}

func (s *passwordService) LoginPassword(input LoginPassword) (string, string, *api_error.Error) {
	// Buscar el usuario por email
	user, err := s.uRepo.FindByEmail(input.Email)
	if err != nil {
		// Log fallido
		al := model.AuthLog{
			ID:        uuid.New(),
			Event:     enum.LoginFailed,
			Ip:        input.Ip,
			UserAgent: input.UserAgent,
		}
		s.repo.CreateAuthLog(al)
		return "", "", api_error.Unauthorized("Invalid email or password")
	}

	// Validar que el usuario tiene contraseña
	if user.Password == nil {
		return "", "", api_error.Unauthorized("Invalid email or password")
	}

	// Comparar contraseña
	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(input.Password)); err != nil {
		// Log fallido
		al := model.AuthLog{
			ID:        uuid.New(),
			Event:     enum.LoginFailed,
			UserID:    user.ID,
			Ip:        input.Ip,
			UserAgent: input.UserAgent,
		}
		s.repo.CreateAuthLog(al)
		return "", "", api_error.Unauthorized("Invalid email or password")
	}

	// Validar que el email está verificado
	if !user.EmailVerified {
		al := model.AuthLog{
			ID:        uuid.New(),
			Event:     enum.EmailNotVerified,
			UserID:    user.ID,
			Ip:        input.Ip,
			UserAgent: input.UserAgent,
		}
		s.repo.CreateAuthLog(al)
		return "", "", api_error.Forbidden("Please verify your email before logging in")
	}

	permissions, err := s.repo.GetUserPermissions(user.ID)
	if err != nil {
		log.Printf("failed to retrieve user permissions: %v", err)
		return "", "", api_error.InternalServerError("Failed to retrieve permissions").WithErr(err)
	}

	// Generar tokens
	var roleIDStr string
	if user.RoleID != uuid.Nil {
		roleIDStr = user.RoleID.String()
	}
	accessToken, err := helper.GenerateJwt(user.ID.String(), user.Email, roleIDStr, permissions)
	if err != nil {
		log.Printf("failed to generate JWT token: %v", err)
		return "", "", api_error.InternalServerError("Failed to generate access token").WithErr(err)
	}

	refreshToken, err := helper.GenerateRefreshToken()
	if err != nil {
		log.Printf("failed to generate refresh token: %v", err)
		return "", "", api_error.InternalServerError("Failed to generate refresh token").WithErr(err)
	}

	// Registrar sesión
	session := model.Session{
		ID:               uuid.New(),
		Provider:         "password",
		RefreshTokenHash: mail.HashToken(refreshToken),
		ExpiresAt:        time.Now().Add(7 * 24 * time.Hour),
		Ip:               input.Ip,
		UserAgent:        input.UserAgent,
		UserID:           user.ID,
	}

	// Revocar sesiones anteriores y guardar la nueva
	_ = s.repo.RevokeAllUserSessions(user.ID)
	if err := s.repo.CreateSession(session); err != nil {
		log.Printf("failed to create session: %v", err)
		return "", "", api_error.InternalServerError("Failed to create session").WithErr(err)
	}

	// Crear auth log exitoso
	al := model.AuthLog{
		ID:        uuid.New(),
		Event:     enum.LoginSuccess,
		UserID:    user.ID,
		Ip:        input.Ip,
		UserAgent: input.UserAgent,
	}

	if err := s.repo.CreateAuthLog(al); err != nil {
		log.Printf("failed to create auth log for login: %v", err)
	}

	return accessToken, refreshToken, nil
}

func (s *passwordService) ForgotPassword(input ForgotPassword) *api_error.Error {
	// Buscar el usuario
	user, err := s.uRepo.FindByEmail(input.Email)
	if err != nil {
		return api_error.InternalServerError("An error occurred")
	}

	// Generar token
	token, err := mail.GeneratePasswordResetToken()
	if err != nil {
		log.Printf("failed to generate password reset token: %v", err)
		return api_error.InternalServerError("An error occurred")
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

	// EL REPO MANEJA LA TRANSACCIÓN
	if err := s.repo.SavePasswordResetTokenWithLog(prt, al); err != nil {
		log.Printf("failed to save password reset token and log: %v", err)
		return api_error.InternalServerError("An error occurred")
	}

	// Construir el link de reset
	appURL := s.getAppURL()
	resetLink := fmt.Sprintf("%s/reset-password/%s", appURL, token)

	// Enviar email
	ctx := context.Background()
	if err := s.mailService.SendPasswordReset(ctx, user.Email, user.Name, resetLink); err != nil {
		log.Printf("failed to send password reset email to %s: %v", user.Email, err)
		return api_error.InternalServerError("An error occurred")
	}

	return nil
}

func (s *passwordService) getAppURL() string {
	if s.appURL != "" {
		return s.appURL
	}
	return "http://localhost:3000"
}

func (s *passwordService) ResetPassword(input ResetPassword) *api_error.Error {
	tokenHash := mail.HashToken(input.Token)

	// Buscar el token en BD
	prt, err := s.repo.GetPasswordResetTokenByHash(tokenHash)
	if err != nil {
		return api_error.Unauthorized("Invalid or expired password reset token")
	}

	// Validar expiración
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

	// EL REPO MANEJA LA TRANSACCIÓN
	if err := s.repo.CompletePasswordReset(prt.UserID, string(passwordHash), al); err != nil {
		log.Printf("failed to complete password reset: %v", err)
		return api_error.InternalServerError("Could not reset password")
	}

	return nil
}

func (s *passwordService) ChangePassword(userID string, input ChangePassword, ip string, userAgent string) *api_error.Error {
	// Obtener el usuario
	user, err := s.uRepo.FindByID(userID)
	if err != nil {
		return api_error.Unauthorized("User not found")
	}

	// Validar que la contraseña actual sea correcta
	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(input.CurrentPassword)); err != nil {
		// Registrar intento fallido
		al := model.AuthLog{
			ID:        uuid.New(),
			Event:     enum.PasswordChangeFailure,
			UserID:    uuid.MustParse(userID),
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
		UserID:    uuid.MustParse(userID),
		Ip:        ip,
		UserAgent: userAgent,
	}

	// Actualizar contraseña
	if err := s.uRepo.UpdatePassword(userID, string(passwordHash)); err != nil {
		log.Printf("failed to update user password: %v", err)
		return api_error.InternalServerError("Could not change password")
	}

	// Guardar log
	if err := s.repo.CreateAuthLog(al); err != nil {
		log.Printf("failed to create auth log: %v", err)
	}

	return nil
}
