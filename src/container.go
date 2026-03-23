package src

import (
	"github.com/MetaDandy/go-fiber-skeleton/config"
	permission "github.com/MetaDandy/go-fiber-skeleton/src/core/Permission"
	authentication "github.com/MetaDandy/go-fiber-skeleton/src/core/auth"
	"github.com/MetaDandy/go-fiber-skeleton/src/core/role"
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

	permissionRepo := permission.NewRepo(config.DB)
	permissionService := permission.NewService(permissionRepo)
	permissionHandler := permission.NewHandler(permissionService)

	roleRepo := role.NewRepo(config.DB)
	roleService := role.NewService(roleRepo, role.PermissionChecker(permissionService))
	roleHandler := role.NewHandler(roleService)

	return &Container{
		Handlers: []interface {
			RegisterRoutes(fiber.Router)
		}{
			userHandler,
			authHandler,
			permissionHandler,
			roleHandler,
		},
	}
}
