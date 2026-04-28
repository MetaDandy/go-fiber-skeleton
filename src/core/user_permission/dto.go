package user_permission

type UpdateDetails struct {
	Add    []string `json:"add"`
	Remove []string `json:"remove"`
}
