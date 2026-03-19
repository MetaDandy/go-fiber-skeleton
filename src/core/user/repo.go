package user

import (
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"gorm.io/gorm"
)

type Repo interface {
	Create(m model.User) error
	FindByID(id string) (model.User, error)
	FindAll(opts *helper.FindAllOptions) ([]model.User, int64, error)
	Update(m model.User) error
	Delete(id string) error

	Exists(id string) error
	ExistsByEmail(email string) error
}

type repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *repo {
	return &repo{db: db}
}

func (r *repo) Create(m model.User) error {
	return r.db.Create(&m).Error
}

func (r *repo) FindByID(id string) (model.User, error) {
	var user model.User
	err := r.db.First(&user, "id = ?", id).Error
	return user, err
}

func (r *repo) FindAll(opts *helper.FindAllOptions) ([]model.User, int64, error) {
	var finded []model.User
	query := r.db.Model(model.User{})

	if opts.Search != "" {
		query = query.Where(
			`name ILIKE ? OR email ILIKE ?`,
			"%"+opts.Search+"%",
			"%"+opts.Search+"%",
		)
	}

	var total int64
	query, total = opts.ApplyFindAllOptions(query)

	err := query.Find(&finded).Error
	return finded, total, err
}

func (r *repo) Update(m model.User) error {
	return r.db.Save(&m).Error
}

func (r *repo) Delete(id string) error {
	return r.db.Delete(&model.User{}, "id = ?", id).Error
}

func (r *repo) Exists(id string) error {
	return r.db.Select("id").First(&model.User{}, "id = ?", id).Error
}

func (r *repo) ExistsByEmail(email string) error {
	return r.db.Select("email").First(&model.User{}, "email = ?", email).Error
}
