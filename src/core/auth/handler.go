package authentication

import (
	"time"

	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/service/cookie"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

// authRateLimiter is a package-level rate limiter shared across auth routes
var authRateLimiter = limiter.New(limiter.Config{
	Max:        5,
	Expiration: 15 * time.Minute,
	LimitReached: func(c fiber.Ctx) error {
		return api_error.TooManyRequests("Too many requests from this IP, please try again later")
	},
})

type Handler interface {
	RegisterRoutes(router fiber.Router)
	UserAuthProviders(c fiber.Ctx) error
	SignUpPassword(c fiber.Ctx) error
	LoginPassword(c fiber.Ctx) error
	RefreshToken(c fiber.Ctx) error
	VerifyEmail(c fiber.Ctx) error
	ResendVerificationEmail(c fiber.Ctx) error
	ForgotPassword(c fiber.Ctx) error
	ResetPassword(c fiber.Ctx) error
	ChangePassword(c fiber.Ctx) error
	Logout(c fiber.Ctx) error
	OAuthLogin(c fiber.Ctx) error
	OAuthCallback(c fiber.Ctx) error
}

type handler struct {
	passwordSvc PasswordService
	oauthSvc    OAuthService
	emailSvc    EmailService
	sessionSvc  SessionService
	jwtMiddle   fiber.Handler
}

func NewHandler(passwordSvc PasswordService, oauthSvc OAuthService, emailSvc EmailService, sessionSvc SessionService, jwtMiddle fiber.Handler) Handler {
	return &handler{
		passwordSvc: passwordSvc,
		oauthSvc:    oauthSvc,
		emailSvc:    emailSvc,
		sessionSvc:  sessionSvc,
		jwtMiddle:   jwtMiddle,
	}
}

func (h *handler) RegisterRoutes(router fiber.Router) {
	auth := router.Group("/auth")
	auth.Get("/providers/:email", h.UserAuthProviders)
	auth.Post("/signup", authRateLimiter, h.SignUpPassword)
	auth.Post("/signin", authRateLimiter, h.LoginPassword)
	auth.Post("/refresh", h.RefreshToken)
	auth.Get("/verify-email/:token", authRateLimiter, h.VerifyEmail)
	auth.Get("/resend-verification-email/:email", h.ResendVerificationEmail)
	auth.Post("/forgot-password", authRateLimiter, h.ForgotPassword)
	auth.Post("/reset-password", authRateLimiter, h.ResetPassword)
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

	providers, err := h.passwordSvc.UserAuthProviders(email)
	if err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
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

	if err := h.passwordSvc.SignUpPassword(input); err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
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

	input.Ip, input.UserAgent = helper.GetClientDetails(c)

	if err := input.Validate(); err != nil {
		return err
	}

	token, refreshToken, err := h.passwordSvc.LoginPassword(input)
	if err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		return api_error.InternalServerError("LoginPassword failed").WithErr(err)
	}

	cookie.SetAuthTokenCookie(c, token)
	cookie.SetRefreshTokenCookie(c, refreshToken)

	return c.JSON(fiber.Map{
		"message": "login successful",
	})
}

func (h *handler) RefreshToken(c fiber.Ctx) error {
	refreshToken := c.Cookies(cookie.CookieNameRefreshToken)
	if refreshToken == "" {
		return api_error.Unauthorized("No refresh token provided")
	}

	ip, userAgent := helper.GetClientDetails(c)

	newToken, newRefreshToken, err := h.sessionSvc.RefreshToken(refreshToken, ip, userAgent)
	if err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		return api_error.Unauthorized("Could not refresh token").WithErr(err)
	}

	cookie.SetAuthTokenCookie(c, newToken)
	cookie.SetRefreshTokenCookie(c, newRefreshToken)

	return c.JSON(fiber.Map{
		"message": "token refreshed successfully",
	})
}

func (h *handler) Logout(c fiber.Ctx) error {
	refreshToken := cookie.GetRefreshTokenCookie(c)

	if refreshToken != "" {
		_ = h.sessionSvc.Logout(refreshToken)
	}

	cookie.ClearAllAuthCookies(c)

	return c.JSON(fiber.Map{
		"message": "logout successful",
	})
}

func (h *handler) VerifyEmail(c fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return api_error.BadRequest("Token is required")
	}

	if err := h.emailSvc.VerifyEmail(token); err != nil {
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

	if err := h.emailSvc.ResendVerificationEmail(email); err != nil {
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

	input.Ip, _ = helper.GetClientDetails(c)

	if err := input.Validate(); err != nil {
		return api_error.BadRequest("Invalid request body: " + err.Error()).WithErr(err)
	}

	if err := h.passwordSvc.ForgotPassword(input); err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		return api_error.InternalServerError("Could not process forgot password request").WithErr(err)
	}

	return c.JSON(fiber.Map{
		"message": "If an account exists with that email, you will receive password reset instructions",
	})
}

func (h *handler) ResetPassword(c fiber.Ctx) error {
	var input ResetPassword
	if err := c.Bind().Body(&input); err != nil {
		return api_error.BadRequest("Invalid request body: " + err.Error()).WithErr(err)
	}

	input.Ip, input.UserAgent = helper.GetClientDetails(c)

	if err := input.Validate(); err != nil {
		return api_error.BadRequest("Invalid request body: " + err.Error()).WithErr(err)
	}

	if err := h.passwordSvc.ResetPassword(input); err != nil {
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

	userIDStr, ok := c.Locals("user_id").(string)
	if !ok || userIDStr == "" {
		return api_error.Unauthorized("User not authenticated")
	}

	ip, userAgent := helper.GetClientDetails(c)

	if err := h.passwordSvc.ChangePassword(userIDStr, input, ip, userAgent); err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		return api_error.InternalServerError("Could not change password").WithErr(err)
	}

	return c.JSON(fiber.Map{
		"message": "password changed successfully",
	})
}

// OAuthLogin initiates the OAuth login flow
func (h *handler) OAuthLogin(c fiber.Ctx) error {
	provider := c.Params("provider")

	// Call OAuth service to get the authorization URL
	authURL, err := h.oauthSvc.OAuthLogin(provider)
	if err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		return api_error.InternalServerError("Could not initiate OAuth login").WithErr(err)
	}

	return c.Redirect().To(authURL)
}

// OAuthCallback handles the OAuth callback from the provider
func (h *handler) OAuthCallback(c fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		return api_error.BadRequest("Missing required parameters: code or state")
	}

	ip, userAgent := helper.GetClientDetails(c)

	// Call OAuth service to handle the callback
	jwtToken, refreshToken, err := h.oauthSvc.OAuthCallback(code, state, ip, userAgent)
	if err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		return api_error.InternalServerError("Authentication failed").WithErr(err)
	}

	// Set cookies with the tokens
	cookie.SetAuthTokenCookie(c, jwtToken)
	cookie.SetRefreshTokenCookie(c, refreshToken)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "login successful",
	})
}
