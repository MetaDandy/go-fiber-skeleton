package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
)

func Logger() fiber.Handler {
	return func(c fiber.Ctx) error {
		fmt.Printf("📢 Ruta accedida: %s %s\n", c.Method(), c.OriginalURL())

		fmt.Printf("Cookies: %s\n", c.Get("Cookie"))
		fmt.Printf("Body: %s\n", c.Body())
		return c.Next()
	}
}
