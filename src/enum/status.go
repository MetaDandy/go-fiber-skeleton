package enum

import "fmt"

type Status string

const (
	StatusPending    Status = "pendiente"
	StatusDone       Status = "hecho"
	StatusInProgress Status = "en_progreso"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusDone, StatusInProgress:
		return true
	}
	return false
}

func (s Status) String() string {
	return string(s)
}

func ParseStatus(s string) (Status, error) {
	switch s {
	case string(StatusPending):
		return StatusPending, nil
	case string(StatusDone):
		return StatusDone, nil
	case string(StatusInProgress):
		return StatusInProgress, nil
	default:
		return "", fmt.Errorf("invalid status: %s", s)
	}
}
