package response

import (
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
)

type Task struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	User        User   `json:"user"`
}

func TaskToDto(m *model.Task) Task {
	var dto Task
	copier.Copy(&dto, m)

	if m.User.ID != uuid.Nil {
		dto.User = UserToDto(&m.User)
	}

	return dto
}

func TaskToListDto(m []model.Task) []Task {
	out := make([]Task, len(m))
	for i := range m {
		out[i] = TaskToDto(&m[i])
	}
	return out
}
