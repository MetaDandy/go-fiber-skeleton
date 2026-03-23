package permission

import (
	"errors"

	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"gorm.io/gorm"
)

type Repo interface {
	FindByID(id string) (model.Permission, error)
	FindAll(opts *helper.FindAllOptions) ([]model.Permission, int64, error)
	AllExists(ids []string) error
}

type repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *repo {
	return &repo{db: db}
}

func (r *repo) FindByID(id string) (model.Permission, error) {
	var permission model.Permission
	err := r.db.First(&permission, "id = ?", id).Error
	return permission, err
}

func (r *repo) FindAll(opts *helper.FindAllOptions) ([]model.Permission, int64, error) {
	var finded []model.Permission
	query := r.db.Model(model.Permission{})

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

func (r *repo) AllExists(ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	var count int64
	if err := r.db.Model(&model.Permission{}).
		Where("id IN ?", ids).
		Count(&count).Error; err != nil {
		return err
	}

	if count != int64(len(ids)) {
		return errors.New("one or more permissions do not exist")
	}

	return nil
}
