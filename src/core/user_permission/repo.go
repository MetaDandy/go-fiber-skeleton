package user_permission

import (
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repo interface {
	BeginTx() *gorm.DB
	UpdatePermissionsTx(tx *gorm.DB, userID string, add []model.UserPermission, remove []string) error
}

type repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) Repo {
	return &repo{db: db}
}

func (r *repo) BeginTx() *gorm.DB {
	return r.db.Begin()
}

func (r *repo) UpdatePermissionsTx(tx *gorm.DB, userID string, add []model.UserPermission, remove []string) error {
	if len(add) > 0 {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(&add, 50).Error; err != nil {
			return err
		}
	}

	if len(remove) > 0 {
		if err := tx.Where("user_id = ? AND permission_id IN ?", userID, remove).
			Delete(&model.UserPermission{}).Error; err != nil {
			return err
		}
	}

	return nil
}
