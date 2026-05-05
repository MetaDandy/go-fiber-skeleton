package user

import (
	"context"

	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/generated"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repo interface {
	Create(m model.User) error
	FindByID(id string) (model.User, error)
	FindByEmail(email string) (model.User, error)
	FindAll(opts *helper.FindAllOptions) ([]model.User, int64, error)
	Update(m model.User) error
	Delete(id string) error

	Exists(id string) error
	ExistsByEmail(email string) error
	UpdatePassword(id string, passwordHash string) error
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
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return model.User{}, err
	}
	return gorm.G[model.User](r.db).Where(generated.User.ID.Eq(parsedID)).First(context.Background())
}

func (r *repo) FindByEmail(email string) (model.User, error) {
	return gorm.G[model.User](r.db).Where(generated.User.Email.Eq(email)).First(context.Background())
}

func (r *repo) FindAll(opts *helper.FindAllOptions) ([]model.User, int64, error) {
	var finded []model.User
	query := r.db.Model(&model.User{})

	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		query = query.Where(
			generated.User.Name.ILike(searchPattern),
			generated.User.Email.ILike(searchPattern),
		)
	}

	// Ordering is explicit per repo - User defaults to created_at desc
	query = query.Order("created_at desc")

	// Count total using a SEPARATE query to avoid breaking the chain
	var total int64
	countQuery := r.db.Model(&model.User{})
	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		countQuery = countQuery.Where(
			generated.User.Name.ILike(searchPattern),
			generated.User.Email.ILike(searchPattern),
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

func (r *repo) Update(m model.User) error {
	return r.db.Save(&m).Error
}

func (r *repo) Delete(id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return r.db.Delete(&model.User{}, "id = ?", parsedID).Error
}

func (r *repo) Exists(id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	_, err = gorm.G[model.User](r.db).Select("id").Where(generated.User.ID.Eq(parsedID)).First(context.Background())
	return err
}

func (r *repo) ExistsByEmail(email string) error {
	_, err := gorm.G[model.User](r.db).Select("email").Where(generated.User.Email.Eq(email)).First(context.Background())
	return err
}

func (r *repo) UpdatePassword(id string, passwordHash string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return r.db.Model(&model.User{}).Where("id = ?", parsedID).Update("password", passwordHash).Error
}
