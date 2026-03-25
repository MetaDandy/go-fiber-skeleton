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
	GetPasswordResetTokenByHash(tokenHash string) (model.PasswordResetToken, error)
	CreateAuthLog(al model.AuthLog) error
	SavePasswordResetTokenWithLog(prt model.PasswordResetToken, al model.AuthLog) error
	CompletePasswordReset(userID string, passwordHash string, al model.AuthLog) error
	CreateOAuthUser(u model.User, al model.AuthLog, ap model.AuthProvider) error
	GetOAuthProvider(userID uuid.UUID, provider string) error
	AddOAuthProviderToUser(userID uuid.UUID, ap model.AuthProvider, al model.AuthLog) error
	SaveOAuthState(state, provider string) error
	ValidateOAuthState(state, provider string) error
	GetOAuthProviderByState(state string) (string, error)
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

func (r *repo) GetPasswordResetTokenByHash(tokenHash string) (model.PasswordResetToken, error) {
	var token model.PasswordResetToken
	err := r.db.Where("token_hash = ? AND used_at IS NULL", tokenHash).First(&token).Error
	return token, err
}

func (r *repo) CreateAuthLog(al model.AuthLog) error {
	return r.db.Create(&al).Error
}

// SavePasswordResetTokenWithLog guarda token + log + invalida tokens anteriores CON TRANSACCIÓN Y ROLLBACK
func (r *repo) SavePasswordResetTokenWithLog(prt model.PasswordResetToken, al model.AuthLog) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Invalidar tokens anteriores del usuario
		if err := tx.Model(&model.PasswordResetToken{}).
			Where("user_id = ? AND used_at IS NULL", prt.UserID).
			Update("used_at", time.Now()).Error; err != nil {
			return err
		}

		// Guardar nuevo token
		if err := tx.Create(&prt).Error; err != nil {
			return err
		}

		// Guardar el log
		if err := tx.Create(&al).Error; err != nil {
			return err
		}

		return nil
	})
}

// CompletePasswordReset completa el reset: actualiza contraseña + invalida token + guarda log CON TRANSACCIÓN Y ROLLBACK
func (r *repo) CompletePasswordReset(userID string, passwordHash string, al model.AuthLog) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Actualizar contraseña del usuario
		if err := tx.Model(&model.User{}).Where("id = ?", userID).Update("password", passwordHash).Error; err != nil {
			return err
		}

		// Invalidar todos los tokens de reset del usuario
		if err := tx.Model(&model.PasswordResetToken{}).
			Where("user_id = ? AND used_at IS NULL", userID).
			Update("used_at", time.Now()).Error; err != nil {
			return err
		}

		// Guardar el log
		if err := tx.Create(&al).Error; err != nil {
			return err
		}

		return nil
	})
}

// CreateOAuthUser crea un nuevo usuario con OAuth: user + authlog + authprovider
// Operación transaccional - rollback si algo falla
func (r *repo) CreateOAuthUser(u model.User, al model.AuthLog, ap model.AuthProvider) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Crear usuario
		if err := tx.Create(&u).Error; err != nil {
			return err
		}

		// 2. Crear auth log
		if err := tx.Create(&al).Error; err != nil {
			return err
		}

		// 3. Crear auth provider
		if err := tx.Create(&ap).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetOAuthProvider busca si un usuario tiene un proveedor específico
// Retorna nil si existe, gorm.ErrRecordNotFound si no existe
func (r *repo) GetOAuthProvider(userID uuid.UUID, provider string) error {
	return r.db.Where("user_id = ? AND provider = ?", userID, provider).First(&model.AuthProvider{}).Error
}

// AddOAuthProviderToUser agrega un nuevo proveedor OAuth a un usuario existente
// Transacción: agrega provider + crea log
func (r *repo) AddOAuthProviderToUser(userID uuid.UUID, ap model.AuthProvider, al model.AuthLog) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Crear auth provider
		if err := tx.Create(&ap).Error; err != nil {
			return err
		}

		// 2. Crear auth log
		if err := tx.Create(&al).Error; err != nil {
			return err
		}

		return nil
	})
}

// SaveOAuthState guarda un estado de OAuth temporal para CSRF protection
func (r *repo) SaveOAuthState(state, provider string) error {
	oauthState := model.OAuthState{
		ID:        uuid.New(),
		State:     state,
		Provider:  provider,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	return r.db.Create(&oauthState).Error
}

// ValidateOAuthState valida un estado de OAuth y lo consume (one-time use)
// Retorna nil si válido, error si inválido/expirado/no encontrado
func (r *repo) ValidateOAuthState(state, provider string) error {
	var oauthState model.OAuthState

	// Buscar el estado y validar que no esté expirado
	if err := r.db.Where("state = ? AND provider = ? AND expires_at > ? AND deleted_at IS NULL",
		state, provider, time.Now()).
		First(&oauthState).Error; err != nil {
		return err
	}

	// Marcar como consumido (soft delete)
	return r.db.Model(&oauthState).Update("deleted_at", time.Now()).Error
}

// GetOAuthProviderByState retorna el provider asociado a un state válido
func (r *repo) GetOAuthProviderByState(state string) (string, error) {
	var oauthState model.OAuthState

	// Buscar el estado y validar que no esté expirado
	if err := r.db.Where("state = ? AND expires_at > ? AND deleted_at IS NULL",
		state, time.Now()).
		First(&oauthState).Error; err != nil {
		return "", err
	}

	return oauthState.Provider, nil
}
