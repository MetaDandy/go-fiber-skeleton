package main

import (
	"os"

	"github.com/MetaDandy/go-fiber-skeleton/cmd/api"
	"github.com/MetaDandy/go-fiber-skeleton/config"
	"github.com/MetaDandy/go-fiber-skeleton/middleware"
	"github.com/MetaDandy/go-fiber-skeleton/src"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	config.Load()

	app := fiber.New()
	app.Use(middleware.Logger())

	app.Use(cors.New(cors.Config{
		AllowOrigins: os.Getenv("ALLOW_ORIGINS"),
		AllowMethods: "GET,POST,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	c := src.SetupContainer()
	api.SetupApi(app, c)

	app.Listen("0.0.0.0:" + config.Port)
}
