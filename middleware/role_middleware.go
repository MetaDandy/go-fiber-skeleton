package middleware

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func RequireRoleMiddleware(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("role")
		log.Println("Required role:", role)
		log.Println("User role:", userRole)
		if userRole != role {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: insufficient permissions",
			})
		}
		return c.Next()
	}
}
