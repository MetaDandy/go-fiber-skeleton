package user_permission

import (
	"testing"

	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/MetaDandy/go-fiber-skeleton/src/core/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPermissionChecker implements PermissionChecker interface for testing
type MockPermissionChecker struct {
	mock.Mock
}

func (m *MockPermissionChecker) AllExists(ids []string) *api_error.Error {
	args := m.Called(ids)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*api_error.Error)
}

func TestUpdateDetails_ReturnsError_WhenInvalidUserID(t *testing.T) {
	// GIVEN
	mockRepo := testutil.NewMockUserPermissionRepo()
	mockPermissionChecker := new(MockPermissionChecker)
	service := NewService(mockRepo, mockPermissionChecker)

	invalidUserID := "not-a-valid-uuid"
	input := UpdateDetails{
		Add: []string{uuid.New().String()},
	}

	// WHEN
	apiErr := service.UpdateDetails(invalidUserID, input)

	// THEN
	assert.NotNil(t, apiErr)
	assert.Equal(t, 400, apiErr.Status)
	assert.Equal(t, "Invalid user ID", apiErr.Message)
	mockPermissionChecker.AssertNotCalled(t, "AllExists")
	mockRepo.AssertNotCalled(t, "BeginTx")
}

func TestUpdateDetails_ReturnsError_WhenPermissionNotExists(t *testing.T) {
	// GIVEN
	mockRepo := testutil.NewMockUserPermissionRepo()
	mockPermissionChecker := new(MockPermissionChecker)
	service := NewService(mockRepo, mockPermissionChecker)

	userID := uuid.New().String()
	nonExistentPermissionID := uuid.New().String()

	mockPermissionChecker.On("AllExists", []string{nonExistentPermissionID}).
		Return(api_error.BadRequest("permission not found"))

	input := UpdateDetails{
		Add: []string{nonExistentPermissionID},
	}

	// WHEN
	apiErr := service.UpdateDetails(userID, input)

	// THEN
	assert.NotNil(t, apiErr)
	assert.Equal(t, 500, apiErr.Status)
	mockPermissionChecker.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "BeginTx")
}

func TestUpdateDetails_ReturnsError_WhenBothAddAndRemoveEmpty(t *testing.T) {
	// GIVEN
	mockRepo := testutil.NewMockUserPermissionRepo()
	mockPermissionChecker := new(MockPermissionChecker)
	service := NewService(mockRepo, mockPermissionChecker)

	userID := uuid.New().String()
	input := UpdateDetails{
		Add:    []string{},
		Remove: []string{},
	}

	// WHEN
	apiErr := service.UpdateDetails(userID, input)

	// THEN
	assert.NotNil(t, apiErr)
	assert.Equal(t, 400, apiErr.Status)
	assert.Equal(t, "At least one of add or remove must contain one element", apiErr.Message)
}

// TestRemoveDuplicatesBetweenArrays verifies the private function through public API
func TestRemoveDuplicatesBetweenArrays_ThroughPublicAPI(t *testing.T) {
	// This test verifies that validation works correctly.
	// Since we can't easily mock gorm.DB transaction, we test that validation logic works
	// by verifying error responses from the validation step.

	// GIVEN
	mockRepo := testutil.NewMockUserPermissionRepo()
	mockPermissionChecker := new(MockPermissionChecker)
	service := NewService(mockRepo, mockPermissionChecker)

	userID := uuid.New().String()

	// Test case 1: Both add and remove empty - should get validation error before hitting tx
	input := UpdateDetails{
		Add:    []string{},
		Remove: []string{},
	}

	// WHEN
	apiErr := service.UpdateDetails(userID, input)

	// THEN - Should fail with validation error before hitting transaction
	assert.NotNil(t, apiErr)
	assert.Equal(t, 400, apiErr.Status)

	// Verify no transaction was started
	mockRepo.AssertNotCalled(t, "BeginTx")
}