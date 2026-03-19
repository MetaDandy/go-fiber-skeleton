package authentication

import (
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"gorm.io/gorm"
)

type Repo interface {
	UserAuthProviders(userId string) []string
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
