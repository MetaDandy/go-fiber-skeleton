package api_error

// BadRequest retorna error 400 - Solicitud inválida
func BadRequest(message string) *Error {
	return New(400, ErrBadRequest, message)
}

// Unauthorized retorna error 401 - No autorizado
func Unauthorized(message string) *Error {
	return New(401, ErrUnauthorized, message)
}

// Forbidden retorna error 403 - Prohibido
func Forbidden(message string) *Error {
	return New(403, ErrForbidden, message)
}

// NotFound retorna error 404 - No encontrado
func NotFound(message string) *Error {
	return New(404, ErrNotFound, message)
}

// Conflict retorna error 409 - Conflicto (ya existe)
func Conflict(message string) *Error {
	return New(409, ErrConflict, message)
}

// UnprocessableEntity retorna error 422 - Entidad no procesable
func UnprocessableEntity(message string) *Error {
	return New(422, ErrUnprocessableEntity, message)
}

// InternalServerError retorna error 500 - Error interno
func InternalServerError(message string) *Error {
	return New(500, ErrInternalServerError, message)
}

// NotImplemented retorna error 501 - No implementado
func NotImplemented(message string) *Error {
	return New(501, ErrNotImplemented, message)
}

// ServiceUnavailable retorna error 503 - Servicio no disponible
func ServiceUnavailable(message string) *Error {
	return New(503, ErrServiceUnavailable, message)
}

// TooManyRequests retorna error 429
func TooManyRequests(message string) *Error {
	return New(429, ErrTooManyRequests, message)
}
