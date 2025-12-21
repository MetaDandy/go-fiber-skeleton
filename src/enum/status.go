package enum

type Status string

const (
	StatusPending  Status = "pending"
	StatusActive   Status = "active"
	StatusApproved Status = "approved"
)
