package role

type Create struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description"`
	RoleID      *string  `json:"role_id"`
	Permissions []string `json:"permissions" validate:"required"`
}

type Update struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	RoleID      *string  `json:"role_id"`
	Permissions []string `json:"permissions" validate:"required"`
}
