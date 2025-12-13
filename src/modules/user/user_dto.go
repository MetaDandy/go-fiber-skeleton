package user

type CreateUserDto struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type UpdateUserDto struct {
	Name  *string `json:"name"`
	Email *string `json:"email" validate:"omitempty,email"`
}
