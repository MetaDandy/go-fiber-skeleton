package api

import (
	"github.com/MetaDandy/go-fiber-skeleton/src"
	"github.com/gofiber/fiber/v2"
)

func SetupApi(app *fiber.App, c *src.Container) {
	v1 := app.Group("/api/v1")

	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Aloha")
	})

	handlers := []func(fiber.Router){
		c.TaskHandler.RegisterRoutes,
		c.UserHandler.RegisterRoutes,
	}

	for _, register := range handlers {
		register(v1)
	}
}
