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
	"gorm.io/gorm"
)

type Service interface {
	UserAuthProviders(email string) ([]string, error)
	SignUpPassword(input SignUpPassword) error
	LoginPassword(input LoginPassword) (string, string, error)
	VerifyEmail(token string) error
	ResendVerificationEmail(email string) error
	ForgotPassword(input ForgotPassword) error
	ResetPassword(input ResetPassword) error
	ChangePassword(userID uuid.UUID, input ChangePassword, ip string, userAgent string) error
	OAuthCreateOrLogin(input OAuthCallbackInternal) (string, string, error)
	RefreshToken(refreshToken string, ip string, userAgent string) (string, string, error)
	Logout(refreshToken string) error
	SaveOAuthState(state, provider string) error
	GetOAuthProviderByState(state string) (string, error)
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
	appURL      string
}

func NewService(repo Repo, uRepo uRepo, mailService mail.EmailService, appURL string) Service {
	return &service{
		repo:        repo,
		uRepo:       uRepo,
		mailService: mailService,
		appURL:      appURL,
	}
}

func (s *service) UserAuthProviders(email string) ([]string, error) {
	user, err := s.uRepo.FindByEmail(email)
	if err != nil {
		return []string{}, err
	}

	providers := s.repo.UserAuthProviders(user.ID.String())
	if user.Password != nil {
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
	appURL := s.appURL
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
	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(input.CurrentPassword)); err != nil {
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

// LoginPassword autentica un usuario con email y contraseña
// Retorna AccessToken y RefreshToken si es exitoso
func (s *service) LoginPassword(input LoginPassword) (string, string, error) {
	// Buscar el usuario por email
	user, err := s.uRepo.FindByEmail(input.Email)
	if err != nil {
		// Log fallido (pero respuesta genérica por seguridad)
		al := model.AuthLog{
			ID:        uuid.New(),
			Event:     enum.LoginFailed,
			Ip:        input.Ip,
			UserAgent: input.UserAgent,
		}
		s.repo.CreateAuthLog(al)
		return "", "", api_error.Unauthorized("Invalid email or password")
	}

	// Validar que el usuario tiene contraseña (no es solo OAuth)
	if user.Password == nil {
		return "", "", api_error.Unauthorized("Invalid email or password")
	}

	// Comparar contraseña con el hash guardado
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
		// Log intento sin email verificado
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

	permissions, err := s.repo.GetUserPermissions(user.ID.String())
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
		RefreshTokenHash: mail.HashToken(refreshToken), // Usamos helper de hash de mail por simplicidad
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

// OAuthCreateOrLogin maneja el flujo OAuth: crea usuario si no existe o hace login si existe
// Email siempre está verificado en OAuth (confiamos en el proveedor)
func (s *service) OAuthCreateOrLogin(input OAuthCallbackInternal) (string, string, error) {
	// 1. Validar que el email no esté vacío
	if input.UserInfo.Email == "" {
		return "", "", api_error.BadRequest("No email provided by OAuth provider")
	}

	// 2. Validar que el provider sea válido
	if !enum.IsValidAuthProvider(input.Provider) {
		return "", "", api_error.BadRequest("Invalid OAuth provider")
	}

	// 4. Buscar si el usuario existe por email
	user, err := s.uRepo.FindByEmail(input.UserInfo.Email)

	if err != nil && err != gorm.ErrRecordNotFound {
		// Error de BD inesperado
		return "", "", api_error.InternalServerError("Could not verify user").WithErr(err)
	}

	// CASO A: Email NO existe - SIGNUP con OAuth
	if err == gorm.ErrRecordNotFound {
		return s.oauthSignUp(input)
	}

	// CASO B: Email EXISTS - LOGIN con OAuth (puede agregar new provider o usar existente)
	return s.oauthLogin(input, user)
}

// oauthSignUp crea un nuevo usuario con OAuth
func (s *service) oauthSignUp(input OAuthCallbackInternal) (string, string, error) {
	// Crear usuario con email verificado (confiamos en OAuth provider)
	user := model.User{
		ID:              uuid.New(),
		Email:           input.UserInfo.Email,
		Name:            input.UserInfo.Name,
		Picture:         input.UserInfo.Image,
		Password:        nil,
		EmailVerified:   true,
		EmailVerifiedAt: ptrTime(time.Now()),
		RoleID:          constant.GenericID,
	}

	// Crear auth log
	authLog := model.AuthLog{
		ID:        uuid.New(),
		Event:     enum.OAuthLogin,
		UserID:    user.ID,
		Ip:        input.Ip,
		UserAgent: input.UserAgent,
	}

	// Crear auth provider
	authProvider := model.AuthProvider{
		ID:             uuid.New(),
		Provider:       input.Provider,
		ProviderUserID: input.UserInfo.ID,
		UserID:         user.ID,
	}

	// Transacción: crear user + authlog + authprovider + consumir state
	if err := s.repo.CreateOAuthUser(user, authLog, authProvider, input.State); err != nil {
		log.Printf("failed to create OAuth user: %v", err)
		return "", "", api_error.InternalServerError("Failed to create user").WithErr(err)
	}

	return s.generateAndSaveSession(user, input.Provider, input.Ip, input.UserAgent)
}

// oauthLogin maneja el login de un usuario existente con OAuth
func (s *service) oauthLogin(input OAuthCallbackInternal, user model.User) (string, string, error) {
	// Verificar si el usuario ya tiene este proveedor
	err := s.repo.GetOAuthProvider(user.ID, input.Provider)

	if err != nil && err != gorm.ErrRecordNotFound {
		// Error inesperado de BD
		return "", "", api_error.InternalServerError("Could not verify auth provider").WithErr(err)
	}

	// CASO B1: El proveedor ya existe para este usuario - Consumir state + crear log
	if err == nil {
		authLog := model.AuthLog{
			ID:        uuid.New(),
			Event:     enum.OAuthLogin,
			UserID:    user.ID,
			Ip:        input.Ip,
			UserAgent: input.UserAgent,
		}

		// Transacción atómica: consumir state + crear log
		if err := s.repo.ConsumeOAuthStateAndLog(input.State, input.Provider, authLog); err != nil {
			log.Printf("failed to consume OAuth state and create log: %v", err)
			// No fallar el login si falla el log/state consumption (state already validated upstream)
		}

		fmt.Printf("✅ [OAuth Login - Existing Provider] Login exitoso\n")
		fmt.Printf("   Email: %s\n", user.Email)
		fmt.Printf("   Provider: %s\n", input.Provider)

	} else {
		// CASO B2: El proveedor NO existe para este usuario - Agregar nuevo proveedor + log
		authProvider := model.AuthProvider{
			ID:             uuid.New(),
			Provider:       input.Provider,
			ProviderUserID: input.UserInfo.ID,
			UserID:         user.ID,
		}

		authLog := model.AuthLog{
			ID:        uuid.New(),
			Event:     enum.OAuthLogin,
			UserID:    user.ID,
			Ip:        input.Ip,
			UserAgent: input.UserAgent,
		}

		if err := s.repo.AddOAuthProviderToUser(user.ID, authProvider, authLog, input.State, input.Provider); err != nil {
			log.Printf("failed to add OAuth provider to user: %v", err)
			return "", "", api_error.InternalServerError("Failed to authenticate").WithErr(err)
		}

		fmt.Printf("✅ [OAuth Login - New Provider] Nuevo proveedor agregado\n")
		fmt.Printf("   Email: %s\n", user.Email)
		fmt.Printf("   New Provider: %s\n", input.Provider)
	}

	return s.generateAndSaveSession(user, input.Provider, input.Ip, input.UserAgent)
}

// generateAndSaveSession es un helper para centralizar la creación de tokens y sesión
func (s *service) generateAndSaveSession(user model.User, provider, ip, userAgent string) (string, string, error) {
	permissions, err := s.repo.GetUserPermissions(user.ID.String())
	if err != nil {
		log.Printf("failed to get user permissions: %v", err)
		return "", "", api_error.InternalServerError("Failed to get user permissions").WithErr(err)
	}

	var roleIDStr string
	if user.RoleID != uuid.Nil {
		roleIDStr = user.RoleID.String()
	}

	accessToken, err := helper.GenerateJwt(user.ID.String(), user.Email, roleIDStr, permissions)
	if err != nil {
		log.Printf("failed to generate access token: %v", err)
		return "", "", api_error.InternalServerError("Failed to generate access token").WithErr(err)
	}

	refreshToken, err := helper.GenerateRefreshToken()
	if err != nil {
		log.Printf("failed to generate refresh token: %v", err)
		return "", "", api_error.InternalServerError("Failed to generate refresh token").WithErr(err)
	}

	session := model.Session{
		ID:               uuid.New(),
		Provider:         provider,
		RefreshTokenHash: mail.HashToken(refreshToken),
		ExpiresAt:        time.Now().Add(7 * 24 * time.Hour),
		Ip:               ip,
		UserAgent:        userAgent,
		UserID:           user.ID,
	}

	// Revocar sesiones anteriores y guardar la nueva
	_ = s.repo.RevokeAllUserSessions(user.ID)
	if err := s.repo.CreateSession(session); err != nil {
		log.Printf("failed to create session: %v", err)
		return "", "", api_error.InternalServerError("Failed to register session").WithErr(err)
	}

	return accessToken, refreshToken, nil
}

// Logout invalida la sesión de forma inmediata
func (s *service) Logout(refreshToken string) error {
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

// ptrTime es un helper para crear un puntero a time.Time
func ptrTime(t time.Time) *time.Time {
	return &t
}

// SaveOAuthState guarda un estado de OAuth en la base de datos
func (s *service) SaveOAuthState(state, provider string) error {
	return s.repo.SaveOAuthState(state, provider)
}

// GetOAuthProviderByState obtiene el provider asociado a un state
func (s *service) GetOAuthProviderByState(state string) (string, error) {
	return s.repo.GetOAuthProviderByState(state)
}

// RefreshToken valida un refresh token y emite un nuevo par (Token Rotation)
func (s *service) RefreshToken(refreshToken, ip, userAgent string) (string, string, error) {
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

	// Actualizar sesión actual (Data API style refresh)
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
