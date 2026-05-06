package middleware

import (
	"fmt"
	"os"
	"strings"

	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/service/cookie"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// TokenProvider define lo que el middleware necesita para refrescar tokens
type TokenProvider interface {
	RefreshToken(refreshToken, ip, userAgent string) (newToken string, newRefreshToken string, err *api_error.Error)
}

func Jwt(provider TokenProvider) fiber.Handler {
	return func(c fiber.Ctx) error {
		// 1. Intentar obtener el token de la cookie primero, luego del header Authorization
		tokenString := c.Cookies(cookie.CookieNameAuthToken)
		if tokenString == "" {
			authHeader := c.Get("Authorization")
			if authHeader != "" {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// 2. Si no hay token de acceso, intentar refrescar usando el refresh token
		if tokenString == "" {
			return tryRefreshToken(c, provider)
		}

		// 3. Validar el token de acceso
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		// 4. Si el token está expirado o es inválido, intentar refrescar
		if err != nil {
			return tryRefreshToken(c, provider)
		}

		// 5. Extraer claims y setear en Locals
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Locals("user_id", claims["sub"])
			c.Locals("email", claims["email"])
			if role, ok := claims["role"].(string); ok {
				c.Locals("role", role)
			}
			if permissions, ok := claims["permissions"].([]interface{}); ok {
				permsStr := make([]string, len(permissions))
				for i, p := range permissions {
					permsStr[i] = p.(string)
				}
				c.Locals("permissions", permsStr)
			}
			return c.Next()
		}

		return tryRefreshToken(c, provider)
	}
}

func tryRefreshToken(c fiber.Ctx, provider TokenProvider) error {
	refreshToken := c.Cookies(cookie.CookieNameRefreshToken)
	if refreshToken == "" {
		return api_error.Unauthorized("No session found (missing refresh token)")
	}

	ip, userAgent := helper.GetClientDetails(c)

	// Usar la interfaz inyectada para rotar el token
	newToken, newRefreshToken, err := provider.RefreshToken(refreshToken, ip, userAgent)
	if err != nil {
		return err
	}

	// Setear las nuevas cookies
	cookie.SetAuthTokenCookie(c, newToken)
	cookie.SetRefreshTokenCookie(c, newRefreshToken)

	// Volver a validar el NUEVO token para llenar Locals
	token, parseErr := jwt.Parse(newToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if parseErr != nil {
		return api_error.Unauthorized("Failed to parse newly generated token").WithErr(parseErr)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		c.Locals("user_id", claims["sub"])
		c.Locals("email", claims["email"])
		if role, ok := claims["role"].(string); ok {
			c.Locals("role", role)
		}
		if permissions, ok := claims["permissions"].([]interface{}); ok {
			permsStr := make([]string, len(permissions))
			for i, p := range permissions {
				permsStr[i] = p.(string)
			}
			c.Locals("permissions", permsStr)
		}
	}

	return c.Next()
}
