package role

import (
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"gorm.io/gorm"
)

type Repo interface {
	BeginTx() *gorm.DB

	Create(role model.Role, rolePermissions []model.RolePermission, roleEffectivePermissions []model.RoleEffectivePermission) error
	FindByID(id string) (model.Role, error)
	FindAll(opts *helper.FindAllOptions) ([]model.Role, int64, error)
	UpdateHeader(role model.Role) error

	FindByIDTx(tx *gorm.DB, id string) (model.Role, error)
	UpdateRolePermissionsTx(tx *gorm.DB, roleID string, add []model.RolePermission, remove []string) error

	FindChildrenTx(tx *gorm.DB, roleID string) ([]model.Role, error)
	FindDescendantsOrderedTx(tx *gorm.DB, roleID string) ([]model.Role, error)

	DescendantsWithDirectPermissionTx(tx *gorm.DB, roleID string, permissionID string) ([]model.Role, error)

	DeleteDirectPermissionTx(tx *gorm.DB, roleID string, permissionID string) error
	DeleteEffectivePermissionTx(tx *gorm.DB, roleID string, permissionID string) error

	UpsertEffectivePermissionTx(tx *gorm.DB, rep model.RoleEffectivePermission) error
	DeleteEffectivePermissionBySourceTx(tx *gorm.DB, roleID string, permissionID string, sourceRoleID string) error

	HasEffectivePermissionTx(tx *gorm.DB, roleID string, permissionID string) (bool, error)
	HasDirectPermissionTx(tx *gorm.DB, roleID string, permissionID string) (bool, error)

	UpdateEffectivePermissionSourceTx(tx *gorm.DB, roleID string, permissionID string, oldSourceRoleID string, newSourceRoleID string) error
	CountDirectPermissionsNotInSetTx(tx *gorm.DB, roleID string, permissionIDs []string) (int64, error)
}

type repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *repo {
	return &repo{db: db}
}

func (r *repo) Create(
	m model.Role,
	rp []model.RolePermission,
	rep []model.RoleEffectivePermission,
) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&m).Error; err != nil {
			return err
		}

		if len(rp) > 0 {
			if err := tx.CreateInBatches(&rp, 50).Error; err != nil {
				return err
			}
		}

		if len(rep) > 0 {
			if err := tx.CreateInBatches(&rep, 50).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *repo) FindByID(id string) (model.Role, error) {
	var role model.Role
	err := r.db.Preload("Role_permissions").Preload("Role_effective_permissions").First(&role, "id = ?", id).Error
	return role, err
}

func (r *repo) FindAll(opts *helper.FindAllOptions) ([]model.Role, int64, error) {
	var finded []model.Role
	query := r.db.Model(model.Role{})

	if opts.Search != "" {
		query = query.Where(
			`name ILIKE ? OR description ILIKE ?`,
			"%"+opts.Search+"%",
			"%"+opts.Search+"%",
		)
	}

	var total int64
	query, total = opts.ApplyFindAllOptions(query)

	err := query.Find(&finded).Error
	return finded, total, err
}
func (r *repo) UpdateHeader(m model.Role) error {
	return r.db.Save(&m).Error
}
func (r *repo) UpdateDetails(roleID string, add []model.RolePermission, remove []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if len(add) > 0 {
			if err := tx.CreateInBatches(&add, 50).Error; err != nil {
				return err
			}
		}

		if len(remove) > 0 {
			if err := tx.
				Where("role_id = ? AND permission_id IN ?", roleID, remove).
				Delete(&model.RolePermission{}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
func (r *repo) BeginTx() *gorm.DB {
	return r.db.Begin()
}

func (r *repo) FindByIDTx(tx *gorm.DB, id string) (model.Role, error) {
	var role model.Role
	err := tx.
		Preload("Role_permissions").
		Preload("Role_effective_permissions").
		First(&role, "id = ?", id).Error
	return role, err
}

func (r *repo) UpdateRolePermissionsTx(tx *gorm.DB, roleID string, add []model.RolePermission, remove []string) error {
	if len(add) > 0 {
		if err := tx.Create(&add).Error; err != nil {
			return err
		}
	}

	if len(remove) > 0 {
		if err := tx.
			Where("role_id = ? AND permission_id IN ?", roleID, remove).
			Delete(&model.RolePermission{}).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *repo) FindChildrenTx(tx *gorm.DB, roleID string) ([]model.Role, error) {
	var roles []model.Role
	err := tx.Where("role_id = ?", roleID).Find(&roles).Error
	return roles, err
}

func (r *repo) FindDescendantsOrderedTx(tx *gorm.DB, roleID string) ([]model.Role, error) {
	var result []model.Role
	queue := []string{roleID}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		var children []model.Role
		if err := tx.Where("role_id = ?", current).Find(&children).Error; err != nil {
			return nil, err
		}

		for _, child := range children {
			result = append(result, child)
			queue = append(queue, child.ID.String())
		}
	}

	return result, nil
}
func (r *repo) DescendantsWithDirectPermissionTx(tx *gorm.DB, roleID string, permissionID string) ([]model.Role, error) {
	descendants, err := r.FindDescendantsOrderedTx(tx, roleID)
	if err != nil {
		return nil, err
	}
	if len(descendants) == 0 {
		return []model.Role{}, nil
	}

	ids := make([]string, 0, len(descendants))
	for _, d := range descendants {
		ids = append(ids, d.ID.String())
	}

	var roles []model.Role
	err = tx.
		Joins("JOIN rolepermissions rp ON rp.role_id = roles.id").
		Where("roles.id IN ? AND rp.permission_id = ?", ids, permissionID).
		Group("roles.id").
		Find(&roles).Error

	return roles, err
}
func (r *repo) HasDirectPermissionTx(tx *gorm.DB, roleID string, permissionID string) (bool, error) {
	var count int64
	err := tx.Model(&model.RolePermission{}).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Count(&count).Error
	return count > 0, err
}

func (r *repo) HasEffectivePermissionTx(tx *gorm.DB, roleID string, permissionID string) (bool, error) {
	var count int64
	err := tx.Model(&model.RoleEffectivePermission{}).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Count(&count).Error
	return count > 0, err
}

func (r *repo) DeleteDirectPermissionTx(tx *gorm.DB, roleID string, permissionID string) error {
	return tx.
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&model.RolePermission{}).Error
}

func (r *repo) DeleteEffectivePermissionTx(tx *gorm.DB, roleID string, permissionID string) error {
	return tx.
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&model.RoleEffectivePermission{}).Error
}

func (r *repo) DeleteEffectivePermissionBySourceTx(tx *gorm.DB, roleID string, permissionID string, sourceRoleID string) error {
	return tx.
		Where("role_id = ? AND permission_id = ? AND source_role_id = ?", roleID, permissionID, sourceRoleID).
		Delete(&model.RoleEffectivePermission{}).Error
}

func (r *repo) UpsertEffectivePermissionTx(tx *gorm.DB, rep model.RoleEffectivePermission) error {
	return tx.Create(&rep).Error
}

func (r *repo) UpdateEffectivePermissionSourceTx(tx *gorm.DB, roleID string, permissionID string, oldSourceRoleID string, newSourceRoleID string) error {
	return tx.
		Model(&model.RoleEffectivePermission{}).
		Where("role_id = ? AND permission_id = ? AND source_role_id = ?", roleID, permissionID, oldSourceRoleID).
		Update("source_role_id", newSourceRoleID).Error
}

func (r *repo) CountDirectPermissionsNotInSetTx(tx *gorm.DB, roleID string, permissionIDs []string) (int64, error) {
	var count int64

	query := tx.Model(&model.RolePermission{}).Where("role_id = ?", roleID)

	if len(permissionIDs) > 0 {
		query = query.Where("permission_id NOT IN ?", permissionIDs)
	}

	err := query.Count(&count).Error
	return count, err
}
