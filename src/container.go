package src

import (
	"github.com/MetaDandy/go-fiber-skeleton/config"
	"github.com/MetaDandy/go-fiber-skeleton/src/modules/task"
	"github.com/MetaDandy/go-fiber-skeleton/src/modules/user"
)

type Container struct {
	UserHandler user.UserHandler
	TaskHandler task.TaskHandler
}

func SetupContainer() *Container {
	userRepo := user.NewRepo(config.DB)
	userService := user.NewService(userRepo)
	userHandler := user.NewUserHandler(userService)

	taskRepo := task.NewRepo(config.DB)
	taskService := task.NewService(taskRepo)
	taskHandler := task.NewTaskHandler(taskService)

	return &Container{
		UserHandler: userHandler,
		TaskHandler: taskHandler,
	}
}
