package task

type CreateTaskDto struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
	Status      string `json:"status" validate:"required"`
	UserID      string `json:"user_id" validate:"required,uuid"`
}

type UpdateTaskDto struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	UserID      *string `json:"user_id" validate:"omitempty,uuid"`
}
