package src

import (
	"github.com/MetaDandy/go-fiber-skeleton/config"
	"github.com/MetaDandy/go-fiber-skeleton/src/modules/task"
	"github.com/MetaDandy/go-fiber-skeleton/src/modules/user"
)

type Container struct {
	UserHandler user.Handler
	TaskHandler task.Handler
}

func SetupContainer() *Container {
	userRepo := user.NewRepo(config.DB)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	taskRepo := task.NewRepo(config.DB)
	taskService := task.NewService(taskRepo)
	taskHandler := task.NewHandler(taskService)

	return &Container{
		UserHandler: userHandler,
		TaskHandler: taskHandler,
	}
}
