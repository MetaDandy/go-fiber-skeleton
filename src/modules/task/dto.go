package task

type Create struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
	UserID      string `json:"user_id" validate:"required,uuid"`
}

type Update struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
}
