package response

import (
	"time"

	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/jinzhu/copier"
)

type Role struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	RoleID      string           `json:"role_id"`
	Permissions []RolePermission `json:"permissions"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func RoleToDto(m *model.Role) Role {
	var dto Role
	copier.Copy(&dto, m)
	dto.ID = m.ID.String()
	if m.RoleID != nil {
		dto.RoleID = m.RoleID.String()
	} else {
		dto.RoleID = ""
	}
	if len(m.Role_permissions) > 0 {
		dto.Permissions = RolePermissionToListDto(m.Role_permissions)
	}
	dto.CreatedAt = m.CreatedAt.Format(time.RFC3339)
	dto.UpdatedAt = m.UpdatedAt.Format(time.RFC3339)

	return dto
}

func RoleToListDto(m []model.Role) []Role {
	out := make([]Role, len(m))
	for i := range m {
		out[i] = RoleToDto(&m[i])
	}
	return out
}
