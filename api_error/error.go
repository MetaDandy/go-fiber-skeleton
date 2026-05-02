package api_error

import "time"

// Error es la representación interna de un error de API
type Error struct {
	Status    int       `json:"status"`
	Code      ErrorCode `json:"code"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Err       error     `json:"-"`
}

// Implementar interface error
func (e *Error) Error() string {
	return e.Message
}

// New crea un nuevo error de API estandarizado
func New(status int, code ErrorCode, message string) *Error {
	return &Error{
		Status:    status,
		Code:      code,
		Message:   message,
		Timestamp: time.Now().UTC(),
	}
}

// WithErr agrega el error interno para logging
func (e *Error) WithErr(err error) *Error {
	e.Err = err
	return e
}
