package src

import (
	"github.com/MetaDandy/go-fiber-skeleton/config"
	"github.com/MetaDandy/go-fiber-skeleton/src/modules/task"
	"github.com/MetaDandy/go-fiber-skeleton/src/modules/user"
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

	taskRepo := task.NewRepo(config.DB)
	taskService := task.NewService(taskRepo, userRepo)
	taskHandler := task.NewHandler(taskService)

	return &Container{
		Handlers: []interface {
			RegisterRoutes(fiber.Router)
		}{
			userHandler,
			taskHandler,
		},
	}
}
