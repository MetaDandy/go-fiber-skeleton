package api_error

import "github.com/gofiber/fiber/v3"

// BadRequest retorna error 400 - Solicitud inválida
func BadRequest(message string) *Error {
	return NewError(
		ErrBadRequest,
		message,
		fiber.StatusBadRequest,
	)
}

// Unauthorized retorna error 401 - No autorizado
func Unauthorized(message string) *Error {
	return NewError(
		ErrUnauthorized,
		message,
		fiber.StatusUnauthorized,
	)
}

// Forbidden retorna error 403 - Prohibido
func Forbidden(message string) *Error {
	return NewError(
		ErrForbidden,
		message,
		fiber.StatusForbidden,
	)
}

// NotFound retorna error 404 - No encontrado
func NotFound(message string) *Error {
	return NewError(
		ErrNotFound,
		message,
		fiber.StatusNotFound,
	)
}

// Conflict retorna error 409 - Conflicto (ya existe)
func Conflict(message string) *Error {
	return NewError(
		ErrConflict,
		message,
		fiber.StatusConflict,
	)
}

// UnprocessableEntity retorna error 422 - Entidad no procesable
func UnprocessableEntity(message string) *Error {
	return NewError(
		ErrUnprocessableEntity,
		message,
		fiber.StatusUnprocessableEntity,
	)
}

// TooManyRequests retorna error 429 - Demasiadas solicitudes
func TooManyRequests(message string) *Error {
	return NewError(
		ErrTooManyRequests,
		message,
		fiber.StatusTooManyRequests,
	)
}

// InternalServerError retorna error 500 - Error interno
func InternalServerError(message string) *Error {
	return NewError(
		ErrInternalServerError,
		message,
		fiber.StatusInternalServerError,
	)
}

// NotImplemented retorna error 501 - No implementado
func NotImplemented(message string) *Error {
	return NewError(
		ErrNotImplemented,
		message,
		fiber.StatusNotImplemented,
	)
}

// ServiceUnavailable retorna error 503 - Servicio no disponible
func ServiceUnavailable(message string) *Error {
	return NewError(
		ErrServiceUnavailable,
		message,
		fiber.StatusServiceUnavailable,
	)
}
