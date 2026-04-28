package response

import (
	"time"

	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/jinzhu/copier"
)

type UserPermission struct {
	ID           string `json:"id"`
	PermissionID string `json:"permission_id"`
	UserID       string `json:"user_id"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

func UserPermissionToDto(m *model.UserPermission) UserPermission {
	var dto UserPermission
	copier.Copy(&dto, m)
	dto.ID = m.ID.String()
	dto.UserID = m.UserID.String()
	dto.CreatedAt = m.CreatedAt.Format(time.RFC3339)
	dto.UpdatedAt = m.UpdatedAt.Format(time.RFC3339)
	return dto
}

func UserPermissionToListDto(m []model.UserPermission) []UserPermission {
	out := make([]UserPermission, len(m))
	for i := range m {
		out[i] = UserPermissionToDto(&m[i])
	}
	return out
}
