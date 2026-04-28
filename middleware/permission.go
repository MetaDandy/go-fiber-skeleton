package middleware

import (
	"slices"

	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/gofiber/fiber/v3"
)

// RequirePermission verifica si el usuario actual tiene el permiso necesario.
// Este middleware se debe apilar estrictamente después de middleware.Jwt
func RequirePermission(requiredPermission string) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Recuperar los permisos extraídos del JWT (array de strings)
		userPermissions, ok := c.Locals("permissions").([]string)
		if !ok {
			return api_error.Forbidden("No permissions found. UnAuthorized access.")
		}

		// Verificar si el permiso requerido está en la lista
		if slices.Contains(userPermissions, requiredPermission) {
			// Tiene permiso = le pasamos el control al Handler Original :)
			return c.Next()
		}

		// Si terminamos de iterar y no lo encontramos
		return api_error.Forbidden("You don't have the necessary permissions to access this resource.")
	}
}
