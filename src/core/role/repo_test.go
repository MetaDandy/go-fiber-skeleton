package role

import (
	"testing"

	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/enum"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Test 4.1: TestRoleRepo_Create (solo rol)
func TestRoleRepo_Create(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	role := model.Role{
		ID:          uuid.New(),
		Name:        "Test Role",
		Description: "A test role",
	}

	err := repo.Create(role, nil, nil)
	assert.NoError(t, err)

	// Verify role was created
	var found model.Role
	err = db.First(&found, "id = ?", role.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "Test Role", found.Name)
	assert.Equal(t, "A test role", found.Description)
}

// Test 4.2: TestRoleRepo_Create (con rp + rep)
func TestRoleRepo_Create_WithPermissions(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	roleID := uuid.New()
	role := model.Role{
		ID:          roleID,
		Name:        "Admin Role",
		Description: "Role with permissions",
	}

	rp := []model.RolePermission{
		{ID: uuid.New(), RoleID: roleID, PermissionID: "role.create"},
		{ID: uuid.New(), RoleID: roleID, PermissionID: "role.update"},
	}

	rep := []model.RoleEffectivePermission{
		{ID: uuid.New(), RoleID: roleID, SourceRoleID: roleID, PermissionID: "role.create"},
		{ID: uuid.New(), RoleID: roleID, SourceRoleID: roleID, PermissionID: "role.update"},
	}

	err := repo.Create(role, rp, rep)
	assert.NoError(t, err)

	// Verify role was created
	var found model.Role
	err = db.First(&found, "id = ?", roleID).Error
	assert.NoError(t, err)
	assert.Equal(t, "Admin Role", found.Name)

	// Verify role permissions were created
	var permissions []model.RolePermission
	result := db.Where("role_id = ?", roleID).Find(&permissions)
	assert.NoError(t, result.Error)
	assert.Len(t, permissions, 2)

	// Verify role effective permissions were created
	var effectivePerms []model.RoleEffectivePermission
	result = db.Where("role_id = ?", roleID).Find(&effectivePerms)
	assert.NoError(t, result.Error)
	assert.Len(t, effectivePerms, 2)
}

// Test 4.3: TestRoleRepo_FindByID (con preloads)
func TestRoleRepo_FindByID(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create role with permissions
	roleID := uuid.New()
	role := model.Role{
		ID:          roleID,
		Name:        "Test Role",
		Description: "Test",
	}
	db.Create(&role)

	rp := model.RolePermission{
		ID: uuid.New(), RoleID: roleID, PermissionID: "role.create",
	}
	db.Create(&rp)

	rep := model.RoleEffectivePermission{
		ID: uuid.New(), RoleID: roleID, SourceRoleID: roleID, PermissionID: "role.create",
	}
	db.Create(&rep)

	// Find by ID with preloads
	found, err := repo.FindByID(roleID)
	assert.NoError(t, err)
	assert.Equal(t, roleID, found.ID)
	assert.Equal(t, "Test Role", found.Name)
	assert.Len(t, found.Role_permissions, 1)
	assert.Len(t, found.Role_effective_permissions, 1)
}

// Test 4.4: TestRoleRepo_FindByID_NotExists
func TestRoleRepo_FindByID_NotExists(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	nonExistentID := uuid.New()
	_, err := repo.FindByID(nonExistentID)
	assert.Error(t, err)
}

// Test 4.5: TestRoleRepo_FindAll_ILIKE
func TestRoleRepo_FindAll_ILIKE(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create test roles
	roles := []model.Role{
		{ID: uuid.New(), Name: "Admin Role", Description: "Administrator"},
		{ID: uuid.New(), Name: "User Role", Description: "Regular user"},
		{ID: uuid.New(), Name: "Guest", Description: "Guest access"},
	}
	for _, r := range roles {
		db.Create(&r)
	}

	// Search by name - case insensitive
	results, total, err := repo.FindAll(&helper.FindAllOptions{Search: "admin"})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, results, 1)
	assert.Equal(t, "Admin Role", results[0].Name)

	// Search by description - "user" appears in "User Role" name and "Regular user" description
	// Search for "user" now returns only "User Role" (1 result)
	results, total, err = repo.FindAll(&helper.FindAllOptions{Search: "user"})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, results, 1)
	assert.Equal(t, "User Role", results[0].Name)
}

// Test 4.6: TestRoleRepo_FindAll_Pagination
func TestRoleRepo_FindAll_Pagination(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create 20 roles
	for i := 0; i < 20; i++ {
		role := model.Role{
			ID:   uuid.New(),
			Name: "Role " + string(rune('A'+i)),
		}
		db.Create(&role)
	}

	// Get first page (10 items)
	results, total, err := repo.FindAll(&helper.FindAllOptions{Limit: 10, Offset: 0})
	assert.NoError(t, err)
	assert.Equal(t, int64(20), total)
	assert.Len(t, results, 10)

	// Get second page
	results, total, err = repo.FindAll(&helper.FindAllOptions{Limit: 10, Offset: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(20), total)
	assert.Len(t, results, 10)
}

// Test 4.7: TestRoleRepo_UpdateHeader
func TestRoleRepo_UpdateHeader(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create role
	roleID := uuid.New()
	role := model.Role{
		ID:          roleID,
		Name:        "Original Name",
		Description: "Original Description",
	}
	db.Create(&role)

	// Update header
	role.Name = "Updated Name"
	role.Description = "Updated Description"
	err := repo.UpdateHeader(role)
	assert.NoError(t, err)

	// Verify update
	var found model.Role
	db.First(&found, "id = ?", roleID)
	assert.Equal(t, "Updated Name", found.Name)
	assert.Equal(t, "Updated Description", found.Description)
}

// Test 4.8: TestRoleRepo_FindDescendantsOrderedTx (jerarquía 3 niveles)
func TestRoleRepo_FindDescendantsOrderedTx(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create 3-level hierarchy
	roles := createRoleHierarchy(t, db)
	parent := roles[0]
	child := roles[1]

	// Find descendants of parent
	descendants, err := repo.FindDescendantsOrderedTx(db, parent.ID)
	assert.NoError(t, err)
	assert.Len(t, descendants, 2)
	assert.Equal(t, child.ID, descendants[0].ID) // Child first (direct descendant)
}

// Test 4.9: TestRoleRepo_FindDescendantsOrderedTx_Leaf (sin hijos)
func TestRoleRepo_FindDescendantsOrderedTx_Leaf(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create a leaf role (no children)
	leafID := uuid.New()
	leaf := model.Role{
		ID:   leafID,
		Name: "Leaf Role",
	}
	db.Create(&leaf)

	// Find descendants of leaf
	descendants, err := repo.FindDescendantsOrderedTx(db, leafID)
	assert.NoError(t, err)
	assert.Len(t, descendants, 0)
}

// Test 4.10: TestRoleRepo_BeginTx
func TestRoleRepo_BeginTx(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	tx := repo.BeginTx()
	assert.NotNil(t, tx)

	// Verify it's a transaction by rolling back
	tx.Rollback()
}

// Test 4.11: TestRoleRepo_FindByIDTx (con lock)
func TestRoleRepo_FindByIDTx(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create role
	roleID := uuid.New()
	role := model.Role{
		ID:   roleID,
		Name: "Locked Role",
	}
	db.Create(&role)

	// Begin transaction and find with lock
	tx := repo.BeginTx()
	found, err := repo.FindByIDTx(tx, roleID)
	assert.NoError(t, err)
	assert.Equal(t, roleID, found.ID)
	assert.Equal(t, "Locked Role", found.Name)

	tx.Commit()
}

// Test 4.12: TestRoleRepo_UpdateRolePermissionsTx (add + remove)
func TestRoleRepo_UpdateRolePermissionsTx(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create role with initial permissions using the repo
	roleID := uuid.New()
	role := model.Role{
		ID:   roleID,
		Name: "Test Role",
	}

	// Use valid seeded permissions from enum
	initialPerms := []model.RolePermission{
		{ID: uuid.New(), RoleID: roleID, PermissionID: enum.RoleCreate.String()},
		{ID: uuid.New(), RoleID: roleID, PermissionID: enum.RoleUpdate.String()},
	}

	err := repo.Create(role, initialPerms, nil)
	assert.NoError(t, err)

	// Begin transaction
	tx := repo.BeginTx()

	// Add new permission (role.list exists in seed) and remove existing one
	add := []model.RolePermission{
		{ID: uuid.New(), RoleID: roleID, PermissionID: enum.RoleList.String()},
	}
	remove := []string{enum.RoleCreate.String()}

	err = repo.UpdateRolePermissionsTx(tx, roleID, add, remove)
	assert.NoError(t, err)

	tx.Commit()

	// Verify final state
	var perms []model.RolePermission
	db.Where("role_id = ?", roleID).Find(&perms)
	assert.Len(t, perms, 2) // role.update + role.list

	// Verify role.create was removed
	permIDs := make([]string, len(perms))
	for i, p := range perms {
		permIDs[i] = p.PermissionID
	}
	assert.NotContains(t, permIDs, enum.RoleCreate.String())
	assert.Contains(t, permIDs, enum.RoleUpdate.String())
	assert.Contains(t, permIDs, enum.RoleList.String())
}

// Test 4.13: TestRoleRepo_UpsertEffectivePermissionsBatchTx (insert)
func TestRoleRepo_UpsertEffectivePermissionsBatchTx_Insert(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	roleID := uuid.New()
	role := model.Role{
		ID:   roleID,
		Name: "Test Role",
	}
	db.Create(&role)

	tx := repo.BeginTx()

	reps := []model.RoleEffectivePermission{
		{ID: uuid.New(), RoleID: roleID, SourceRoleID: roleID, PermissionID: "role.create"},
		{ID: uuid.New(), RoleID: roleID, SourceRoleID: roleID, PermissionID: "role.update"},
	}

	err := repo.UpsertEffectivePermissionsBatchTx(tx, reps)
	assert.NoError(t, err)

	tx.Commit()

	// Verify insert
	var count int64
	db.Model(&model.RoleEffectivePermission{}).Where("role_id = ?", roleID).Count(&count)
	assert.Equal(t, int64(2), count)
}

// Test 4.14: TestRoleRepo_UpsertEffectivePermissionsBatchTx (update)
func TestRoleRepo_UpsertEffectivePermissionsBatchTx_Update(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create a real role to use as source_role_id
	sourceRole1 := model.Role{ID: uuid.New(), Name: "Source Role 1"}
	sourceRole2 := model.Role{ID: uuid.New(), Name: "Source Role 2"}
	db.Create(&sourceRole1)
	db.Create(&sourceRole2)

	roleID := uuid.New()
	role := model.Role{
		ID:   roleID,
		Name: "Test Role",
	}
	db.Create(&role)

	// Initial insert using valid source_role_id (seeded permission)
	tx := repo.BeginTx()
	reps := []model.RoleEffectivePermission{
		{ID: uuid.New(), RoleID: roleID, SourceRoleID: sourceRole1.ID, PermissionID: enum.RoleCreate.String()},
	}
	repo.UpsertEffectivePermissionsBatchTx(tx, reps)
	tx.Commit()

	// Update via upsert (same role_id + permission_id, different source)
	tx = repo.BeginTx()
	reps = []model.RoleEffectivePermission{
		{ID: uuid.New(), RoleID: roleID, SourceRoleID: sourceRole2.ID, PermissionID: enum.RoleCreate.String()},
	}
	err := repo.UpsertEffectivePermissionsBatchTx(tx, reps)
	assert.NoError(t, err)
	tx.Commit()

	// Verify source_role_id was updated
	var rep model.RoleEffectivePermission
	db.Where("role_id = ? AND permission_id = ?", roleID, enum.RoleCreate.String()).First(&rep)
	assert.Equal(t, sourceRole2.ID, rep.SourceRoleID)
}

// Test 4.15: TestRoleRepo_HasEffectivePermissionTx_True
func TestRoleRepo_HasEffectivePermissionTx_True(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	roleID := uuid.New()
	role := model.Role{
		ID:   roleID,
		Name: "Test Role",
	}
	db.Create(&role)

	// Add effective permission
	rep := model.RoleEffectivePermission{
		ID: uuid.New(), RoleID: roleID, SourceRoleID: roleID, PermissionID: "role.create",
	}
	db.Create(&rep)

	// Check if role has permission
	tx := repo.BeginTx()
	has, err := repo.HasEffectivePermissionTx(tx, roleID, "role.create")
	assert.NoError(t, err)
	assert.True(t, has)
	tx.Commit()
}

// Test 4.16: TestRoleRepo_HasEffectivePermissionTx_False
func TestRoleRepo_HasEffectivePermissionTx_False(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	roleID := uuid.New()
	role := model.Role{
		ID:   roleID,
		Name: "Test Role",
	}
	db.Create(&role)

	// Check if role has permission (it doesn't)
	tx := repo.BeginTx()
	has, err := repo.HasEffectivePermissionTx(tx, roleID, "role.create")
	assert.NoError(t, err)
	assert.False(t, has)
	tx.Commit()
}

// Test 4.17: TestRoleRepo_HasDirectPermissionTx_True
func TestRoleRepo_HasDirectPermissionTx_True(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	roleID := uuid.New()
	role := model.Role{
		ID:   roleID,
		Name: "Test Role",
	}
	db.Create(&role)

	// Add direct permission
	rp := model.RolePermission{
		ID: uuid.New(), RoleID: roleID, PermissionID: "role.create",
	}
	db.Create(&rp)

	// Check if role has direct permission
	tx := repo.BeginTx()
	has, err := repo.HasDirectPermissionTx(tx, roleID, "role.create")
	assert.NoError(t, err)
	assert.True(t, has)
	tx.Commit()
}

// Test 4.18: TestRoleRepo_HasDirectPermissionTx_False
func TestRoleRepo_HasDirectPermissionTx_False(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	roleID := uuid.New()
	role := model.Role{
		ID:   roleID,
		Name: "Test Role",
	}
	db.Create(&role)

	// Check if role has direct permission (it doesn't)
	tx := repo.BeginTx()
	has, err := repo.HasDirectPermissionTx(tx, roleID, "role.create")
	assert.NoError(t, err)
	assert.False(t, has)
	tx.Commit()
}

// Test 4.19: TestRoleRepo_DeleteDirectPermissionTx
func TestRoleRepo_DeleteDirectPermissionTx(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	roleID := uuid.New()
	role := model.Role{
		ID:   roleID,
		Name: "Test Role",
	}
	db.Create(&role)

	// Add direct permission
	rp := model.RolePermission{
		ID: uuid.New(), RoleID: roleID, PermissionID: "role.create",
	}
	db.Create(&rp)

	// Verify permission exists
	var count int64
	db.Model(&model.RolePermission{}).Where("role_id = ? AND permission_id = ?", roleID, "role.create").Count(&count)
	assert.Equal(t, int64(1), count)

	// Delete direct permission
	tx := repo.BeginTx()
	err := repo.DeleteDirectPermissionTx(tx, roleID, "role.create")
	assert.NoError(t, err)
	tx.Commit()

	// Verify permission was deleted
	db.Model(&model.RolePermission{}).Where("role_id = ? AND permission_id = ?", roleID, "role.create").Count(&count)
	assert.Equal(t, int64(0), count)
}

// Test 4.20: TestRoleRepo_DeleteEffectivePermissionTx
func TestRoleRepo_DeleteEffectivePermissionTx(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	roleID := uuid.New()
	role := model.Role{
		ID:   roleID,
		Name: "Test Role",
	}
	db.Create(&role)

	// Add effective permission
	rep := model.RoleEffectivePermission{
		ID: uuid.New(), RoleID: roleID, SourceRoleID: roleID, PermissionID: "role.create",
	}
	db.Create(&rep)

	// Delete effective permission
	tx := repo.BeginTx()
	err := repo.DeleteEffectivePermissionTx(tx, roleID, "role.create")
	assert.NoError(t, err)
	tx.Commit()

	// Verify permission was deleted
	var count int64
	db.Model(&model.RoleEffectivePermission{}).Where("role_id = ? AND permission_id = ?", roleID, "role.create").Count(&count)
	assert.Equal(t, int64(0), count)
}

// Test 4.21: TestRoleRepo_UpdateEffectivePermissionSourceTx
func TestRoleRepo_UpdateEffectivePermissionSourceTx(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	roleID := uuid.New()

	// Create real roles to use as source_role_id
	oldSource := model.Role{ID: uuid.New(), Name: "Old Source Role"}
	newSource := model.Role{ID: uuid.New(), Name: "New Source Role"}
	db.Create(&oldSource)
	db.Create(&newSource)

	role := model.Role{
		ID:   roleID,
		Name: "Test Role",
	}
	db.Create(&role)

	// Create effective permission with old source (using valid seeded permission)
	rep := model.RoleEffectivePermission{
		ID: uuid.New(), RoleID: roleID, SourceRoleID: oldSource.ID, PermissionID: enum.RoleCreate.String(),
	}
	db.Create(&rep)

	// Update source
	tx := repo.BeginTx()
	err := repo.UpdateEffectivePermissionSourceTx(tx, roleID, enum.RoleCreate.String(), oldSource.ID, newSource.ID)
	assert.NoError(t, err)
	tx.Commit()

	// Verify source was updated
	var updatedRep model.RoleEffectivePermission
	db.Where("role_id = ? AND permission_id = ?", roleID, enum.RoleCreate.String()).First(&updatedRep)
	assert.Equal(t, newSource.ID, updatedRep.SourceRoleID)
}

// Test 4.22: TestRoleRepo_CountDirectPermissionsNotInSetTx
func TestRoleRepo_CountDirectPermissionsNotInSetTx(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	roleID := uuid.New()
	role := model.Role{
		ID:   roleID,
		Name: "Test Role",
	}
	db.Create(&role)

	// Add REAL seeded permissions: role.create, role.update, role.list
	perms := []model.RolePermission{
		{ID: uuid.New(), RoleID: roleID, PermissionID: "role.create"},
		{ID: uuid.New(), RoleID: roleID, PermissionID: "role.update"},
		{ID: uuid.New(), RoleID: roleID, PermissionID: "role.list"},
	}
	db.Create(&perms)

	// Count permissions NOT IN set [role.create, role.update] - should be 1 (role.list)
	tx := repo.BeginTx()
	count, err := repo.CountDirectPermissionsNotInSetTx(tx, roleID, []string{"role.create", "role.update"})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
	tx.Commit()
}

// Test 4.23: TestRoleRepo_DescendantsWithDirectPermissionTx
func TestRoleRepo_DescendantsWithDirectPermissionTx(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create hierarchy: Parent -> Child -> Grandchild
	roles := createRoleHierarchy(t, db)
	parent := roles[0]
	child := roles[1]
	_ = roles[2]

	// Add direct permission "role.create" to child only
	rp := model.RolePermission{
		ID: uuid.New(), RoleID: child.ID, PermissionID: "role.create",
	}
	db.Create(&rp)

	// Find descendants with direct permission "role.create"
	tx := repo.BeginTx()
	descendants, err := repo.DescendantsWithDirectPermissionTx(tx, parent.ID, "role.create")
	assert.NoError(t, err)
	assert.Len(t, descendants, 1)
	assert.Equal(t, child.ID, descendants[0].ID)
	tx.Commit()
}

// Test 4.24: TestRoleRepo_DeleteOwnEffectivePermissionsTx
func TestRoleRepo_DeleteOwnEffectivePermissionsTx(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	roleID := uuid.New()
	role := model.Role{
		ID:   roleID,
		Name: "Test Role",
	}
	db.Create(&role)

	// Create a real source role for the "inherited" permission
	sourceRole := model.Role{ID: uuid.New(), Name: "Source Role"}
	db.Create(&sourceRole)

	// Add effective permissions where source_role_id = role_id (own permissions)
	ownPerm := model.RoleEffectivePermission{
		ID: uuid.New(), RoleID: roleID, SourceRoleID: roleID, PermissionID: enum.RoleCreate.String(),
	}
	db.Create(&ownPerm)

	// Add effective permission where source_role_id != role_id (inherited)
	inheritedPerm := model.RoleEffectivePermission{
		ID: uuid.New(), RoleID: roleID, SourceRoleID: sourceRole.ID, PermissionID: enum.RoleUpdate.String(),
	}
	db.Create(&inheritedPerm)

	// Delete own effective permissions for "role.create"
	tx := repo.BeginTx()
	err := repo.DeleteOwnEffectivePermissionsTx(tx, []uuid.UUID{roleID}, enum.RoleCreate.String())
	assert.NoError(t, err)
	tx.Commit()

	// Verify own permission was deleted
	var count int64
	db.Model(&model.RoleEffectivePermission{}).Where("role_id = ? AND permission_id = ?", roleID, enum.RoleCreate.String()).Count(&count)
	assert.Equal(t, int64(0), count)

	// Verify inherited permission still exists
	db.Model(&model.RoleEffectivePermission{}).Where("role_id = ? AND permission_id = ?", roleID, enum.RoleUpdate.String()).Count(&count)
	assert.Equal(t, int64(1), count)
}
