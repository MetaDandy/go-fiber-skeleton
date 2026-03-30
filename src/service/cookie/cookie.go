package cookie

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
)

// Cookie names constants
const (
	CookieNameOAuthState   = "oauth_state"
	CookieNameAuthToken    = "auth_token"
	CookieNameRefreshToken = "refresh_token"
	CookieNameSessionToken = "session_token"
)

// CookieConfig holds the configuration for cookies based on environment
type CookieConfig struct {
	Secure   bool
	SameSite string
	Domain   string
}

// GetCookieConfig returns the appropriate cookie configuration based on environment
func GetCookieConfig() CookieConfig {
	env := os.Getenv("ENVIROMENT")

	config := CookieConfig{
		Domain: os.Getenv("COOKIE_DOMAIN"),
	}

	if env == "production" {
		config.Secure = true
		config.SameSite = "Strict"
		if config.Domain == "" {
			config.Domain = "example.com" // Change to your domain
		}
	} else {
		// Development
		config.Secure = false
		config.SameSite = "Lax"
		config.Domain = ""
	}

	return config
}

// =====================================================
// OAuth State Cookie - Short lived, for CSRF protection
// =====================================================

// SetOAuthStateCookie sets the OAuth state cookie (15 minutes expiration)
func SetOAuthStateCookie(c fiber.Ctx, state string) {
	cfg := GetCookieConfig()

	c.Cookie(&fiber.Cookie{
		Name:     CookieNameOAuthState,
		Value:    state,
		Path:     "/",
		Domain:   cfg.Domain,
		MaxAge:   15 * 60, // 15 minutes
		Secure:   cfg.Secure,
		HTTPOnly: true,
		SameSite: cfg.SameSite,
	})
}

// GetOAuthStateCookie retrieves the OAuth state cookie
func GetOAuthStateCookie(c fiber.Ctx) string {
	return c.Cookies(CookieNameOAuthState)
}

// ClearOAuthStateCookie removes the OAuth state cookie
func ClearOAuthStateCookie(c fiber.Ctx) {
	cfg := GetCookieConfig()

	c.Cookie(&fiber.Cookie{
		Name:     CookieNameOAuthState,
		Value:    "",
		Path:     "/",
		Domain:   cfg.Domain,
		Expires:  time.Now().Add(-24 * time.Hour), // Expiración inmediata en el pasado
		MaxAge:   -1,                            // Indica al navegador que la borre
		Secure:   cfg.Secure,
		HTTPOnly: true,
		SameSite: cfg.SameSite,
	})
}

// =====================================================
// Auth Token Cookie - Medium lived, for authentication
// =====================================================

// SetAuthTokenCookie sets the authentication token cookie (24 hours by default)
func SetAuthTokenCookie(c fiber.Ctx, token string, duration ...time.Duration) {
	cfg := GetCookieConfig()

	// Default 24 hours if not specified
	maxAge := 24 * 60 * 60
	if len(duration) > 0 {
		maxAge = int(duration[0].Seconds())
	}

	c.Cookie(&fiber.Cookie{
		Name:     CookieNameAuthToken,
		Value:    token,
		Path:     "/",
		Domain:   cfg.Domain,
		MaxAge:   maxAge,
		Secure:   cfg.Secure,
		HTTPOnly: true,
		SameSite: cfg.SameSite,
	})
}

// GetAuthTokenCookie retrieves the authentication token cookie
func GetAuthTokenCookie(c fiber.Ctx) string {
	return c.Cookies(CookieNameAuthToken)
}

// ClearAuthTokenCookie removes the authentication token cookie
func ClearAuthTokenCookie(c fiber.Ctx) {
	cfg := GetCookieConfig()

	c.Cookie(&fiber.Cookie{
		Name:     CookieNameAuthToken,
		Value:    "",
		Path:     "/",
		Domain:   cfg.Domain,
		Expires:  time.Now().Add(-24 * time.Hour),
		MaxAge:   -1,
		Secure:   cfg.Secure,
		HTTPOnly: true,
		SameSite: cfg.SameSite,
	})
}

// =====================================================
// Refresh Token Cookie - Long lived, for token refresh
// =====================================================

// SetRefreshTokenCookie sets the refresh token cookie (7 days by default)
func SetRefreshTokenCookie(c fiber.Ctx, token string, duration ...time.Duration) {
	cfg := GetCookieConfig()

	// Default 7 days if not specified
	maxAge := 7 * 24 * 60 * 60
	if len(duration) > 0 {
		maxAge = int(duration[0].Seconds())
	}

	c.Cookie(&fiber.Cookie{
		Name:     CookieNameRefreshToken,
		Value:    token,
		Path:     "/", // Cambiado de /api/auth/refresh a / para que llegue a /api/v1/auth/refresh
		Domain:   cfg.Domain,
		MaxAge:   maxAge,
		Secure:   cfg.Secure,
		HTTPOnly: true,
		SameSite: cfg.SameSite,
	})
}

// GetRefreshTokenCookie retrieves the refresh token cookie
func GetRefreshTokenCookie(c fiber.Ctx) string {
	return c.Cookies(CookieNameRefreshToken)
}

// ClearRefreshTokenCookie removes the refresh token cookie
func ClearRefreshTokenCookie(c fiber.Ctx) {
	cfg := GetCookieConfig()

	c.Cookie(&fiber.Cookie{
		Name:     CookieNameRefreshToken,
		Value:    "",
		Path:     "/", // Asegurar que coincida con el path de creación
		Domain:   cfg.Domain,
		Expires:  time.Now().Add(-24 * time.Hour),
		MaxAge:   -1,
		Secure:   cfg.Secure,
		HTTPOnly: true,
		SameSite: cfg.SameSite,
	})
}

// =====================================================
// Session Cookie - Generic session management
// =====================================================

// SetSessionCookie sets a session cookie with custom duration
func SetSessionCookie(c fiber.Ctx, sessionID string, duration time.Duration) {
	cfg := GetCookieConfig()

	c.Cookie(&fiber.Cookie{
		Name:     CookieNameSessionToken,
		Value:    sessionID,
		Path:     "/",
		Domain:   cfg.Domain,
		MaxAge:   int(duration.Seconds()),
		Secure:   cfg.Secure,
		HTTPOnly: true,
		SameSite: cfg.SameSite,
	})
}

// GetSessionCookie retrieves the session cookie
func GetSessionCookie(c fiber.Ctx) string {
	return c.Cookies(CookieNameSessionToken)
}

// ClearSessionCookie removes the session cookie
func ClearSessionCookie(c fiber.Ctx) {
	cfg := GetCookieConfig()

	c.Cookie(&fiber.Cookie{
		Name:     CookieNameSessionToken,
		Value:    "",
		Path:     "/",
		Domain:   cfg.Domain,
		Expires:  time.Now().Add(-24 * time.Hour),
		MaxAge:   -1,
		Secure:   cfg.Secure,
		HTTPOnly: true,
		SameSite: cfg.SameSite,
	})
}

// =====================================================
// Bulk operations
// =====================================================

// ClearAllAuthCookies removes all authentication related cookies
func ClearAllAuthCookies(c fiber.Ctx) {
	ClearOAuthStateCookie(c)
	ClearAuthTokenCookie(c)
	ClearRefreshTokenCookie(c)
	ClearSessionCookie(c)
}
