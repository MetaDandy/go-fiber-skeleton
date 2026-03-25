package role

import (
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"gorm.io/gorm"
)

type Repo interface {
	Create(m model.Role, rp []model.RolePermission) error
	FindByID(id string) (model.Role, error)
	FindAll(opts *helper.FindAllOptions) ([]model.Role, int64, error)
	UpdateHeader(m model.Role) error
	UpdateDetails(roleID string, add []model.RolePermission, remove []string) error
}

type repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *repo {
	return &repo{db: db}
}

func (r *repo) Create(m model.Role, rp []model.RolePermission) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&m).Error; err != nil {
			return err
		}

		if len(rp) > 0 {
			if err := tx.CreateInBatches(&rp, 50).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *repo) FindByID(id string) (model.Role, error) {
	var role model.Role
	err := r.db.Preload("Role_permissions").First(&role, "id = ?", id).Error
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
