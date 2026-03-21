package authentication

import (
	"time"

	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repo interface {
	UserAuthProviders(userId string) []string
	Create(u model.User, al model.AuthLog, ap *model.AuthProvider) error
	SaveEmailVerificationToken(token model.EmailVerificationToken) error
	GetEmailVerificationTokenByHash(tokenHash string) (model.EmailVerificationToken, error)
	MarkEmailAsVerified(userID uuid.UUID) error
	InvalidateOldEmailTokens(userID uuid.UUID) error
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

func (r *repo) SaveEmailVerificationToken(token model.EmailVerificationToken) error {
	return r.db.Create(&token).Error
}

func (r *repo) GetEmailVerificationTokenByHash(tokenHash string) (model.EmailVerificationToken, error) {
	var token model.EmailVerificationToken
	err := r.db.Where("token_hash = ? AND used_at IS NULL", tokenHash).First(&token).Error
	return token, err
}

func (r *repo) MarkEmailAsVerified(userID uuid.UUID) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"email_verified":    true,
		"email_verified_at": time.Now(),
	}).Error
}

func (r *repo) InvalidateOldEmailTokens(userID uuid.UUID) error {
	return r.db.Model(&model.EmailVerificationToken{}).
		Where("user_id = ? AND used_at IS NULL", userID).
		Update("used_at", time.Now()).Error
}
