package testutil

import (
	"testing"

	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestMockUserRepoFactory tests that the factory creates a properly mocked UserRepo
func TestMockUserRepoFactory(t *testing.T) {
	mockRepo := NewMockUserRepo()

	// Verify mock is initialized
	assert.NotNil(t, mockRepo, "mockRepo should not be nil")

	// Test SetupFindByID
	testUser := model.User{
		ID:    uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	}
	mockRepo.On("FindByID", mock.Anything).Return(testUser, nil)

	user, err := mockRepo.FindByID("any-id")
	assert.NoError(t, err)
	assert.Equal(t, testUser, user)
	mockRepo.AssertExpectations(t)
}

// TestMockUserRepo_SetupFindByEmail tests email lookup mock
func TestMockUserRepo_SetupFindByEmail(t *testing.T) {
	mockRepo := NewMockUserRepo()

	email := "test@example.com"
	testUser := model.User{ID: uuid.New(), Email: email}
	mockRepo.On("FindByEmail", email).Return(testUser, nil)

	user, err := mockRepo.FindByEmail(email)
	assert.NoError(t, err)
	assert.Equal(t, email, user.Email)
}

// TestMockUserRepo_SetupCreate tests create mock
func TestMockUserRepo_SetupCreate(t *testing.T) {
	mockRepo := NewMockUserRepo()

	newUser := model.User{
		ID:    uuid.New(),
		Name:  "New User",
		Email: "new@example.com",
	}
	mockRepo.On("Create", mock.Anything).Return(nil)

	err := mockRepo.Create(newUser)
	assert.NoError(t, err)
}

// TestMockUserRepo_SetupFindAll tests FindAll mock with options
func TestMockUserRepo_SetupFindAll(t *testing.T) {
	mockRepo := NewMockUserRepo()

	users := []model.User{
		{ID: uuid.New(), Name: "User 1"},
		{ID: uuid.New(), Name: "User 2"},
	}
	opts := &helper.FindAllOptions{Limit: 10, Offset: 0}
	mockRepo.On("FindAll", opts).Return(users, int64(2), nil)

	result, count, err := mockRepo.FindAll(opts)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, int64(2), count)
}

// TestMockRoleRepoFactory tests Role mock factory
func TestMockRoleRepoFactory(t *testing.T) {
	mockRepo := NewMockRoleRepo()

	testRole := model.Role{
		ID:   uuid.New(),
		Name: "admin",
	}
	mockRepo.On("FindByID", mock.Anything).Return(testRole, nil)

	role, err := mockRepo.FindByID(uuid.New())
	assert.NoError(t, err)
	assert.Equal(t, "admin", role.Name)
}

// TestMockPermissionRepoFactory tests Permission mock factory
func TestMockPermissionRepoFactory(t *testing.T) {
	mockRepo := NewMockPermissionRepo()

	testPerm := model.Permission{
		ID:          "read",
		Name:        "Read",
		Description: "Read permission",
	}
	mockRepo.On("FindByID", mock.Anything).Return(testPerm, nil)

	perm, err := mockRepo.FindByID("any-id")
	assert.NoError(t, err)
	assert.Equal(t, "read", perm.ID)
}

// TestMockUserPermissionRepoFactory tests UserPermission mock factory
func TestMockUserPermissionRepoFactory(t *testing.T) {
	mockRepo := NewMockUserPermissionRepo()

	assert.NotNil(t, mockRepo, "mockRepo should not be nil")

	// Test that we can set up expectations
	mockRepo.On("BeginTx").Return(nil)

	tx := mockRepo.BeginTx()
	assert.Nil(t, tx)
}