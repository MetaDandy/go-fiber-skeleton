package helper

import "github.com/gofiber/fiber/v3"

func GetClientDetails(c fiber.Ctx) (string, string) {
	ip := c.IP()
	if ip == "" {
		ip = c.Get("X-Forwarded-For")
	}
	userAgent := c.Get("User-Agent")
	return ip, userAgent
}
