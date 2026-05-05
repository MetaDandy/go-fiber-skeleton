package role

import (
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/generated"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repo interface {
	BeginTx() *gorm.DB

	Create(role model.Role, rolePermissions []model.RolePermission, roleEffectivePermissions []model.RoleEffectivePermission) error
	FindByID(id uuid.UUID) (model.Role, error)
	FindAll(opts *helper.FindAllOptions) ([]model.Role, int64, error)
	UpdateHeader(role model.Role) error

	FindByIDTx(tx *gorm.DB, id uuid.UUID) (model.Role, error)
	UpdateRolePermissionsTx(tx *gorm.DB, roleID uuid.UUID, add []model.RolePermission, remove []string) error

	FindChildrenTx(tx *gorm.DB, roleID uuid.UUID) ([]model.Role, error)
	FindDescendantsOrderedTx(tx *gorm.DB, roleID uuid.UUID) ([]model.Role, error)

	DescendantsWithDirectPermissionTx(tx *gorm.DB, roleID uuid.UUID, permissionID string) ([]model.Role, error)

	DeleteDirectPermissionTx(tx *gorm.DB, roleID uuid.UUID, permissionID string) error
	DeleteEffectivePermissionTx(tx *gorm.DB, roleID uuid.UUID, permissionID string) error

	UpsertEffectivePermissionTx(tx *gorm.DB, rep model.RoleEffectivePermission) error
	DeleteEffectivePermissionBySourceTx(tx *gorm.DB, roleID uuid.UUID, permissionID string, sourceRoleID uuid.UUID) error

	HasEffectivePermissionTx(tx *gorm.DB, roleID uuid.UUID, permissionID string) (bool, error)
	HasDirectPermissionTx(tx *gorm.DB, roleID uuid.UUID, permissionID string) (bool, error)

	UpdateEffectivePermissionSourceTx(tx *gorm.DB, roleID uuid.UUID, permissionID string, oldSourceRoleID uuid.UUID, newSourceRoleID uuid.UUID) error
	CountDirectPermissionsNotInSetTx(tx *gorm.DB, roleID uuid.UUID, permissionIDs []string) (int64, error)

	// Batch methods
	UpsertEffectivePermissionsBatchTx(tx *gorm.DB, reps []model.RoleEffectivePermission) error
	DeleteEffectivePermissionsBySourceAndRolesTx(tx *gorm.DB, roleIDs []uuid.UUID, permissionID string, sourceRoleID uuid.UUID) error
	GetRolesWithDirectPermissionTx(tx *gorm.DB, roleIDs []uuid.UUID, permissionID string) ([]uuid.UUID, error)
	GetDirectPermissionsCountsTx(tx *gorm.DB, roleIDs []uuid.UUID) (map[uuid.UUID]int64, error)
	DeleteDirectPermissionsBatchTx(tx *gorm.DB, roleIDs []uuid.UUID, permissionID string) error
	DeleteOwnEffectivePermissionsTx(tx *gorm.DB, roleIDs []uuid.UUID, permissionID string) error
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

func (r *repo) FindByID(id uuid.UUID) (model.Role, error) {
	var role model.Role
	err := r.db.Preload("Role_permissions").Preload("Role_effective_permissions").
		Where(generated.Role.ID.Eq(id)).
		First(&role).Error
	return role, err
}

func (r *repo) FindAll(opts *helper.FindAllOptions) ([]model.Role, int64, error) {
	var finded []model.Role
	query := r.db.Model(model.Role{})

	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		query = query.Where(
			generated.Role.Name.ILike(searchPattern),
			generated.Role.Description.ILike(searchPattern),
		)
	}

	// Count total using a SEPARATE query to avoid breaking the chain
	var total int64
	countQuery := r.db.Model(model.Role{})
	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		countQuery = countQuery.Where(
			generated.Role.Name.ILike(searchPattern),
			generated.Role.Description.ILike(searchPattern),
		)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply Limit/Offset - use defaults if not set
	limit := int(opts.Limit)
	if limit == 0 {
		limit = 10 // Default limit
	}
	query = query.Limit(limit).Offset(int(opts.Offset))

	err := query.Find(&finded).Error
	return finded, total, err
}

func (r *repo) UpdateHeader(m model.Role) error {
	return r.db.Save(&m).Error
}

func (r *repo) UpdateDetails(roleID uuid.UUID, add []model.RolePermission, remove []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
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
	})
}

func (r *repo) BeginTx() *gorm.DB {
	return r.db.Begin()
}

func (r *repo) FindByIDTx(tx *gorm.DB, id uuid.UUID) (model.Role, error) {
	var role model.Role
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Role_permissions").
		Preload("Role_effective_permissions").
		First(&role, "id = ?", id).Error
	return role, err
}

func (r *repo) UpdateRolePermissionsTx(tx *gorm.DB, roleID uuid.UUID, add []model.RolePermission, remove []string) error {
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

func (r *repo) FindChildrenTx(tx *gorm.DB, roleID uuid.UUID) ([]model.Role, error) {
	var roles []model.Role
	err := tx.Where("role_id = ?", roleID).Find(&roles).Error
	return roles, err
}

func (r *repo) FindDescendantsOrderedTx(tx *gorm.DB, roleID uuid.UUID) ([]model.Role, error) {
	var result []model.Role
	query := `
		WITH RECURSIVE role_tree AS (
			SELECT * FROM roles WHERE role_id = ?
			UNION ALL
			SELECT r.* FROM roles r
			INNER JOIN role_tree rt ON rt.id = r.role_id
		)
		SELECT * FROM role_tree
	`
	err := tx.Raw(query, roleID).Scan(&result).Error
	return result, err
}

func (r *repo) DescendantsWithDirectPermissionTx(tx *gorm.DB, roleID uuid.UUID, permissionID string) ([]model.Role, error) {
	descendants, err := r.FindDescendantsOrderedTx(tx, roleID)
	if err != nil {
		return nil, err
	}
	if len(descendants) == 0 {
		return []model.Role{}, nil
	}

	ids := make([]uuid.UUID, 0, len(descendants))
	for _, d := range descendants {
		ids = append(ids, d.ID)
	}

	var roles []model.Role
	err = tx.
		Joins("JOIN rolepermissions rp ON rp.role_id = roles.id").
		Where("roles.id IN ? AND rp.permission_id = ?", ids, permissionID).
		Group("roles.id").
		Find(&roles).Error

	return roles, err
}

func (r *repo) HasDirectPermissionTx(tx *gorm.DB, roleID uuid.UUID, permissionID string) (bool, error) {
	var count int64
	err := tx.Model(&model.RolePermission{}).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Count(&count).Error
	return count > 0, err
}

func (r *repo) HasEffectivePermissionTx(tx *gorm.DB, roleID uuid.UUID, permissionID string) (bool, error) {
	var count int64
	err := tx.Model(&model.RoleEffectivePermission{}).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Count(&count).Error
	return count > 0, err
}

func (r *repo) DeleteDirectPermissionTx(tx *gorm.DB, roleID uuid.UUID, permissionID string) error {
	return tx.
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&model.RolePermission{}).Error
}

func (r *repo) DeleteEffectivePermissionTx(tx *gorm.DB, roleID uuid.UUID, permissionID string) error {
	return tx.
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&model.RoleEffectivePermission{}).Error
}

func (r *repo) DeleteEffectivePermissionBySourceTx(tx *gorm.DB, roleID uuid.UUID, permissionID string, sourceRoleID uuid.UUID) error {
	return tx.
		Where("role_id = ? AND permission_id = ? AND source_role_id = ?", roleID, permissionID, sourceRoleID).
		Delete(&model.RoleEffectivePermission{}).Error
}

func (r *repo) UpsertEffectivePermissionTx(tx *gorm.DB, rep model.RoleEffectivePermission) error {
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "role_id"}, {Name: "permission_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"source_role_id", "updated_at"}),
	}).Create(&rep).Error
}

func (r *repo) UpdateEffectivePermissionSourceTx(tx *gorm.DB, roleID uuid.UUID, permissionID string, oldSourceRoleID uuid.UUID, newSourceRoleID uuid.UUID) error {
	return tx.
		Model(&model.RoleEffectivePermission{}).
		Where("role_id = ? AND permission_id = ? AND source_role_id = ?", roleID, permissionID, oldSourceRoleID).
		Update("source_role_id", newSourceRoleID).Error
}

func (r *repo) CountDirectPermissionsNotInSetTx(tx *gorm.DB, roleID uuid.UUID, permissionIDs []string) (int64, error) {
	var count int64

	query := tx.Model(&model.RolePermission{}).Where("role_id = ?", roleID)

	if len(permissionIDs) > 0 {
		query = query.Where("permission_id NOT IN ?", permissionIDs)
	}

	err := query.Count(&count).Error
	return count, err
}

func (r *repo) UpsertEffectivePermissionsBatchTx(tx *gorm.DB, reps []model.RoleEffectivePermission) error {
	if len(reps) == 0 {
		return nil
	}
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "role_id"}, {Name: "permission_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"source_role_id", "updated_at"}),
	}).CreateInBatches(&reps, 100).Error
}

func (r *repo) DeleteEffectivePermissionsBySourceAndRolesTx(tx *gorm.DB, roleIDs []uuid.UUID, permissionID string, sourceRoleID uuid.UUID) error {
	if len(roleIDs) == 0 {
		return nil
	}
	return tx.Where("role_id IN ? AND permission_id = ? AND source_role_id = ?", roleIDs, permissionID, sourceRoleID).
		Delete(&model.RoleEffectivePermission{}).Error
}

func (r *repo) GetRolesWithDirectPermissionTx(tx *gorm.DB, roleIDs []uuid.UUID, permissionID string) ([]uuid.UUID, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}
	var ids []uuid.UUID
	err := tx.Model(&model.RolePermission{}).
		Where("role_id IN ? AND permission_id = ?", roleIDs, permissionID).
		Pluck("role_id", &ids).Error
	return ids, err
}

func (r *repo) GetDirectPermissionsCountsTx(tx *gorm.DB, roleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}
	type result struct {
		RoleID uuid.UUID
		Count  int64
	}
	var results []result
	err := tx.Model(&model.RolePermission{}).
		Select("role_id, count(permission_id) as count").
		Where("role_id IN ?", roleIDs).
		Group("role_id").
		Scan(&results).Error

	counts := make(map[uuid.UUID]int64)
	for _, res := range results {
		counts[res.RoleID] = res.Count
	}
	return counts, err
}

func (r *repo) DeleteDirectPermissionsBatchTx(tx *gorm.DB, roleIDs []uuid.UUID, permissionID string) error {
	if len(roleIDs) == 0 {
		return nil
	}
	return tx.Where("role_id IN ? AND permission_id = ?", roleIDs, permissionID).
		Delete(&model.RolePermission{}).Error
}

func (r *repo) DeleteOwnEffectivePermissionsTx(tx *gorm.DB, roleIDs []uuid.UUID, permissionID string) error {
	if len(roleIDs) == 0 {
		return nil
	}
	return tx.Exec("DELETE FROM roleeffectivepermissions WHERE role_id IN ? AND permission_id = ? AND source_role_id = role_id", roleIDs, permissionID).Error
}
