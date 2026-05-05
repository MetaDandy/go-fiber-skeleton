package user_permission

import (
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repo interface {
	BeginTx() *gorm.DB
	UpdatePermissionsTx(tx *gorm.DB, userID uuid.UUID, add []model.UserPermission, remove []string) error
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

func (r *repo) UpdatePermissionsTx(tx *gorm.DB, userID uuid.UUID, add []model.UserPermission, remove []string) error {
	if len(add) > 0 {
		// Filter out permissions that already exist for this user
		var existingPerms []model.UserPermission
		existingIDs := make(map[string]bool)
		
		// Get existing permission IDs for this user
		if err := tx.Where("user_id = ?", userID).Find(&existingPerms).Error; err != nil {
			return err
		}
		for _, p := range existingPerms {
			existingIDs[p.PermissionID] = true
		}
		
		// Only add permissions that don't already exist
		var toAdd []model.UserPermission
		for _, p := range add {
			if !existingIDs[p.PermissionID] {
				toAdd = append(toAdd, p)
			}
		}
		
		if len(toAdd) > 0 {
			if err := tx.CreateInBatches(&toAdd, 50).Error; err != nil {
				return err
			}
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
