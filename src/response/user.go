package response

import (
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/jinzhu/copier"
)

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func UserToDto(m *model.User) User {
	var dto User
	copier.Copy(&dto, m)

	return dto
}

func UserToListDto(m []model.User) []User {
	out := make([]User, len(m))
	for i := range m {
		out[i] = UserToDto(&m[i])
	}
	return out
}
