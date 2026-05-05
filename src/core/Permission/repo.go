package permission

import (
	"context"
	"errors"

	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/generated"
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
	return gorm.G[model.Permission](r.db).Where(generated.Permission.ID.Eq(id)).First(context.Background())
}

func (r *repo) FindAll(opts *helper.FindAllOptions) ([]model.Permission, int64, error) {
	var finded []model.Permission
	
	// Build the base query with conditions
	query := r.db.Model(&model.Permission{})

	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		query = query.Where(
			generated.Permission.Name.ILike(searchPattern),
			generated.Permission.Description.ILike(searchPattern),
		)
	}

	// Count total using a SEPARATE query to avoid breaking the chain
	var total int64
	countQuery := r.db.Model(&model.Permission{})
	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		countQuery = countQuery.Where(
			generated.Permission.Name.ILike(searchPattern),
			generated.Permission.Description.ILike(searchPattern),
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

func (r *repo) AllExists(ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	var count int64
	if err := r.db.Model(&model.Permission{}).
		Where(generated.Permission.ID.In(ids...)).
		Count(&count).Error; err != nil {
		return err
	}

	if count != int64(len(ids)) {
		return errors.New("one or more permissions do not exist")
	}

	return nil
}
