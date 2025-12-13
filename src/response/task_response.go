package response

import (
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
)

type TaskResponse struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Status      string       `json:"status"`
	User        UserResponse `json:"user"`
}

func TaskToDto(m *model.Task) TaskResponse {
	var dto TaskResponse
	copier.Copy(&dto, m)

	if m.User.ID != uuid.Nil {
		dto.User = UserToDto(&m.User)
	}

	return dto
}

func TaskToListDto(m []model.Task) []TaskResponse {
	out := make([]TaskResponse, len(m))
	for i := range m {
		out[i] = TaskToDto(&m[i])
	}
	return out
}
