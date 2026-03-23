package src

import (
	"log"

	"github.com/MetaDandy/go-fiber-skeleton/config"
	permission "github.com/MetaDandy/go-fiber-skeleton/src/core/Permission"
	authentication "github.com/MetaDandy/go-fiber-skeleton/src/core/auth"
	"github.com/MetaDandy/go-fiber-skeleton/src/core/role"
	"github.com/MetaDandy/go-fiber-skeleton/src/core/user"
	"github.com/MetaDandy/go-fiber-skeleton/src/service/mail"
	"github.com/gofiber/fiber/v3"
)

type Container struct {
	Handlers []interface {
		RegisterRoutes(fiber.Router)
	}
}

func SetupContainer() *Container {
	// Setup mail service (agnóstico - Mailpit o Resend)
	mailService, err := mail.NewEmailService()
	if err != nil {
		log.Fatalf("Failed to initialize mail service: %v", err)
	}

	userRepo := user.NewRepo(config.DB)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	authRepo := authentication.NewRepo(config.DB)
	authService := authentication.NewService(authRepo, userRepo, mailService)
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
