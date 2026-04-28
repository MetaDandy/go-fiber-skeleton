package authentication

import (
	"context"
	"math/rand"
	"os"
	"time"

	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/enum"
	"github.com/MetaDandy/go-fiber-skeleton/src/service/auth"
	"github.com/MetaDandy/go-fiber-skeleton/src/service/cookie"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/google/uuid"
)

type Handler interface {
	RegisterRoutes(router fiber.Router)
	UserAuthProviders(c fiber.Ctx) error
	SignUpPassword(c fiber.Ctx) error
	LoginPassword(c fiber.Ctx) error
	RefreshToken(c fiber.Ctx) error
	SendTestEmail(c fiber.Ctx) error
	VerifyEmail(c fiber.Ctx) error
	ResendVerificationEmail(c fiber.Ctx) error
	ForgotPassword(c fiber.Ctx) error
	ResetPassword(c fiber.Ctx) error
	ChangePassword(c fiber.Ctx) error
	Logout(c fiber.Ctx) error
	OAuthCallback(c fiber.Ctx) error
}

type handler struct {
	service   Service
	jwtMiddle fiber.Handler
}

func NewHandler(service Service, jwtMiddle fiber.Handler) Handler {
	return &handler{
		service:   service,
		jwtMiddle: jwtMiddle,
	}
}

func (h *handler) RegisterRoutes(router fiber.Router) {
	authRateLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: 15 * time.Minute,
		LimitReached: func(c fiber.Ctx) error {
			return api_error.TooManyRequests("Too many requests from this IP, please try again later")
		},
	})

	auth := router.Group("/auth")
	auth.Get("/providers/:email", h.UserAuthProviders)
	auth.Post("/signup", h.SignUpPassword)
	auth.Post("/signin", authRateLimiter, h.LoginPassword)
	auth.Post("/refresh", h.RefreshToken)
	auth.Post("/send-test-email", h.SendTestEmail)
	auth.Get("/verify-email/:token", h.VerifyEmail)
	auth.Get("/resend-verification-email/:email", h.ResendVerificationEmail)
	auth.Post("/forgot-password", authRateLimiter, h.ForgotPassword)
	auth.Post("/reset-password", h.ResetPassword)
	auth.Get("/login/:provider", h.OAuthLogin)
	auth.Get("/callback", h.OAuthCallback)

	protected := auth.Group("/", h.jwtMiddle)
	protected.Post("/change-password", h.ChangePassword)

	auth.Post("/logout", h.Logout)
}

func (h *handler) UserAuthProviders(c fiber.Ctx) error {
	email := c.Params("email")
	if email == "" {
		return api_error.BadRequest("Email parameter is required")
	}

	providers, err := h.service.UserAuthProviders(email)
	if err != nil {
		// Si el error es un *api_error.Error, retornarlo directamente
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		// Si es otro tipo de error, envolverlo
		return api_error.InternalServerError("Could not retrieve authentication providers").WithErr(err)
	}

	return c.JSON(fiber.Map{
		"providers": providers,
	})
}

func (h *handler) SignUpPassword(c fiber.Ctx) error {
	var input SignUpPassword
	if err := c.Bind().Body(&input); err != nil {
		return api_error.BadRequest("Invalid request body")
	}

	if err := input.Validate(); err != nil {
		return err
	}

	// Si el error es un *api_error.Error, retornarlo directamente
	if err := h.service.SignUpPassword(input); err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		// Si es otro tipo de error, envolverlo
		return api_error.InternalServerError(err.Error()).WithErr(err)
	}

	return c.JSON(fiber.Map{
		"message": "user created successfully",
	})
}

func (h *handler) LoginPassword(c fiber.Ctx) error {
	var input LoginPassword
	if err := c.Bind().Body(&input); err != nil {
		return api_error.BadRequest("Invalid request body")
	}

	// Extraer IP y User-Agent del contexto
	input.Ip, input.UserAgent = helper.GetClientDetails(c)

	if err := input.Validate(); err != nil {
		return err
	}

	token, refreshToken, err := h.service.LoginPassword(input)
	if err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		return api_error.InternalServerError("LoginPassword failed").WithErr(err)
	}

	// Setear cookies con los tokens
	cookie.SetAuthTokenCookie(c, token)
	cookie.SetRefreshTokenCookie(c, refreshToken)

	return c.JSON(fiber.Map{
		"message": "login successful",
	})
}

func (h *handler) RefreshToken(c fiber.Ctx) error {
	// 1. Obtener el refresh token de la cookie
	refreshToken := c.Cookies(cookie.CookieNameRefreshToken)
	if refreshToken == "" {
		return api_error.Unauthorized("No refresh token provided")
	}

	ip, userAgent := helper.GetClientDetails(c)

	// 2. Llamar al servicio para rotar el token
	newToken, newRefreshToken, err := h.service.RefreshToken(refreshToken, ip, userAgent)
	if err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		return api_error.Unauthorized("Could not refresh token").WithErr(err)
	}

	// 3. Setear las nuevas cookies
	cookie.SetAuthTokenCookie(c, newToken)
	cookie.SetRefreshTokenCookie(c, newRefreshToken)

	return c.JSON(fiber.Map{
		"message": "token refreshed successfully",
	})
}

func (h *handler) Logout(c fiber.Ctx) error {
	// 1. Obtener el refresh token para invalidar la sesión en DB
	refreshToken := cookie.GetRefreshTokenCookie(c)

	if refreshToken != "" {
		_ = h.service.Logout(refreshToken)
	}

	// 2. Limpiar todas las cookies de autenticación
	cookie.ClearAllAuthCookies(c)

	return c.JSON(fiber.Map{
		"message": "logout successful",
	})
}

func (h *handler) SendTestEmail(c fiber.Ctx) error {
	var input struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := c.Bind().Body(&input); err != nil {
		return api_error.BadRequest("Invalid request body")
	}

	if input.Email == "" {
		return api_error.BadRequest("Email is required")
	}

	if err := h.service.SendTestEmail(input.Email, input.Name); err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		return api_error.InternalServerError("Failed to send test email").WithErr(err)
	}

	return c.JSON(fiber.Map{
		"message": "test email sent successfully",
		"email":   input.Email,
	})
}

func (h *handler) VerifyEmail(c fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return api_error.BadRequest("Token is required")
	}

	if err := h.service.VerifyEmail(token); err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		return api_error.InternalServerError("Failed to verify email").WithErr(err)
	}

	return c.JSON(fiber.Map{
		"message": "email verified successfully",
	})
}

func (h *handler) ResendVerificationEmail(c fiber.Ctx) error {
	email := c.Params("email")

	if email == "" {
		return api_error.BadRequest("Email is required")
	}

	if err := h.service.ResendVerificationEmail(email); err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		return api_error.InternalServerError("Failed to resend verification email").WithErr(err)
	}

	return c.JSON(fiber.Map{
		"message": "verification email resent successfully",
	})
}

func (h *handler) ForgotPassword(c fiber.Ctx) error {
	var input ForgotPassword
	if err := c.Bind().Body(&input); err != nil {
		return api_error.BadRequest("Invalid request body: " + err.Error()).WithErr(err)
	}

	// Extraer IP del contexto/header
	input.Ip = c.IP()
	if input.Ip == "" {
		input.Ip = c.Get("X-Forwarded-For")
	}

	if err := input.Validate(); err != nil {
		return api_error.BadRequest("Invalid request body: " + err.Error()).WithErr(err)
	}

	if err := h.service.ForgotPassword(input); err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		return api_error.InternalServerError("Could not process forgot password request").WithErr(err)
	}

	// Retornar mensaje genérico por seguridad (no revelar si el email existe)
	return c.JSON(fiber.Map{
		"message": "If an account exists with that email, you will receive password reset instructions",
	})
}

func (h *handler) ResetPassword(c fiber.Ctx) error {
	var input ResetPassword
	if err := c.Bind().Body(&input); err != nil {
		return api_error.BadRequest("Invalid request body: " + err.Error()).WithErr(err)
	}

	// Extraer IP y User-Agent del contexto
	input.Ip = c.IP()
	if input.Ip == "" {
		input.Ip = c.Get("X-Forwarded-For")
	}
	input.UserAgent = c.Get("User-Agent")

	if err := input.Validate(); err != nil {
		return api_error.BadRequest("Invalid request body: " + err.Error()).WithErr(err)
	}

	if err := h.service.ResetPassword(input); err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		return api_error.InternalServerError("Could not reset password").WithErr(err)
	}

	return c.JSON(fiber.Map{
		"message": "password reset successfully",
	})
}

func (h *handler) ChangePassword(c fiber.Ctx) error {
	var input ChangePassword
	if err := c.Bind().Body(&input); err != nil {
		return api_error.BadRequest("Invalid request body: " + err.Error()).WithErr(err)
	}

	if err := input.Validate(); err != nil {
		return err
	}

	// Obtener userID desde el middleware JWT
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok || userIDStr == "" {
		return api_error.Unauthorized("User not authenticated")
	}

	// Validar que sea un UUID válido
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return api_error.BadRequest("Invalid user ID format in session")
	}

	// Extraer IP y User-Agent
	ip := c.IP()
	if ip == "" {
		ip = c.Get("X-Forwarded-For")
	}
	userAgent := c.Get("User-Agent")

	if err := h.service.ChangePassword(uid, input, ip, userAgent); err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		return api_error.InternalServerError("Could not change password").WithErr(err)
	}

	return c.JSON(fiber.Map{
		"message": "password changed successfully",
	})
}

// generateState genera un estado aleatorio para CSRF protection
func generateState() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	state := make([]byte, 32)
	for i := range state {
		state[i] = charset[rand.Intn(len(charset))]
	}
	return string(state)
}

// OAuthLogin inicia el flujo de login OAuth
// GET /auth/login/:provider
func (h *handler) OAuthLogin(c fiber.Ctx) error {
	provider := c.Params("provider")

	// Validar proveedor
	if !enum.IsValidAuthProvider(provider) {
		return api_error.BadRequest("Unsupported oauth provider")
	}

	// Cargar credenciales
	creds, err := auth.LoadCredentials(provider)
	if err != nil {
		return api_error.InternalServerError("Could not load provider credentials").WithErr(err)
	}

	// Generar estado para CSRF
	state := generateState()

	// Guardar estado en BD para validación en callback
	if err := h.service.SaveOAuthState(state, provider); err != nil {
		return api_error.InternalServerError("Failed to save OAuth state").WithErr(err)
	}

	// Obtener URL de autorización
	authURL, err := auth.GetAuthURL(provider, creds, os.Getenv("URI_REDIRECT"), state)
	if err != nil {
		return api_error.InternalServerError("Could not generate auth URL").WithErr(err)
	}

	return c.Redirect().To(authURL)
}

// OAuthCallback maneja el callback de OAuth desde el proveedor
// GET /auth/callback?code=xxx&state=yyy
func (h *handler) OAuthCallback(c fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		return api_error.BadRequest("Missing required parameters: code or state")
	}

	// Obtener el provider desde BD usando el state (CSRF token)
	provider, err := h.service.GetOAuthProviderByState(state)
	if err != nil {
		return api_error.BadRequest("Invalid or expired OAuth state")
	}

	// Validar que el provider sea válido
	if !enum.IsValidAuthProvider(provider) {
		return api_error.BadRequest("Invalid OAuth provider")
	}

	// Cargar credenciales del proveedor (desde servicio auth)
	creds, err := auth.LoadCredentials(provider)
	if err != nil {
		return api_error.InternalServerError("Could not load provider credentials").WithErr(err)
	}

	// Intercambiar código por token con el proveedor
	token, err := auth.ExchangeCode(provider, creds, os.Getenv("URI_REDIRECT"), code)
	if err != nil {
		return api_error.InternalServerError("Failed to exchange authorization code").WithErr(err)
	}

	// Obtener información del usuario desde el proveedor
	userInfo, err := auth.GetUserInfo(context.Background(), provider, token)
	if err != nil {
		return api_error.InternalServerError("Failed to retrieve user information").WithErr(err)
	}

	// Preparar DTO interno para el servicio
	oauthInput := OAuthCallbackInternal{
		Provider: provider,
		State:    state, // Pasar el state para validación y consumo atómico en el servicio
		UserInfo: OAuthUserInfo{
			ID:    userInfo.ID,
			Email: userInfo.Email,
			Name:  userInfo.Name,
			Image: userInfo.Image,
		},
		Ip:        c.IP(),
		UserAgent: c.Get("User-Agent"),
	}

	// Si la IP no se pudo obtener, intentar con X-Forwarded-For
	if oauthInput.Ip == "" {
		oauthInput.Ip = c.Get("X-Forwarded-For")
	}

	// Llamar al servicio para crear o hacer login del usuario
	// Ahora el servicio se encarga de validar el state ANTES de persistir
	jwtToken, refreshToken, err := h.service.OAuthCreateOrLogin(oauthInput)
	if err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		return api_error.InternalServerError("Authentication failed").WithErr(err)
	}

	// Setear cookies con los tokens
	cookie.SetAuthTokenCookie(c, jwtToken)
	cookie.SetRefreshTokenCookie(c, refreshToken)

	// Retornar respuesta exitosa
	return c.JSON(fiber.Map{
		"success":  true,
		"message":  "login successful",
		"provider": provider,
	})
}
