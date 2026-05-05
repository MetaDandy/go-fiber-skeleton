package user

import (
	"errors"
	"testing"

	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepo is a mock of the Repo interface
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) Create(u model.User) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *MockRepo) FindByID(id string) (model.User, error) {
	args := m.Called(id)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockRepo) FindByEmail(email string) (model.User, error) {
	args := m.Called(email)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockRepo) FindAll(opts *helper.FindAllOptions) ([]model.User, int64, error) {
	args := m.Called(opts)
	return args.Get(0).([]model.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepo) Update(u model.User) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *MockRepo) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRepo) Exists(id string) error { return nil }
func (m *MockRepo) ExistsByEmail(email string) error { return nil }
func (m *MockRepo) UpdatePassword(id string, pass string) error { return nil }

func TestServiceErrors(t *testing.T) {
	t.Run("FindByID returns typed 404 error when user not found", func(t *testing.T) {
		mockRepo := new(MockRepo)
		service := NewService(mockRepo)
		
		nonExistentID := uuid.New().String()
		mockRepo.On("FindByID", nonExistentID).Return(model.User{}, errors.New("not found"))
		
		user, apiErr := service.FindByID(nonExistentID)
		
		assert.Nil(t, user)
		assert.NotNil(t, apiErr)
		assert.Equal(t, 404, apiErr.Status)
		assert.Equal(t, api_error.ErrNotFound, apiErr.Code)
		assert.Equal(t, "User not found", apiErr.Message)
	})

	t.Run("Create returns typed 500 error on repo failure", func(t *testing.T) {
		mockRepo := new(MockRepo)
		service := NewService(mockRepo)
		
		mockRepo.On("Create", mock.Anything).Return(errors.New("db error"))
		
		apiErr := service.Create(Create{Name: "Test", Email: "test@example.com"})
		
		assert.NotNil(t, apiErr)
		assert.Equal(t, 500, apiErr.Status)
		assert.Equal(t, api_error.ErrInternalServerError, apiErr.Code)
	})
}
