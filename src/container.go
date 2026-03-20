package src

import (
	"github.com/MetaDandy/go-fiber-skeleton/config"
	authentication "github.com/MetaDandy/go-fiber-skeleton/src/core/auth"
	"github.com/MetaDandy/go-fiber-skeleton/src/core/user"
	"github.com/gofiber/fiber/v2"
)

type Container struct {
	Handlers []interface {
		RegisterRoutes(fiber.Router)
	}
}

func SetupContainer() *Container {
	userRepo := user.NewRepo(config.DB)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	authRepo := authentication.NewRepo(config.DB)
	authService := authentication.NewService(authRepo, userRepo)
	authHandler := authentication.NewHandler(authService)

	return &Container{
		Handlers: []interface {
			RegisterRoutes(fiber.Router)
		}{
			userHandler,
			authHandler,
		},
	}
}
