package role

type Create struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description"`
	RoleID      *string  `json:"role_id"`
	Permissions []string `json:"permissions" validate:"required"`
}

type UpdateHeader struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	RoleID      *string `json:"role_id"`
	StrictMode  bool    `json:"strict_mode"`
}

type UpdateDetails struct {
	Add        []string `json:"add"`
	Remove     []string `json:"remove"`
	StrictMode bool     `json:"strict_mode"`
}
