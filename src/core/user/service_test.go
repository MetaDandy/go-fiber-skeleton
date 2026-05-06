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

func TestCreate_ReturnsNil_WhenSuccess(t *testing.T) {
	mockRepo := new(MockRepo)
	service := NewService(mockRepo)
	
	mockRepo.On("Create", mock.Anything).Return(nil)
	
	apiErr := service.Create(Create{Name: "Test User", Email: "test@example.com"})
	
	assert.Nil(t, apiErr)
	mockRepo.AssertExpectations(t)
}

func TestFindByID_ReturnsUser_WhenExists(t *testing.T) {
	mockRepo := new(MockRepo)
	service := NewService(mockRepo)
	
	testUser := model.User{
		ID:    uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	}
	
	mockRepo.On("FindByID", mock.Anything).Return(testUser, nil)
	
	user, apiErr := service.FindByID(testUser.ID.String())
	
	assert.Nil(t, apiErr)
	assert.NotNil(t, user)
	assert.Equal(t, testUser.Name, user.Name)
	assert.Equal(t, testUser.Email, user.Email)
	mockRepo.AssertExpectations(t)
}

func TestFindAll_ReturnsPaginatedUsers(t *testing.T) {
	mockRepo := new(MockRepo)
	service := NewService(mockRepo)
	
	testUsers := []model.User{
		{ID: uuid.New(), Name: "User 1", Email: "user1@example.com"},
		{ID: uuid.New(), Name: "User 2", Email: "user2@example.com"},
	}
	
	mockRepo.On("FindAll", mock.Anything).Return(testUsers, int64(2), nil)
	
	opts := &helper.FindAllOptions{Limit: 10, Offset: 0}
	result, apiErr := service.FindAll(opts)
	
	assert.Nil(t, apiErr)
	assert.NotNil(t, result)
	assert.Len(t, result.Data, 2)
	assert.Equal(t, int64(2), result.Total)
	assert.Equal(t, uint(10), result.Limit)
	assert.Equal(t, uint(0), result.Offset)
	mockRepo.AssertExpectations(t)
}

func TestUpdate_UpdatesUserFields(t *testing.T) {
	mockRepo := new(MockRepo)
	service := NewService(mockRepo)
	
	testUserID := uuid.New()
	testUser := model.User{
		ID:    testUserID,
		Name:  "Original Name",
		Email: "original@example.com",
	}
	
	newName := "Updated Name"
	newEmail := "updated@example.com"
	
	mockRepo.On("FindByID", testUserID.String()).Return(testUser, nil)
	mockRepo.On("Update", mock.Anything).Return(nil)
	
	apiErr := service.Update(testUserID.String(), Update{Name: &newName, Email: &newEmail})
	
	assert.Nil(t, apiErr)
	mockRepo.AssertCalled(t, "Update", mock.Anything)
}

func TestUpdate_Returns404_WhenUserNotFound(t *testing.T) {
	mockRepo := new(MockRepo)
	service := NewService(mockRepo)
	
	nonExistentID := uuid.New().String()
	
	mockRepo.On("FindByID", nonExistentID).Return(model.User{}, errors.New("not found"))
	
	apiErr := service.Update(nonExistentID, Update{Name: pointerTo("New Name")})
	
	assert.NotNil(t, apiErr)
	assert.Equal(t, 404, apiErr.Status)
	assert.Equal(t, api_error.ErrNotFound, apiErr.Code)
	assert.Equal(t, "User not found", apiErr.Message)
}

// pointerTo is a helper to create string pointers
func pointerTo(s string) *string {
	return &s
}

func TestDelete_ReturnsNil_WhenSuccess(t *testing.T) {
	mockRepo := new(MockRepo)
	service := NewService(mockRepo)
	
	testUserID := uuid.New().String()
	
	mockRepo.On("Delete", testUserID).Return(nil)
	
	apiErr := service.Delete(testUserID)
	
	assert.Nil(t, apiErr)
	mockRepo.AssertExpectations(t)
}

func TestDelete_ReturnsError_WhenRepoFails(t *testing.T) {
	mockRepo := new(MockRepo)
	service := NewService(mockRepo)
	
	testUserID := uuid.New().String()
	
	mockRepo.On("Delete", testUserID).Return(errors.New("db error"))
	
	apiErr := service.Delete(testUserID)
	
	assert.NotNil(t, apiErr)
	assert.Equal(t, 500, apiErr.Status)
	assert.Equal(t, api_error.ErrInternalServerError, apiErr.Code)
	mockRepo.AssertExpectations(t)
}
