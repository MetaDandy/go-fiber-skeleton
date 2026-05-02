package api_error

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	err := New(404, ErrNotFound, "Resource not found")
	
	assert.Equal(t, 404, err.Status)
	assert.Equal(t, ErrNotFound, err.Code)
	assert.Equal(t, "Resource not found", err.Message)
	assert.NotNil(t, err.Timestamp)
}

func TestHelpers(t *testing.T) {
	badRequest := BadRequest("Bad input")
	assert.Equal(t, 400, badRequest.Status)
	assert.Equal(t, ErrBadRequest, badRequest.Code)
	assert.Equal(t, "Bad input", badRequest.Message)

	notFound := NotFound("Not found")
	assert.Equal(t, 404, notFound.Status)
	assert.Equal(t, ErrNotFound, notFound.Code)

	internal := InternalServerError("Server error")
	assert.Equal(t, 500, internal.Status)
	assert.Equal(t, ErrInternalServerError, internal.Code)
}

func TestErrorInterface(t *testing.T) {
	var err error = NotFound("Missing")
	assert.Equal(t, "Missing", err.Error())
}
