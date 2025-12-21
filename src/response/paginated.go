package response

type Paginated[T any] struct {
	Data   []T   `json:"data"`
	Total  int64 `json:"total"`
	Limit  uint  `json:"limit"`
	Offset uint  `json:"offset"`
	Pages  uint  `json:"pages"`
}
