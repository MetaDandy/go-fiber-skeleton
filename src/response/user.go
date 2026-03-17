package response

import (
	"time"

	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/jinzhu/copier"
)

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func UserToDto(m *model.User) User {
	var dto User
	copier.Copy(&dto, m)
	dto.ID = m.ID.String()
	dto.CreatedAt = m.CreatedAt.Format(time.RFC3339)
	dto.UpdatedAt = m.UpdatedAt.Format(time.RFC3339)

	return dto
}

func UserToListDto(m []model.User) []User {
	out := make([]User, len(m))
	for i := range m {
		out[i] = UserToDto(&m[i])
	}
	return out
}
