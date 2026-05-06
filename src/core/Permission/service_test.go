package permission

import (
	"errors"
	"testing"
	"time"

	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/core/testutil"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFindByID_ReturnsPermission_WhenExists(t *testing.T) {
	// GIVEN
	mockRepo := testutil.NewMockPermissionRepo()
	service := NewService(mockRepo)

	permissionID := uuid.New().String()
	expectedPermission := model.Permission{
		ID:          permissionID,
		Name:        "test_permission",
		Description: "Test description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.On("FindByID", permissionID).Return(expectedPermission, nil)

	// WHEN
	result, apiErr := service.FindByID(permissionID)

	// THEN
	assert.Nil(t, apiErr)
	assert.NotNil(t, result)
	assert.Equal(t, permissionID, result.ID)
	assert.Equal(t, "test_permission", result.Name)
	assert.Equal(t, "Test description", result.Description)
	mockRepo.AssertExpectations(t)
}

func TestFindByID_Returns404_WhenNotFound(t *testing.T) {
	// GIVEN
	mockRepo := testutil.NewMockPermissionRepo()
	service := NewService(mockRepo)

	nonExistentID := uuid.New().String()
	mockRepo.On("FindByID", nonExistentID).Return(model.Permission{}, errors.New("record not found"))

	// WHEN
	result, apiErr := service.FindByID(nonExistentID)

	// THEN
	assert.Nil(t, result)
	assert.NotNil(t, apiErr)
	assert.Equal(t, 404, apiErr.Status)
	assert.Equal(t, api_error.ErrNotFound, apiErr.Code)
	assert.Equal(t, "Permission not found", apiErr.Message)
	mockRepo.AssertExpectations(t)
}

func TestFindAll_ReturnsPaginatedPermissions(t *testing.T) {
	// GIVEN
	mockRepo := testutil.NewMockPermissionRepo()
	service := NewService(mockRepo)

	permissions := []model.Permission{
		{
			ID:          uuid.New().String(),
			Name:        "perm1",
			Description: "Permission 1",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New().String(),
			Name:        "perm2",
			Description: "Permission 2",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	opts := &helper.FindAllOptions{Limit: 10, Offset: 0}
	mockRepo.On("FindAll", opts).Return(permissions, int64(2), nil)

	// WHEN
	result, apiErr := service.FindAll(opts)

	// THEN
	assert.Nil(t, apiErr)
	assert.NotNil(t, result)
	assert.Equal(t, int64(2), result.Total)
	assert.Equal(t, uint(10), result.Limit)
	assert.Equal(t, uint(0), result.Offset)
	assert.Equal(t, uint(1), result.Pages)
	assert.Len(t, result.Data, 2)
	mockRepo.AssertExpectations(t)
}

func TestAllExists_ReturnsNil_WhenAllExist(t *testing.T) {
	// GIVEN
	mockRepo := testutil.NewMockPermissionRepo()
	service := NewService(mockRepo)

	permissionIDs := []string{uuid.New().String(), uuid.New().String()}
	mockRepo.On("AllExists", permissionIDs).Return(nil)

	// WHEN
	apiErr := service.AllExists(permissionIDs)

	// THEN
	assert.Nil(t, apiErr)
	mockRepo.AssertExpectations(t)
}

func TestAllExists_ReturnsError_WhenSomeDontExist(t *testing.T) {
	// GIVEN
	mockRepo := testutil.NewMockPermissionRepo()
	service := NewService(mockRepo)

	permissionIDs := []string{uuid.New().String(), uuid.New().String()}
	mockRepo.On("AllExists", permissionIDs).Return(api_error.InternalServerError("permission not found"))

	// WHEN
	apiErr := service.AllExists(permissionIDs)

	// THEN
	assert.NotNil(t, apiErr)
	assert.Equal(t, 500, apiErr.Status)
	mockRepo.AssertExpectations(t)
}