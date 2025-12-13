package task

import (
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"gorm.io/gorm"
)

type TaskRepo interface {
	Create(m model.Task) error
	FindByID(id string) (model.Task, error)
	FindAll(opts *helper.FindAllOptions) ([]model.Task, int64, error)
	Update(m model.Task) error
	Delete(id string) error
}

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(m model.Task) error {
	return r.db.Create(&m).Error
}

func (r *Repo) FindByID(id string) (model.Task, error) {
	var task model.Task
	err := r.db.Preload("User").First(&task, "id = ?", id).Error
	return task, err
}

func (r *Repo) FindAll(opts *helper.FindAllOptions) ([]model.Task, int64, error) {
	var finded []model.Task
	query := r.db.Preload("User").Model(model.Task{})
	var total int64
	query, total = helper.ApplyFindAllOptions(query, opts)

	err := query.Find(&finded).Error
	return finded, total, err
}

func (r *Repo) Update(m model.Task) error {
	return r.db.Save(&m).Error
}

func (r *Repo) Delete(id string) error {
	return r.db.Delete(&model.Task{}, "id = ?", id).Error
}
