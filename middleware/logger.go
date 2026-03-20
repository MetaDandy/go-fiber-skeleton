package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
)

func Logger() fiber.Handler {
	return func(c fiber.Ctx) error {
		fmt.Printf("📢 Ruta accedida: %s %s\n", c.Method(), c.OriginalURL())
		fmt.Printf("Body: %s", c.Body())
		return c.Next()
	}
}
