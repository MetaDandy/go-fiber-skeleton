package testutil

import (
	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockUserRepo is a mock implementation of user.Repo interface for testing.
// Use NewMockUserRepo() to create an instance.
type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(u model.User) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *MockUserRepo) FindByID(id string) (model.User, error) {
	args := m.Called(id)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockUserRepo) FindByEmail(email string) (model.User, error) {
	args := m.Called(email)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockUserRepo) FindAll(opts *helper.FindAllOptions) ([]model.User, int64, error) {
	args := m.Called(opts)
	return args.Get(0).([]model.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepo) Update(u model.User) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *MockUserRepo) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepo) Exists(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepo) ExistsByEmail(email string) error {
	args := m.Called(email)
	return args.Error(0)
}

func (m *MockUserRepo) UpdatePassword(id string, passwordHash string) error {
	args := m.Called(id, passwordHash)
	return args.Error(0)
}

// NewMockUserRepo creates a new MockUserRepo for testing.
// Use mockRepo.On() to set up expectations and mockRepo.AssertExpectations(t) to verify.
func NewMockUserRepo() *MockUserRepo {
	return &MockUserRepo{}
}

// MockRoleRepo is a mock implementation of role.Repo interface for testing.
// This covers the basic methods; extend as needed for specific tests.
type MockRoleRepo struct {
	mock.Mock
}

func (m *MockRoleRepo) BeginTx() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockRoleRepo) Create(role model.Role, rolePermissions []model.RolePermission, roleEffectivePermissions []model.RoleEffectivePermission) error {
	args := m.Called(role, rolePermissions, roleEffectivePermissions)
	return args.Error(0)
}

func (m *MockRoleRepo) FindByID(id uuid.UUID) (model.Role, error) {
	args := m.Called(id)
	return args.Get(0).(model.Role), args.Error(1)
}

func (m *MockRoleRepo) FindAll(opts *helper.FindAllOptions) ([]model.Role, int64, error) {
	args := m.Called(opts)
	return args.Get(0).([]model.Role), args.Get(1).(int64), args.Error(2)
}

func (m *MockRoleRepo) UpdateHeader(role model.Role) error {
	args := m.Called(role)
	return args.Error(0)
}

func (m *MockRoleRepo) FindByIDTx(tx interface{}, id uuid.UUID) (model.Role, error) {
	args := m.Called(tx, id)
	return args.Get(0).(model.Role), args.Error(1)
}

func (m *MockRoleRepo) UpdateRolePermissionsTx(tx interface{}, roleID uuid.UUID, add []model.RolePermission, remove []string) error {
	args := m.Called(tx, roleID, add, remove)
	return args.Error(0)
}

func (m *MockRoleRepo) FindChildrenTx(tx interface{}, roleID uuid.UUID) ([]model.Role, error) {
	args := m.Called(tx, roleID)
	return args.Get(0).([]model.Role), args.Error(1)
}

func (m *MockRoleRepo) FindDescendantsOrderedTx(tx interface{}, roleID uuid.UUID) ([]model.Role, error) {
	args := m.Called(tx, roleID)
	return args.Get(0).([]model.Role), args.Error(1)
}

func (m *MockRoleRepo) DescendantsWithDirectPermissionTx(tx interface{}, roleID uuid.UUID, permissionID string) ([]model.Role, error) {
	args := m.Called(tx, roleID, permissionID)
	return args.Get(0).([]model.Role), args.Error(1)
}

func (m *MockRoleRepo) DeleteDirectPermissionTx(tx interface{}, roleID uuid.UUID, permissionID string) error {
	args := m.Called(tx, roleID, permissionID)
	return args.Error(0)
}

func (m *MockRoleRepo) DeleteEffectivePermissionTx(tx interface{}, roleID uuid.UUID, permissionID string) error {
	args := m.Called(tx, roleID, permissionID)
	return args.Error(0)
}

func (m *MockRoleRepo) UpsertEffectivePermissionTx(tx interface{}, rep model.RoleEffectivePermission) error {
	args := m.Called(tx, rep)
	return args.Error(0)
}

func (m *MockRoleRepo) DeleteEffectivePermissionBySourceTx(tx interface{}, roleID uuid.UUID, permissionID string, sourceRoleID uuid.UUID) error {
	args := m.Called(tx, roleID, permissionID, sourceRoleID)
	return args.Error(0)
}

func (m *MockRoleRepo) HasEffectivePermissionTx(tx interface{}, roleID uuid.UUID, permissionID string) (bool, error) {
	args := m.Called(tx, roleID, permissionID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRoleRepo) HasDirectPermissionTx(tx interface{}, roleID uuid.UUID, permissionID string) (bool, error) {
	args := m.Called(tx, roleID, permissionID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRoleRepo) UpdateEffectivePermissionSourceTx(tx interface{}, roleID uuid.UUID, permissionID string, oldSourceRoleID uuid.UUID, newSourceRoleID uuid.UUID) error {
	args := m.Called(tx, roleID, permissionID, oldSourceRoleID, newSourceRoleID)
	return args.Error(0)
}

func (m *MockRoleRepo) CountDirectPermissionsNotInSetTx(tx interface{}, roleID uuid.UUID, permissionIDs []string) (int64, error) {
	args := m.Called(tx, roleID, permissionIDs)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRoleRepo) UpsertEffectivePermissionsBatchTx(tx interface{}, reps []model.RoleEffectivePermission) error {
	args := m.Called(tx, reps)
	return args.Error(0)
}

func (m *MockRoleRepo) DeleteEffectivePermissionsBySourceAndRolesTx(tx interface{}, roleIDs []uuid.UUID, permissionID string, sourceRoleID uuid.UUID) error {
	args := m.Called(tx, roleIDs, permissionID, sourceRoleID)
	return args.Error(0)
}

func (m *MockRoleRepo) GetRolesWithDirectPermissionTx(tx interface{}, roleIDs []uuid.UUID, permissionID string) ([]uuid.UUID, error) {
	args := m.Called(tx, roleIDs, permissionID)
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func (m *MockRoleRepo) GetDirectPermissionsCountsTx(tx interface{}, roleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	args := m.Called(tx, roleIDs)
	return args.Get(0).(map[uuid.UUID]int64), args.Error(1)
}

func (m *MockRoleRepo) DeleteDirectPermissionsBatchTx(tx interface{}, roleIDs []uuid.UUID, permissionID string) error {
	args := m.Called(tx, roleIDs, permissionID)
	return args.Error(0)
}

func (m *MockRoleRepo) DeleteOwnEffectivePermissionsTx(tx interface{}, roleIDs []uuid.UUID, permissionID string) error {
	args := m.Called(tx, roleIDs, permissionID)
	return args.Error(0)
}

// NewMockRoleRepo creates a new MockRoleRepo for testing.
func NewMockRoleRepo() *MockRoleRepo {
	return &MockRoleRepo{}
}

// MockPermissionRepo is a mock implementation of permission.Repo interface for testing.
type MockPermissionRepo struct {
	mock.Mock
}

func (m *MockPermissionRepo) FindByID(id string) (model.Permission, error) {
	args := m.Called(id)
	return args.Get(0).(model.Permission), args.Error(1)
}

func (m *MockPermissionRepo) FindAll(opts *helper.FindAllOptions) ([]model.Permission, int64, error) {
	args := m.Called(opts)
	return args.Get(0).([]model.Permission), args.Get(1).(int64), args.Error(2)
}

func (m *MockPermissionRepo) AllExists(ids []string) *api_error.Error {
	args := m.Called(ids)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*api_error.Error)
}

// NewMockPermissionRepo creates a new MockPermissionRepo for testing.
func NewMockPermissionRepo() *MockPermissionRepo {
	return &MockPermissionRepo{}
}

// MockUserPermissionRepo is a mock implementation of user_permission.Repo interface for testing.
type MockUserPermissionRepo struct {
	mock.Mock
}

func (m *MockUserPermissionRepo) BeginTx() *gorm.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockUserPermissionRepo) UpdatePermissionsTx(tx *gorm.DB, userID uuid.UUID, add []model.UserPermission, remove []string) error {
	args := m.Called(tx, userID, add, remove)
	return args.Error(0)
}

// NewMockUserPermissionRepo creates a new MockUserPermissionRepo for testing.
func NewMockUserPermissionRepo() *MockUserPermissionRepo {
	return &MockUserPermissionRepo{}
}