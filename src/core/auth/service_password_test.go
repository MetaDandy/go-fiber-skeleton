package authentication

import (
	"errors"
	"testing"

	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPasswordURepo struct {
	mock.Mock
}

func (m *MockPasswordURepo) FindByEmail(email string) (model.User, error) {
	args := m.Called(email)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockPasswordURepo) ExistsByEmail(email string) error {
	args := m.Called(email)
	return args.Error(0)
}

func (m *MockPasswordURepo) FindByID(id string) (model.User, error) {
	args := m.Called(id)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockPasswordURepo) UpdatePassword(userID string, passwordHash string) error {
	args := m.Called(userID, passwordHash)
	return args.Error(0)
}

// Ensure the new signature is tested
func TestUserAuthProviders_ReturnsApiError(t *testing.T) {
	uRepo := new(MockPasswordURepo)
	
	// mock finding by email fails
	uRepo.On("FindByEmail", "test@test.com").Return(model.User{}, errors.New("db error"))

	// We only care about the service returning *api_error.Error
	service := NewPasswordService(nil, uRepo, nil, "")

	_, err := service.UserAuthProviders("test@test.com")
	
	assert.Error(t, err)
	var apiErr *api_error.Error
	assert.True(t, errors.As(err, &apiErr))
	assert.Equal(t, 500, apiErr.Status)
}
