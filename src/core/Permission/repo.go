package permission

import (
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"gorm.io/gorm"
)

type Repo interface {
	FindByID(id string) (model.Permission, error)
	FindAll(opts *helper.FindAllOptions) ([]model.Permission, int64, error)
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

	var total int64
	query, total = opts.ApplyFindAllOptions(query)

	err := query.Find(&finded).Error
	return finded, total, err
}
