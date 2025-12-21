package user

type Create struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type Update struct {
	Name  *string `json:"name"`
	Email *string `json:"email" validate:"omitempty,email"`
}
