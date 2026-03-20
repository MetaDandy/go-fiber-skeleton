package authentication

import (
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"gorm.io/gorm"
)

type Repo interface {
	UserAuthProviders(userId string) []string
	Create(u model.User, al model.AuthLog, ap *model.AuthProvider) error
}

type repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) Repo {
	return &repo{db: db}
}

func (r *repo) UserAuthProviders(userId string) []string {
	var authProviders []string
	r.db.Model(&model.AuthProvider{}).Where("user_id = ?", userId).Pluck("provider", &authProviders)
	return authProviders
}

func (r *repo) Create(u model.User, al model.AuthLog, ap *model.AuthProvider) error {
	if err := r.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(&u).Error; err != nil {
			return err
		}

		if err := tx.Create(&al).Error; err != nil {
			return err
		}

		if ap != nil {
			if err := tx.Create(ap).Error; err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
