package api_error

import "time"

// Error es la representación interna de un error de API
type Error struct {
	Code      ErrorCode `json:"-"`
	Message   string    `json:"-"`
	Status    int       `json:"-"`
	Err       error     `json:"-"`
	Timestamp time.Time `json:"-"`
}

// ErrorResponse es lo que se retorna al cliente
type ErrorResponse struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// Implementar interface error
func (e *Error) Error() string {
	return e.Message
}

// NewError crea un nuevo error de API
func NewError(code ErrorCode, message string, status int) *Error {
	return &Error{
		Code:      code,
		Message:   message,
		Status:    status,
		Timestamp: time.Now().UTC(),
	}
}

// WithErr agrega el error interno para logging
func (e *Error) WithErr(err error) *Error {
	e.Err = err
	return e
}

// ToResponse convierte a formato de respuesta
func (e *Error) ToResponse() ErrorResponse {
	return ErrorResponse{
		Code:      e.Code.String(),
		Message:   e.Message,
		Timestamp: e.Timestamp,
	}
}
