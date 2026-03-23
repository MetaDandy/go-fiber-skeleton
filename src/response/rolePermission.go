package response

import (
	"time"

	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/jinzhu/copier"
)

type RolePermission struct {
	ID string `json:"id"`

	RoleID       string `json:"role_id"`
	PermissionID string `json:"permission_id"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func RolePermissionToDto(m *model.RolePermission) RolePermission {
	var dto RolePermission
	copier.Copy(&dto, m)
	dto.ID = m.ID.String()

	dto.RoleID = m.RoleID.String()
	dto.PermissionID = m.PermissionID

	dto.CreatedAt = m.CreatedAt.Format(time.RFC3339)
	dto.UpdatedAt = m.UpdatedAt.Format(time.RFC3339)

	return dto
}

func RolePermissionToListDto(m []model.RolePermission) []RolePermission {
	out := make([]RolePermission, len(m))
	for i := range m {
		out[i] = RolePermissionToDto(&m[i])
	}
	return out
}
