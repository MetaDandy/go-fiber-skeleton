package permission

import (
	"fmt"
	"testing"

	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/stretchr/testify/assert"
)

// TestPermissionRepo_FindByID_Exists tests FindByID with a permission that exists
func TestPermissionRepo_FindByID_Exists(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create a test permission
	perm := model.Permission{
		ID:          "test-perm-1",
		Name:        "Test Permission",
		Description: "A test permission",
	}
	err := db.Create(&perm).Error
	assert.NoError(t, err, "Failed to create test permission")

	// Retrieve by ID
	found, err := repo.FindByID("test-perm-1")
	assert.NoError(t, err, "FindByID should not error for existing permission")
	assert.Equal(t, "test-perm-1", found.ID)
	assert.Equal(t, "Test Permission", found.Name)
}

// TestPermissionRepo_FindByID_NotExists tests FindByID with a permission that does not exist
func TestPermissionRepo_FindByID_NotExists(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Try to find a non-existent permission
	_, err := repo.FindByID("non-existent")
	assert.Error(t, err, "FindByID should error for non-existent permission")
}

// TestPermissionRepo_FindAll_ILIKE tests the ILIKE search functionality
func TestPermissionRepo_FindAll_ILIKE(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create test permissions
	perms := []model.Permission{
		{ID: "perm-1", Name: "User Management", Description: "Manage users"},
		{ID: "perm-2", Name: "Role Management", Description: "Manage roles"},
		{ID: "perm-3", Name: "Permission Management", Description: "Manage permissions"},
	}
	for _, p := range perms {
		err := db.Create(&p).Error
		assert.NoError(t, err)
	}

	// Search by name (case-insensitive)
	opts := &helper.FindAllOptions{Search: "management"}
	results, total, err := repo.FindAll(opts)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total, "Should find all 3 permissions with 'management'")
	assert.Len(t, results, 3, "Results slice should have 3 items")

	// Search with different case (ILIKE is case-insensitive)
	opts = &helper.FindAllOptions{Search: "MANAGEMENT"}
	results, total, err = repo.FindAll(opts)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total, "ILIKE should be case-insensitive")

	// Search by description
	opts = &helper.FindAllOptions{Search: "users"}
	results, total, err = repo.FindAll(opts)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total, "Should find 1 permission with 'users' in description")
}

// TestPermissionRepo_FindAll_Pagination tests pagination
func TestPermissionRepo_FindAll_Pagination(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create multiple test permissions
	for i := 1; i <= 5; i++ {
		perm := model.Permission{
			ID:   fmt.Sprintf("perm-%d", i),
			Name: fmt.Sprintf("Permission %d", i),
		}
		err := db.Create(&perm).Error
		assert.NoError(t, err)
	}

	// Test first page (limit 2, offset 0)
	opts := &helper.FindAllOptions{Limit: 2, Offset: 0}
	results, total, err := repo.FindAll(opts)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total, "Total should be 5")
	assert.Len(t, results, 2, "First page should have 2 items")

	// Test second page (limit 2, offset 2)
	opts = &helper.FindAllOptions{Limit: 2, Offset: 2}
	results, total, err = repo.FindAll(opts)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, results, 2, "Second page should have 2 items")

	// Test third page (limit 2, offset 4)
	opts = &helper.FindAllOptions{Limit: 2, Offset: 4}
	results, total, err = repo.FindAll(opts)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, results, 1, "Third page should have 1 item")
}

// TestPermissionRepo_AllExists_Valid tests AllExists with all valid IDs
func TestPermissionRepo_AllExists_Valid(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create test permissions
	perms := []model.Permission{
		{ID: "perm-a", Name: "A"},
		{ID: "perm-b", Name: "B"},
		{ID: "perm-c", Name: "C"},
	}
	for _, p := range perms {
		err := db.Create(&p).Error
		assert.NoError(t, err)
	}

	// Test with all existing IDs
	err := repo.AllExists([]string{"perm-a", "perm-b", "perm-c"})
	assert.NoError(t, err, "AllExists should succeed when all IDs exist")
}

// TestPermissionRepo_AllExists_Invalid tests AllExists with some invalid IDs
func TestPermissionRepo_AllExists_Invalid(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create test permissions
	perms := []model.Permission{
		{ID: "perm-a", Name: "A"},
		{ID: "perm-b", Name: "B"},
	}
	for _, p := range perms {
		err := db.Create(&p).Error
		assert.NoError(t, err)
	}

	// Test with one non-existent ID
	err := repo.AllExists([]string{"perm-a", "non-existent"})
	assert.Error(t, err, "AllExists should fail when some IDs don't exist")
	assert.Contains(t, err.Error(), "do not exist")
}

// TestPermissionRepo_AllExists_Empty tests AllExists with empty slice
func TestPermissionRepo_AllExists_Empty(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Test with empty slice - should return nil without DB call
	err := repo.AllExists([]string{})
	assert.NoError(t, err, "AllExists with empty slice should return nil")
}
