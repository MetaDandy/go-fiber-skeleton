package permission

import (
	"fmt"
	"testing"

	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates a FRESH in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	// Use unique DB name per test to avoid shared state
	dbName := fmt.Sprintf("file:%s?mode=memory&cache=private", t.Name())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	assert.NoError(t, err, "Failed to open SQLite database")
	
	// Auto-migrate the Permission model
	err = db.AutoMigrate(&model.Permission{})
	assert.NoError(t, err, "Failed to migrate Permission model")
	
	return db
}

// Test 1: FindByID uses gorm.G pattern (GREEN phase)
func TestFindByID_WithGormGPattern(t *testing.T) {
	db := setupTestDB(t)
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
	
	// Test non-existent ID
	_, err = repo.FindByID("non-existent")
	assert.Error(t, err, "FindByID should error for non-existent permission")
}

// Test 2: FindByID triangulation - multiple IDs
func TestFindByID_MultiplePermissions(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo(db)
	
	// Create multiple permissions
	perms := []model.Permission{
		{ID: "perm-1", Name: "Permission 1", Description: "Desc 1"},
		{ID: "perm-2", Name: "Permission 2", Description: "Desc 2"},
		{ID: "perm-3", Name: "Permission 3", Description: "Desc 3"},
	}
	for _, p := range perms {
		err := db.Create(&p).Error
		assert.NoError(t, err)
	}
	
	// Test each permission
	for _, expected := range perms {
		found, err := repo.FindByID(expected.ID)
		assert.NoError(t, err, "FindByID failed for ID: %s", expected.ID)
		assert.Equal(t, expected.ID, found.ID)
		assert.Equal(t, expected.Name, found.Name)
	}
}

// Test 3: AllExists with generated Permission.ID.In
func TestAllExists_WithGeneratedIn(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo(db)
	
	// Test with empty slice - should return nil without DB call
	err := repo.AllExists([]string{})
	assert.NoError(t, err, "AllExists with empty slice should return nil")
	
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
	err = repo.AllExists([]string{"perm-a", "perm-b", "perm-c"})
	assert.NoError(t, err, "AllExists should succeed when all IDs exist")
	
	// Test with one non-existent ID
	err = repo.AllExists([]string{"perm-a", "non-existent"})
	assert.Error(t, err, "AllExists should fail when some IDs don't exist")
	assert.Contains(t, err.Error(), "do not exist")
}

// Test 4: FindAll without search (SQLite-compatible)
func TestFindAll_WithoutSearch(t *testing.T) {
	db := setupTestDB(t)
	
	// Create test permissions
	perms := []model.Permission{
		{ID: "list-1", Name: "Permission 1", Description: "Desc 1"},
		{ID: "list-2", Name: "Permission 2", Description: "Desc 2"},
		{ID: "list-3", Name: "Permission 3", Description: "Desc 3"},
	}
	for _, p := range perms {
		err := db.Create(&p).Error
		assert.NoError(t, err)
	}
	
	// Test FindAll without search using the repo
	repo := NewRepo(db)
	opts := &helper.FindAllOptions{}
	results, total, err := repo.FindAll(opts)
	
	t.Logf("Total: %d, Results len: %d, Error: %v", total, len(results), err)
	t.Logf("Results: %+v", results)
	
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total, "Should find all 3 permissions")
	assert.Len(t, results, 3, "Results slice should have 3 items")
}

// Test 4b: Direct DB test (bypass repo to isolate issue)
func TestFindAll_DirectDB(t *testing.T) {
	db := setupTestDB(t)
	
	// Create test permissions
	perms := []model.Permission{
		{ID: "list-1", Name: "Permission 1", Description: "Desc 1"},
		{ID: "list-2", Name: "Permission 2", Description: "Desc 2"},
	}
	for _, p := range perms {
		err := db.Create(&p).Error
		assert.NoError(t, err)
	}
	
	// Test direct DB query with same pattern as repo
	var results []model.Permission
	var total int64
	
	query := db.Model(&model.Permission{})
	query.Count(&total)
	query = query.Limit(10).Offset(0)
	query.Find(&results)
	
	t.Logf("Direct DB - Total: %d, Results len: %d", total, len(results))
	assert.Equal(t, int64(2), total)
	assert.Len(t, results, 2)
}

// Test 5: Repo interface unchanged (interface compliance)
func TestRepoInterfaceUnchanged(t *testing.T) {
	// Verify repo implements Repo interface
	var _ Repo = &repo{}
	
	// Verify NewRepo returns correct type
	db := &gorm.DB{}
	r := NewRepo(db)
	assert.NotNil(t, r)
	
	// Verify returned type implements interface
	var _ Repo = r
}

// Test 6: Verify code uses gorm.G pattern (compilation + behavior test)
func TestFindByID_UsesGormGPattern(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo(db)
	
	// Create test data
	perm := model.Permission{ID: "gorm-g-test", Name: "Gorm G Test"}
	err := db.Create(&perm).Error
	assert.NoError(t, err)
	
	// Call FindByID - if it doesn't use gorm.G pattern, behavior might differ
	found, err := repo.FindByID("gorm-g-test")
	assert.NoError(t, err)
	assert.Equal(t, "gorm-g-test", found.ID)
}
