package src

import (
	"log"
	"os"

	"github.com/MetaDandy/go-fiber-skeleton/config"
	"github.com/MetaDandy/go-fiber-skeleton/middleware"
	permission "github.com/MetaDandy/go-fiber-skeleton/src/core/Permission"
	authentication "github.com/MetaDandy/go-fiber-skeleton/src/core/auth"
	"github.com/MetaDandy/go-fiber-skeleton/src/core/role"
	"github.com/MetaDandy/go-fiber-skeleton/src/core/user"
	"github.com/MetaDandy/go-fiber-skeleton/src/core/user_permission"
	"github.com/MetaDandy/go-fiber-skeleton/src/service/mail"
	"github.com/gofiber/fiber/v3"
)

var AuthMiddleware fiber.Handler

type Container struct {
	Handlers []interface {
		RegisterRoutes(fiber.Router)
	}
}

func SetupContainer() *Container {
	mailService, err := mail.NewEmailService()
	if err != nil {
		log.Fatalf("Failed to initialize mail service: %v", err)
	}

	userRepo := user.NewRepo(config.DB)
	userService := user.NewService(userRepo)

	authRepo := authentication.NewRepo(config.DB)

	passwordSvc := authentication.NewPasswordService(authRepo, userRepo, mailService, os.Getenv("APP_URL"))
	oauthSvc := authentication.NewOAuthService(authRepo, userRepo, os.Getenv("URI_REDIRECT"))
	emailSvc := authentication.NewEmailService(authRepo, userRepo, mailService)
	sessionSvc := authentication.NewSessionService(authRepo, userRepo)

	AuthMiddleware = middleware.Jwt(sessionSvc)

	authHandler := authentication.NewHandler(passwordSvc, oauthSvc, emailSvc, sessionSvc, AuthMiddleware)

	userHandler := user.NewHandler(userService, AuthMiddleware)

	permissionRepo := permission.NewRepo(config.DB)
	permissionService := permission.NewService(permissionRepo)
	permissionHandler := permission.NewHandler(permissionService, AuthMiddleware)

	roleRepo := role.NewRepo(config.DB)
	roleService := role.NewService(roleRepo, role.PermissionChecker(permissionService))
	roleHandler := role.NewHandler(roleService, AuthMiddleware)

	userPermissionRepo := user_permission.NewRepo(config.DB)
	userPermissionService := user_permission.NewService(userPermissionRepo, user_permission.PermissionChecker(permissionService))
	userPermissionHandler := user_permission.NewHandler(userPermissionService, AuthMiddleware)

	return &Container{
		Handlers: []interface {
			RegisterRoutes(fiber.Router)
		}{
			userHandler,
			authHandler,
			permissionHandler,
			roleHandler,
			userPermissionHandler,
		},
	}
}
