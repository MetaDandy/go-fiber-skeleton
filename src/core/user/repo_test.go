package user

import (
	"testing"

	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// RED phase: Write failing tests FIRST for all 10 User Repo tasks

// Task 3.1: TestUserRepo_Create
func TestUserRepo_Create(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create a role first to satisfy foreign key constraint
	role := model.Role{ID: uuid.New(), Name: "Test Role"}
	db.Create(&role)

	user := model.User{
		ID:     uuid.New(),
		Name:   "John Doe",
		Email:  "john@example.com",
		RoleID: role.ID,
	}
	err := repo.Create(user)

	assert.NoError(t, err)

	// Verify user was persisted
	var found model.User
	err = db.Where("id = ?", user.ID).First(&found).Error
	assert.NoError(t, err)
	assert.Equal(t, "John Doe", found.Name)
	assert.Equal(t, "john@example.com", found.Email)
}

// Task 3.2: TestUserRepo_FindByID_Exists
func TestUserRepo_FindByID_Exists(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create a user first
	user := createTestUser(t, db)

	// Find by ID
	found, err := repo.FindByID(user.ID.String())

	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, user.Name, found.Name)
	assert.Equal(t, user.Email, found.Email)
}

// Task 3.3: TestUserRepo_FindByID_NotExists
func TestUserRepo_FindByID_NotExists(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	nonExistentID := uuid.New().String()

	_, err := repo.FindByID(nonExistentID)

	assert.Error(t, err)
}

// Task 3.4: TestUserRepo_FindByEmail_Exists
func TestUserRepo_FindByEmail_Exists(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create a user first
	user := createTestUser(t, db)

	// Find by email
	found, err := repo.FindByEmail(user.Email)

	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, user.Email, found.Email)
}

// Task 3.5: TestUserRepo_FindByEmail_NotExists
func TestUserRepo_FindByEmail_NotExists(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	_, err := repo.FindByEmail("nonexistent@example.com")

	assert.Error(t, err)
}

// Task 3.6: TestUserRepo_FindAll_ILIKE
func TestUserRepo_FindAll_ILIKE(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create a role first
	role := model.Role{ID: uuid.New(), Name: "Test Role"}
	db.Create(&role)

	// Create users with similar names
	users := []model.User{
		{ID: uuid.New(), Name: "John Smith", Email: "john@example.com", RoleID: role.ID},
		{ID: uuid.New(), Name: "Jane Doe", Email: "jane@example.com", RoleID: role.ID},
		{ID: uuid.New(), Name: "johnny Bravo", Email: "johnny@example.com", RoleID: role.ID},
	}
	for i := range users {
		db.Create(&users[i])
	}

	// Search for "john" - should find "John Smith" and "johnny Bravo" (case-insensitive)
	opts := &helper.FindAllOptions{Search: "john"}
	results, total, err := repo.FindAll(opts)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, results, 2)
}

// Task 3.7: TestUserRepo_FindAll_Pagination
func TestUserRepo_FindAll_Pagination(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create a role first
	role := model.Role{ID: uuid.New(), Name: "Test Role"}
	db.Create(&role)

	// Create 20 users
	for i := 0; i < 20; i++ {
		user := model.User{
			ID:     uuid.New(),
			Name:   "User " + string(rune('A'+i)),
			Email:  "user" + string(rune('A'+i)) + "@example.com",
			RoleID: role.ID,
		}
		db.Create(&user)
	}

	// Get first page (limit 10)
	opts := &helper.FindAllOptions{Limit: 10, Offset: 0}
	results, total, err := repo.FindAll(opts)

	assert.NoError(t, err)
	assert.Equal(t, int64(20), total)
	assert.Len(t, results, 10)

	// Get second page
	opts2 := &helper.FindAllOptions{Limit: 10, Offset: 10}
	results2, total2, err2 := repo.FindAll(opts2)

	assert.NoError(t, err2)
	assert.Equal(t, int64(20), total2)
	assert.Len(t, results2, 10)
}

// Task 3.8: TestUserRepo_Update
func TestUserRepo_Update(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create a user first
	userPtr := createTestUser(t, db)
	user := *userPtr // Dereference pointer to match Update signature

	// Update user
	user.Name = "Updated Name"
	user.Email = "updated@example.com"
	err := repo.Update(user)

	assert.NoError(t, err)

	// Verify update
	var found model.User
	db.Where("id = ?", user.ID).First(&found)
	assert.Equal(t, "Updated Name", found.Name)
	assert.Equal(t, "updated@example.com", found.Email)
}

// Task 3.9: TestUserRepo_Delete_SoftDelete
func TestUserRepo_Delete_SoftDelete(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create a user first
	user := createTestUser(t, db)

	// Delete user (soft delete)
	err := repo.Delete(user.ID.String())

	assert.NoError(t, err)

	// Verify user is soft-deleted (not found with normal query)
	_, err = repo.FindByID(user.ID.String())
	assert.Error(t, err)

	// Verify user still exists with Unscoped
	var found model.User
	err = db.Unscoped().Where("id = ?", user.ID).First(&found).Error
	assert.NoError(t, err)
	assert.NotNil(t, found.DeletedAt)
}

// Task 3.10: TestUserRepo_UpdatePassword
func TestUserRepo_UpdatePassword(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create a user first
	user := createTestUser(t, db)
	oldPassword := *user.Password

	// Update password
	newHash := "newhashedpassword123"
	err := repo.UpdatePassword(user.ID.String(), newHash)

	assert.NoError(t, err)

	// Verify password was updated
	var found model.User
	db.Where("id = ?", user.ID).First(&found)
	assert.NotEqual(t, oldPassword, *found.Password)
	assert.Equal(t, newHash, *found.Password)
}
