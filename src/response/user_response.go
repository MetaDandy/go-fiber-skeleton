package response

import (
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/jinzhu/copier"
)

type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func UserToDto(m *model.User) UserResponse {
	var dto UserResponse
	copier.Copy(&dto, m)

	return dto
}

func UserToListDto(m []model.User) []UserResponse {
	out := make([]UserResponse, len(m))
	for i := range m {
		out[i] = UserToDto(&m[i])
	}
	return out
}
