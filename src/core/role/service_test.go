package role

import (
	"strings"
	"testing"

	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	permissionRepo "github.com/MetaDandy/go-fiber-skeleton/src/core/Permission"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// mockPermissionChecker adapts permission.Repo to the PermissionChecker interface
type mockPermissionChecker struct {
	repo interface {
		AllExists(ids []string) *api_error.Error
	}
}

func (m *mockPermissionChecker) AllExists(ids []string) *api_error.Error {
	return m.repo.AllExists(ids)
}

// Test 3.2: TestCreate_ReturnsError_WhenNoPermissions
func TestCreate_ReturnsError_WhenNoPermissions(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	permRepo := permissionRepo.NewRepo(db)
	checker := &mockPermissionChecker{repo: permRepo}

	svc := NewService(repo, checker)

	input := Create{
		Name:        "Test Role",
		Description: "Test Description",
		Permissions: []string{}, // Empty - should fail
	}

	err := svc.Create(input)
	assert.Error(t, err, "Create should fail with no permissions")
	assert.Contains(t, err.Message, "at least one direct permission")
}

// Test 3.3: TestCreate_ReturnsError_WhenPermissionAlreadyInherited
func TestCreate_ReturnsError_WhenPermissionAlreadyInherited(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	permRepo := permissionRepo.NewRepo(db)
	checker := &mockPermissionChecker{repo: permRepo}

	svc := NewService(repo, checker)

	// Create parent role via service with role.create permission
	parentInput := Create{
		Name:        "Parent Role Inherited",
		Description: "Parent role with permission",
		Permissions: []string{"role.create"},
	}
	createErr := svc.Create(parentInput)
	if createErr != nil {
		t.Logf("Parent creation error: %s - %s", createErr.Code, createErr.Message)
	}
	if createErr != nil {
		t.Fatalf("Parent role should be created successfully but got: %s", createErr.Message)
	}

	// Query parent back to get the actual ID assigned by the service
	var parentRole model.Role
	dbErr := db.Where("name = ?", "Parent Role Inherited").First(&parentRole).Error
	if dbErr != nil {
		t.Fatalf("Should find the created parent role: %v", dbErr)
	}
	parentIDStr := parentRole.ID.String()

	// Now try to create child via service with SAME permission that parent has inherited
	// This should fail because the permission is already inherited from parent
	childInput := Create{
		Name:        "Child Role",
		Permissions: []string{"role.create"}, // Same as parent's - should fail
		RoleID:      &parentIDStr,
	}
	childErr := svc.Create(childInput)

	// Verify error is returned
	if childErr == nil {
		t.Fatalf("Create should fail when permission is already inherited from parent")
	}
	if !strings.Contains(childErr.Message, "already inherited from parent or ancestor role") {
		t.Fatalf("Expected error about inherited permission, got: %s", childErr.Message)
	}
}

// Test 3.4: TestCreate_Success_WithParentRole
func TestCreate_Success_WithParentRole(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	permRepo := permissionRepo.NewRepo(db)
	checker := &mockPermissionChecker{repo: permRepo}

	svc := NewService(repo, checker)

	// Create parent role via service with role.list permission
	parentInput := Create{
		Name:        "Parent Role Test",
		Description: "Parent role with permission",
		Permissions: []string{"role.list"},
	}
	createErr := svc.Create(parentInput)
	if createErr != nil {
		t.Logf("Parent creation error: %s - %s", createErr.Code, createErr.Message)
	}
	if createErr != nil {
		t.Fatalf("Parent role should be created successfully but got: %s", createErr.Message)
	}

	// Query parent back to get the actual ID assigned by the service
	var parentRole model.Role
	dbErr := db.Where("name = ?", "Parent Role Test").First(&parentRole).Error
	if dbErr != nil {
		t.Fatalf("Should find the created parent role: %v", dbErr)
	}
	parentIDStr := parentRole.ID.String()

	// Create child role via service with DIFFERENT permission (not inherited from parent)
	childInput := Create{
		Name:        "Child Role Test",
		Permissions: []string{"role.create"}, // Different from parent's role.list - should succeed
		RoleID:      &parentIDStr,
	}
	childErr := svc.Create(childInput)
	if childErr != nil {
		t.Fatalf("Child role should be created successfully with non-inherited permission but got: %s", childErr.Message)
	}

	// Query child back to get its actual ID
	var childRole model.Role
	dbErr = db.Where("name = ?", "Child Role Test").First(&childRole).Error
	if dbErr != nil {
		t.Fatalf("Should find the created child role: %v", dbErr)
	}

	// Verify child has both inherited and own permissions
	childRet, findErr := repo.FindByID(childRole.ID)
	if findErr != nil {
		t.Fatalf("Failed to find child role: %v", findErr)
	}
	if len(childRet.Role_effective_permissions) != 2 {
		t.Fatalf("Child should have 2 effective permissions (1 inherited + 1 own), got %d", len(childRet.Role_effective_permissions))
	}
}

// Test 3.5: TestFindByID_ReturnsRole_WhenExists
func TestFindByID_ReturnsRole_WhenExists(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	permRepo := permissionRepo.NewRepo(db)
	checker := &mockPermissionChecker{repo: permRepo}

	svc := NewService(repo, checker)

	// Create a role directly via repo
	roleID := uuid.New()
	role := model.Role{
		ID:          roleID,
		Name:        "Find Me Role",
		Description: "To be found",
	}
	rp := []model.RolePermission{
		{ID: uuid.New(), RoleID: roleID, PermissionID: "role.create"},
	}
	rep := []model.RoleEffectivePermission{
		{ID: uuid.New(), RoleID: roleID, SourceRoleID: roleID, PermissionID: "role.create"},
	}
	err := repo.Create(role, rp, rep)
	assert.NoError(t, err)

	// Find via service - need to wait for container to be ready
	found, findErr := svc.FindByID(roleID.String())
	if findErr != nil {
		t.Logf("FindByID error: %v", findErr)
	}
	// The test might need adjustment based on seeded data
	assert.NotNil(t, found, "Should find created role")
	if found != nil {
		assert.Equal(t, "Find Me Role", found.Name)
	}
}

// Test 3.6: TestFindAll_ReturnsPaginatedRoles
func TestFindAll_ReturnsPaginatedRoles(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	permRepo := permissionRepo.NewRepo(db)
	checker := &mockPermissionChecker{repo: permRepo}

	svc := NewService(repo, checker)

	// Create test roles
	for i := 0; i < 5; i++ {
		roleID := uuid.New()
		role := model.Role{
			ID:   roleID,
			Name: "TestRole_" + string(rune('A'+i)),
		}
		rp := []model.RolePermission{
			{ID: uuid.New(), RoleID: roleID, PermissionID: "role.create"},
		}
		rep := []model.RoleEffectivePermission{
			{ID: uuid.New(), RoleID: roleID, SourceRoleID: roleID, PermissionID: "role.create"},
		}
		err := repo.Create(role, rp, rep)
		assert.NoError(t, err)
	}

	// Find roles - verify pagination works
	opts := &helper.FindAllOptions{Limit: 10, Offset: 0}
	result, err := svc.FindAll(opts)
	// Note: FindAll returns nil error and fills the Paginated result
	if err != nil {
		t.Logf("FindAll returned error: %v", err)
	}
	assert.NotNil(t, result)
	// Total should be >= 5 (our test roles + seeded)
	assert.GreaterOrEqual(t, result.Total, int64(5))
	// Verify pagination fields are set correctly
	assert.Equal(t, uint(10), result.Limit)
	assert.Equal(t, uint(0), result.Offset)
}

// Test 3.7: TestUpdateHeader_UpdatesRoleParent
func TestUpdateHeader_UpdatesRoleParent(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	permRepo := permissionRepo.NewRepo(db)
	checker := &mockPermissionChecker{repo: permRepo}

	svc := NewService(repo, checker)

	// Create first parent role via service
	parent1Input := Create{
		Name:        "First Parent Update",
		Permissions: []string{"role.create"},
	}
	err := svc.Create(parent1Input)
	if err != nil {
		t.Logf("First parent creation error: %s - %s", err.Code, err.Message)
	}
	if err != nil {
		t.Fatalf("First parent should be created: %s", err.Message)
	}

	// Query first parent back to get its actual ID
	var parent1 model.Role
	dbErr := db.Where("name = ?", "First Parent Update").First(&parent1).Error
	if dbErr != nil {
		t.Fatalf("Should find first parent: %v", dbErr)
	}
	parent1IDStr := parent1.ID.String()

	// Create second parent role via service
	parent2Input := Create{
		Name:        "Second Parent Update",
		Permissions: []string{"role.update"},
	}
	err = svc.Create(parent2Input)
	if err != nil {
		t.Fatalf("Second parent should be created: %s", err.Message)
	}

	// Query second parent back to get its actual ID
	var parent2 model.Role
	dbErr = db.Where("name = ?", "Second Parent Update").First(&parent2).Error
	if dbErr != nil {
		t.Fatalf("Should find second parent: %v", dbErr)
	}
	parent2IDStr := parent2.ID.String()

	// Create child role initially under first parent
	childInput := Create{
		Name:        "Child Role Update",
		Permissions: []string{"role.list"},
		RoleID:      &parent1IDStr,
	}
	err = svc.Create(childInput)
	if err != nil {
		t.Fatalf("Child should be created: %s", err.Message)
	}

	// Query child back to get its actual ID
	var child model.Role
	dbErr = db.Where("name = ?", "Child Role Update").First(&child).Error
	if dbErr != nil {
		t.Fatalf("Should find child: %v", dbErr)
	}
	childIDStr := child.ID.String()

	// Now call UpdateHeader to change the parent from first to second parent
	updateInput := UpdateHeader{
		RoleID:     &parent2IDStr,
		StrictMode: true,
	}
	updateErr := svc.UpdateHeader(childIDStr, updateInput)
	if updateErr != nil {
		t.Fatalf("UpdateHeader should succeed but got: %s", updateErr.Message)
	}

	// Verify the child's parent was changed
	updatedChild, findErr := repo.FindByID(child.ID)
	if findErr != nil {
		t.Fatalf("Failed to find updated child: %v", findErr)
	}
	if updatedChild.RoleID == nil {
		t.Fatal("Child should have a parent")
	}
	if *updatedChild.RoleID != parent2.ID {
		t.Fatalf("Child's parent should be second parent, got: %s", updatedChild.RoleID)
	}
}

// Test 3.8: TestUpdateHeader_NormalizesPermissionsOnParentChange
func TestUpdateHeader_NormalizesPermissionsOnParentChange(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	permRepo := permissionRepo.NewRepo(db)
	checker := &mockPermissionChecker{repo: permRepo}

	svc := NewService(repo, checker)

	// Create original parent role with role.create permission
	originalParentID := uuid.New()
	originalParent := model.Role{
		ID:   originalParentID,
		Name: "Original Parent",
	}
	originalParentRP := []model.RolePermission{
		{ID: uuid.New(), RoleID: originalParentID, PermissionID: "role.create"},
	}
	originalParentREP := []model.RoleEffectivePermission{
		{ID: uuid.New(), RoleID: originalParentID, SourceRoleID: originalParentID, PermissionID: "role.create"},
	}
	err := repo.Create(originalParent, originalParentRP, originalParentREP)
	assert.NoError(t, err)

	// Create new parent role with role.update permission (overlaps with child's direct)
	newParentID := uuid.New()
	newParent := model.Role{
		ID:   newParentID,
		Name: "New Parent",
	}
	newParentRP := []model.RolePermission{
		{ID: uuid.New(), RoleID: newParentID, PermissionID: "role.update"},
	}
	newParentREP := []model.RoleEffectivePermission{
		{ID: uuid.New(), RoleID: newParentID, SourceRoleID: newParentID, PermissionID: "role.update"},
	}
	err = repo.Create(newParent, newParentRP, newParentREP)
	assert.NoError(t, err)

	// Create child role with TWO direct permissions: role.update (will be normalized) and role.list (will remain)
	childID := uuid.New()
	child := model.Role{
		ID:     childID,
		Name:   "Child Role",
		RoleID: &originalParentID,
	}
	childRP := []model.RolePermission{
		{ID: uuid.New(), RoleID: childID, PermissionID: "role.update"}, // This will be normalized away
		{ID: uuid.New(), RoleID: childID, PermissionID: "role.list"},  // This will remain
	}
	childREP := []model.RoleEffectivePermission{
		{ID: uuid.New(), RoleID: childID, SourceRoleID: originalParentID, PermissionID: "role.create"},
		{ID: uuid.New(), RoleID: childID, SourceRoleID: childID, PermissionID: "role.update"},
		{ID: uuid.New(), RoleID: childID, SourceRoleID: childID, PermissionID: "role.list"},
	}
	err = repo.Create(child, childRP, childREP)
	assert.NoError(t, err)

	// Call UpdateHeader to change parent with strictMode=true
	// This should normalize permissions - remove child's direct role.update
	// because it's now inherited from new parent
	newParentIDStr := newParentID.String()
	input := UpdateHeader{
		RoleID:     &newParentIDStr,
		StrictMode: true,
	}

	var apiErr *api_error.Error
	apiErr = svc.UpdateHeader(childID.String(), input)
	if apiErr != nil {
		t.Logf("UpdateHeader returned error: Code=%s, Message=%s", apiErr.Code, apiErr.Message)
		if apiErr.Err != nil {
			t.Logf("Inner error: %v", apiErr.Err)
		}
	}

	// Verify the call succeeded - permission was normalized
	if apiErr != nil {
		t.Errorf("UpdateHeader should succeed but got error: Code=%s, Message=%s", apiErr.Code, apiErr.Message)
		return
	}

	// Verify the child's role now has new parent
	updatedChild, err := repo.FindByID(childID)
	assert.NoError(t, err)
	assert.NotNil(t, updatedChild.RoleID, "Child should have a parent")
	assert.Equal(t, newParentID, *updatedChild.RoleID)

	// Verify child's direct permissions were normalized (role.update should be gone from direct)
	hasDirectRoleUpdate := false
	for _, rp := range updatedChild.Role_permissions {
		if rp.PermissionID == "role.update" {
			hasDirectRoleUpdate = true
			break
		}
	}
	assert.False(t, hasDirectRoleUpdate, "Direct permission role.update should be removed due to parent overlap")
}

// Test 3.9: TestUpdateDetails_AddsPermissions_WithPropagation
func TestUpdateDetails_AddsPermissions_WithPropagation(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	permRepo := permissionRepo.NewRepo(db)
	checker := &mockPermissionChecker{repo: permRepo}

	svc := NewService(repo, checker)

	// Create a simple role with role.list as direct permission
	roleID := uuid.New()
	role := model.Role{
		ID:   roleID,
		Name: "Test Role",
	}
	rp := []model.RolePermission{
		{ID: uuid.New(), RoleID: roleID, PermissionID: "role.list"},
	}
	rep := []model.RoleEffectivePermission{
		{ID: uuid.New(), RoleID: roleID, SourceRoleID: roleID, PermissionID: "role.list"},
	}
	err := repo.Create(role, rp, rep)
	assert.NoError(t, err)

	// Call UpdateDetails to add a new permission
	input := UpdateDetails{
		Add: []string{"role.update"},
	}

	var apiErr *api_error.Error
	apiErr = svc.UpdateDetails(roleID.String(), input)
	if apiErr != nil {
		t.Logf("UpdateDetails returned error: %v, Message: %s", apiErr, apiErr.Message)
	}

	// Verify the permission was added successfully - use assert.Nil instead
	if apiErr != nil {
		t.Errorf("UpdateDetails should succeed but got error: %v, Message: %s", apiErr, apiErr.Message)
		return
	}

	// Verify the new permission exists in direct permissions
	updatedRole, err := repo.FindByID(roleID)
	assert.NoError(t, err)

	hasRoleUpdate := false
	for _, rp := range updatedRole.Role_permissions {
		if rp.PermissionID == "role.update" {
			hasRoleUpdate = true
			break
		}
	}
	assert.True(t, hasRoleUpdate, "Permission role.update should be added as direct permission")

	// Verify permission is in effective permissions (propagated)
	hasEffectiveRoleUpdate := false
	for _, ep := range updatedRole.Role_effective_permissions {
		if ep.PermissionID == "role.update" {
			hasEffectiveRoleUpdate = true
			break
		}
	}
	assert.True(t, hasEffectiveRoleUpdate, "Permission role.update should be in effective permissions")
}

// Test 3.10: TestUpdateDetails_RemovesPermissions_WithPropagation
func TestUpdateDetails_RemovesPermissions_WithPropagation(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	permRepo := permissionRepo.NewRepo(db)
	checker := &mockPermissionChecker{repo: permRepo}

	svc := NewService(repo, checker)

	// Create role with multiple direct permissions
	roleID := uuid.New()
	role := model.Role{
		ID:   roleID,
		Name: "Test Role",
	}
	rp := []model.RolePermission{
		{ID: uuid.New(), RoleID: roleID, PermissionID: "role.list"},
		{ID: uuid.New(), RoleID: roleID, PermissionID: "role.create"},
	}
	rep := []model.RoleEffectivePermission{
		{ID: uuid.New(), RoleID: roleID, SourceRoleID: roleID, PermissionID: "role.list"},
		{ID: uuid.New(), RoleID: roleID, SourceRoleID: roleID, PermissionID: "role.create"},
	}
	err := repo.Create(role, rp, rep)
	assert.NoError(t, err)

	// Call UpdateDetails to remove role.list (keeping role.create)
	input := UpdateDetails{
		Remove: []string{"role.list"},
	}

	var apiErr *api_error.Error
	apiErr = svc.UpdateDetails(roleID.String(), input)
	if apiErr != nil {
		t.Logf("UpdateDetails returned error: %v, Message: %s", apiErr, apiErr.Message)
	}

	// Verify the permission was removed successfully - use explicit check instead
	if apiErr != nil {
		t.Errorf("UpdateDetails should succeed but got error: %v, Message: %s", apiErr, apiErr.Message)
		return
	}

	// Verify role.list is no longer a direct permission
	updatedRole, err := repo.FindByID(roleID)
	assert.NoError(t, err)

	hasRoleList := false
	for _, rp := range updatedRole.Role_permissions {
		if rp.PermissionID == "role.list" {
			hasRoleList = true
			break
		}
	}
	assert.False(t, hasRoleList, "Permission role.list should be removed from direct permissions")

	// Verify role.create still exists as direct
	hasRoleCreate := false
	for _, rp := range updatedRole.Role_permissions {
		if rp.PermissionID == "role.create" {
			hasRoleCreate = true
			break
		}
	}
	assert.True(t, hasRoleCreate, "Permission role.create should still exist as direct permission")

	// Verify role.list is no longer in effective permissions (propagation removes it)
	hasEffectiveRoleList := false
	for _, ep := range updatedRole.Role_effective_permissions {
		if ep.PermissionID == "role.list" {
			hasEffectiveRoleList = true
			break
		}
	}
	assert.False(t, hasEffectiveRoleList, "Permission role.list should be removed from effective permissions")
}

// Test 3.11: TestUpdateDetails_ReturnsError_WhenRemovingLastDirectPermission
func TestUpdateDetails_ReturnsError_WhenRemovingLastDirectPermission(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	permRepo := permissionRepo.NewRepo(db)
	checker := &mockPermissionChecker{repo: permRepo}

	svc := NewService(repo, checker)

	// Create role with single direct permission
	roleID := uuid.New()
	role := model.Role{
		ID:   roleID,
		Name: "Solo Role",
	}
	rp := []model.RolePermission{
		{ID: uuid.New(), RoleID: roleID, PermissionID: "role.create"},
	}
	rep := []model.RoleEffectivePermission{
		{ID: uuid.New(), RoleID: roleID, SourceRoleID: roleID, PermissionID: "role.create"},
	}
	err := repo.Create(role, rp, rep)
	assert.NoError(t, err)

	// Try to remove the only direct permission
	input := UpdateDetails{
		Remove: []string{"role.create"},
	}

	err = svc.UpdateDetails(roleID.String(), input)
	// The service returns error on validation issues
	// Error message should contain info about direct permissions
	assert.NotNil(t, err, "Should return error when attempt violates rules")
}