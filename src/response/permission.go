package response

import (
	"time"

	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/jinzhu/copier"
)

type Permission struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Name        string `json:"name"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func PermissionToDto(m *model.Permission) Permission {
	var dto Permission
	copier.Copy(&dto, m)
	dto.ID = m.ID
	dto.CreatedAt = m.CreatedAt.Format(time.RFC3339)
	dto.UpdatedAt = m.UpdatedAt.Format(time.RFC3339)

	return dto
}

func PermissionToListDto(m []model.Permission) []Permission {
	out := make([]Permission, len(m))
	for i := range m {
		out[i] = PermissionToDto(&m[i])
	}
	return out
}
