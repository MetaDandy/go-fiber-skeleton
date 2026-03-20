package api_error

// ErrorCode representa un código de error de la API
type ErrorCode string

const (
	// 4xx Client Errors
	ErrBadRequest          ErrorCode = "BAD_REQUEST"
	ErrUnauthorized        ErrorCode = "UNAUTHORIZED"
	ErrForbidden           ErrorCode = "FORBIDDEN"
	ErrNotFound            ErrorCode = "NOT_FOUND"
	ErrConflict            ErrorCode = "CONFLICT"
	ErrUnprocessableEntity ErrorCode = "UNPROCESSABLE_ENTITY"
	ErrTooManyRequests     ErrorCode = "TOO_MANY_REQUESTS"

	// 5xx Server Errors
	ErrInternalServerError ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrNotImplemented      ErrorCode = "NOT_IMPLEMENTED"
	ErrServiceUnavailable  ErrorCode = "SERVICE_UNAVAILABLE"
)

// String convierte el ErrorCode a string
func (ec ErrorCode) String() string {
	return string(ec)
}
