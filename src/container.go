package src

import (
	"github.com/MetaDandy/go-fiber-skeleton/config"
	permission "github.com/MetaDandy/go-fiber-skeleton/src/core/Permission"
	authentication "github.com/MetaDandy/go-fiber-skeleton/src/core/auth"
	"github.com/MetaDandy/go-fiber-skeleton/src/core/user"
	"github.com/gofiber/fiber/v3"
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

	PermissionRepo := permission.NewRepo(config.DB)
	PermissionService := permission.NewService(PermissionRepo)
	PermissionHandler := permission.NewHandler(PermissionService)

	return &Container{
		Handlers: []interface {
			RegisterRoutes(fiber.Router)
		}{
			userHandler,
			authHandler,
			PermissionHandler,
		},
	}
}
