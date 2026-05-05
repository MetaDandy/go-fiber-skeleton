package user_permission

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/MetaDandy/go-fiber-skeleton/src/core/user"
)

// createTestUser creates a user for testing purposes
func createTestUser(t *testing.T, db *gorm.DB) model.User {
	t.Helper()
	userRepo := user.NewRepo(db)
	password := "hashedpassword"
	
	// Create a role first (seeded data only has permissions, not roles)
	roleID := uuid.New()
	db.Create(&model.Role{
		ID:   roleID,
		Name: "TestRole",
	})
	
	u := model.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Name:     "Test User",
		Password: &password,
		RoleID:   roleID,
	}
	err := userRepo.Create(u)
	assert.NoError(t, err)
	return u
}

// getSeededPermissionID returns a valid permission ID from the seeded permissions
func getSeededPermissionID(t *testing.T, db *gorm.DB, index int) string {
	t.Helper()
	var permissions []model.Permission
	db.Find(&permissions)
	assert.True(t, len(permissions) > index, "not enough seeded permissions: %d found, need index %d", len(permissions), index)
	return permissions[index].ID
}

// TestUserPermissionRepo_UpdatePermissionsTx_Add tests adding permissions (Task 6.1)
func TestUserPermissionRepo_UpdatePermissionsTx_Add(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create a test user
	usr := createTestUser(t, db)

	// Get seeded permission IDs (use first two from seed)
	perm1 := getSeededPermissionID(t, db, 0) // e.g., "permission.read"
	perm2 := getSeededPermissionID(t, db, 1) // e.g., "permission.list"

	// RED: Write test for adding permissions
	add := []model.UserPermission{
		{ID: uuid.New(), UserID: usr.ID, PermissionID: perm1},
		{ID: uuid.New(), UserID: usr.ID, PermissionID: perm2},
	}

	tx := repo.BeginTx()
	err := repo.UpdatePermissionsTx(tx, usr.ID, add, nil)
	assert.NoError(t, err)
	tx.Commit()

	// Verify permissions were added
	var count int64
	db.Model(&model.UserPermission{}).Where("user_id = ?", usr.ID).Count(&count)
	assert.Equal(t, int64(2), count)

	// Verify specific permissions exist
	var permissions []model.UserPermission
	db.Where("user_id = ?", usr.ID).Find(&permissions)
	permissionIDs := make(map[string]bool)
	for _, p := range permissions {
		permissionIDs[p.PermissionID] = true
	}
	assert.True(t, permissionIDs[perm1], "first permission should exist")
	assert.True(t, permissionIDs[perm2], "second permission should exist")
}

// TestUserPermissionRepo_UpdatePermissionsTx_Add_NoDuplicates tests that adding duplicate permissions doesn't create duplicates (Task 6.1 - triangulate)
func TestUserPermissionRepo_UpdatePermissionsTx_Add_NoDuplicates(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create a test user
	usr := createTestUser(t, db)

	// Get seeded permission IDs
	perm1 := getSeededPermissionID(t, db, 0)
	perm2 := getSeededPermissionID(t, db, 1)
	perm3 := getSeededPermissionID(t, db, 2)

	// Add permissions first time
	add1 := []model.UserPermission{
		{ID: uuid.New(), UserID: usr.ID, PermissionID: perm1},
		{ID: uuid.New(), UserID: usr.ID, PermissionID: perm2},
	}
	tx1 := repo.BeginTx()
	err := repo.UpdatePermissionsTx(tx1, usr.ID, add1, nil)
	assert.NoError(t, err)
	tx1.Commit()

	// Try to add same permissions again (should not create duplicates due to check)
	add2 := []model.UserPermission{
		{ID: uuid.New(), UserID: usr.ID, PermissionID: perm1},
		{ID: uuid.New(), UserID: usr.ID, PermissionID: perm2},
		{ID: uuid.New(), UserID: usr.ID, PermissionID: perm3}, // New permission
	}
	tx2 := repo.BeginTx()
	err = repo.UpdatePermissionsTx(tx2, usr.ID, add2, nil)
	assert.NoError(t, err)
	tx2.Commit()

	// Verify: should have 3 permissions (2 original + 1 new, no duplicates)
	var count int64
	db.Model(&model.UserPermission{}).Where("user_id = ?", usr.ID).Count(&count)
	assert.Equal(t, int64(3), count)
}

// TestUserPermissionRepo_UpdatePermissionsTx_Remove tests removing permissions (Task 6.2)
func TestUserPermissionRepo_UpdatePermissionsTx_Remove(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create a test user
	usr := createTestUser(t, db)

	// Get seeded permission IDs
	perm1 := getSeededPermissionID(t, db, 0)
	perm2 := getSeededPermissionID(t, db, 1)
	perm3 := getSeededPermissionID(t, db, 2)

	// First add permissions
	add := []model.UserPermission{
		{ID: uuid.New(), UserID: usr.ID, PermissionID: perm1},
		{ID: uuid.New(), UserID: usr.ID, PermissionID: perm2},
		{ID: uuid.New(), UserID: usr.ID, PermissionID: perm3},
	}
	tx1 := repo.BeginTx()
	err := repo.UpdatePermissionsTx(tx1, usr.ID, add, nil)
	assert.NoError(t, err)
	tx1.Commit()

	// Verify 3 permissions exist
	var countBefore int64
	db.Model(&model.UserPermission{}).Where("user_id = ?", usr.ID).Count(&countBefore)
	assert.Equal(t, int64(3), countBefore)

	// Now remove one permission (perm1)
	tx2 := repo.BeginTx()
	err = repo.UpdatePermissionsTx(tx2, usr.ID, nil, []string{perm1})
	assert.NoError(t, err)
	tx2.Commit()

	// Verify: should have 2 permissions (perm2 and perm3)
	var countAfter int64
	db.Model(&model.UserPermission{}).Where("user_id = ?", usr.ID).Count(&countAfter)
	assert.Equal(t, int64(2), countAfter)

	// Verify specific permissions remain
	var permissions []model.UserPermission
	db.Where("user_id = ?", usr.ID).Find(&permissions)
	permissionIDs := make(map[string]bool)
	for _, p := range permissions {
		permissionIDs[p.PermissionID] = true
	}
	assert.False(t, permissionIDs[perm1], "first permission should be removed")
	assert.True(t, permissionIDs[perm2], "second permission should still exist")
	assert.True(t, permissionIDs[perm3], "third permission should still exist")
}

// TestUserPermissionRepo_UpdatePermissionsTx_Remove_All tests removing all permissions (Task 6.2 - triangulate)
func TestUserPermissionRepo_UpdatePermissionsTx_Remove_All(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create a test user
	usr := createTestUser(t, db)

	// Get seeded permission IDs
	perm1 := getSeededPermissionID(t, db, 0)
	perm2 := getSeededPermissionID(t, db, 1)

	// First add permissions
	add := []model.UserPermission{
		{ID: uuid.New(), UserID: usr.ID, PermissionID: perm1},
		{ID: uuid.New(), UserID: usr.ID, PermissionID: perm2},
	}
	tx1 := repo.BeginTx()
	err := repo.UpdatePermissionsTx(tx1, usr.ID, add, nil)
	assert.NoError(t, err)
	tx1.Commit()

	// Remove all permissions
	tx2 := repo.BeginTx()
	err = repo.UpdatePermissionsTx(tx2, usr.ID, nil, []string{perm1, perm2})
	assert.NoError(t, err)
	tx2.Commit()

	// Verify: should have 0 permissions
	var count int64
	db.Model(&model.UserPermission{}).Where("user_id = ?", usr.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

// TestUserPermissionRepo_BeginTx tests beginning a transaction (Task 6.3)
func TestUserPermissionRepo_BeginTx(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// RED: Write test for BeginTx
	tx := repo.BeginTx()
	assert.NotNil(t, tx, "transaction should not be nil")

	// Verify it's actually a transaction by using it
	// (if it's not a valid transaction, this will fail)
	usr := createTestUser(t, db)
	perm1 := getSeededPermissionID(t, db, 0)
	add := []model.UserPermission{
		{ID: uuid.New(), UserID: usr.ID, PermissionID: perm1},
	}
	err := repo.UpdatePermissionsTx(tx, usr.ID, add, nil)
	assert.NoError(t, err)

	// Commit the transaction
	tx.Commit()

	// Verify the data was actually committed
	var count int64
	db.Model(&model.UserPermission{}).Where("user_id = ?", usr.ID).Count(&count)
	assert.Equal(t, int64(1), count)
}

// TestUserPermissionRepo_BeginTx_Rollback tests that transaction can be rolled back (Task 6.3 - triangulate)
func TestUserPermissionRepo_BeginTx_Rollback(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Begin a transaction
	tx := repo.BeginTx()
	assert.NotNil(t, tx, "transaction should not be nil")

	// Add a permission within the transaction
	usr := createTestUser(t, db)
	perm1 := getSeededPermissionID(t, db, 0)
	add := []model.UserPermission{
		{ID: uuid.New(), UserID: usr.ID, PermissionID: perm1},
	}
	err := repo.UpdatePermissionsTx(tx, usr.ID, add, nil)
	assert.NoError(t, err)

	// Rollback the transaction
	tx.Rollback()

	// Verify the data was NOT committed
	var count int64
	db.Model(&model.UserPermission{}).Where("user_id = ?", usr.ID).Count(&count)
	assert.Equal(t, int64(0), count, "data should not be committed after rollback")
}
