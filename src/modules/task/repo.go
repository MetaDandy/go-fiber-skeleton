package task

import (
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"gorm.io/gorm"
)

type Repo interface {
	Create(m model.Task) error
	FindByID(id string) (model.Task, error)
	FindAll(opts *helper.FindAllOptions) ([]model.Task, int64, error)
	Update(m model.Task) error
	Delete(id string) error
}

type repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) Repo {
	return &repo{db: db}
}

func (r *repo) Create(m model.Task) error {
	return r.db.Create(&m).Error
}

func (r *repo) FindByID(id string) (model.Task, error) {
	var task model.Task
	err := r.db.Preload("User").First(&task, "id = ?", id).Error
	return task, err
}

func (r *repo) FindAll(opts *helper.FindAllOptions) ([]model.Task, int64, error) {
	var finded []model.Task
	query := r.db.Preload("User").Model(model.Task{})
	var total int64
	query, total = opts.ApplyFindAllOptions(query)

	err := query.Find(&finded).Error
	return finded, total, err
}

func (r *repo) Update(m model.Task) error {
	return r.db.Save(&m).Error
}

func (r *repo) Delete(id string) error {
	return r.db.Delete(&model.Task{}, "id = ?", id).Error
}
