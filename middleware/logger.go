package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		fmt.Printf("📢 Ruta accedida: %s %s\n", c.Method(), c.OriginalURL())
		return c.Next()
	}
}
