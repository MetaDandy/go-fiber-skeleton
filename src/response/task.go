package response

import (
	"time"

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

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func TaskToDto(m *model.Task) Task {
	var dto Task
	copier.Copy(&dto, m)
	dto.ID = m.ID.String()
	dto.Status = m.Status.String()

	if m.User.ID != uuid.Nil {
		dto.User = UserToDto(&m.User)
	}
	dto.CreatedAt = m.CreatedAt.Format(time.RFC3339)
	dto.UpdatedAt = m.UpdatedAt.Format(time.RFC3339)

	return dto
}

func TaskToListDto(m []model.Task) []Task {
	out := make([]Task, len(m))
	for i := range m {
		out[i] = TaskToDto(&m[i])
	}
	return out
}
