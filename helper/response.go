package helper

type PaginatedResponse[T any] struct {
	Data   []T   `json:"data"`
	Total  int64 `json:"total"`
	Limit  uint  `json:"limit"`
	Offset uint  `json:"offset"`
	Pages  uint  `json:"pages"`
}

// Todo: Revisar el tema del any, ver si poner T o no
type Response struct {
	Data    any    `json:"data"`
	Token   string `json:"token,omitempty"`
	Message string `json:"message"`
}
