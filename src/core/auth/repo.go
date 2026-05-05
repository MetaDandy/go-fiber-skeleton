package authentication

import (
	"context"
	"time"

	"github.com/MetaDandy/go-fiber-skeleton/src/generated"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repo interface {
	UserAuthProviders(userId uuid.UUID) []string
	Create(u model.User, al model.AuthLog, ap *model.AuthProvider) error
	SaveEmailVerificationToken(token model.EmailVerificationToken) error
	GetEmailVerificationTokenByHash(tokenHash string) (model.EmailVerificationToken, error)
	MarkEmailAsVerified(userID uuid.UUID) error
	InvalidateOldEmailTokens(userID uuid.UUID) error
	GetPasswordResetTokenByHash(tokenHash string) (model.PasswordResetToken, error)
	CreateAuthLog(al model.AuthLog) error
	SavePasswordResetTokenWithLog(prt model.PasswordResetToken, al model.AuthLog) error
	CompletePasswordReset(userID uuid.UUID, passwordHash string, al model.AuthLog) error
	CreateOAuthUser(u model.User, al model.AuthLog, ap model.AuthProvider, state string) error
	GetOAuthProvider(userID uuid.UUID, provider string) error
	AddOAuthProviderToUser(userID uuid.UUID, ap model.AuthProvider, al model.AuthLog, state string, provider string) error
	SaveOAuthState(state, provider string) error
	ValidateOAuthState(state, provider string) error
	ConsumeOAuthStateAndLog(state, provider string, al model.AuthLog) error
	GetOAuthProviderByState(state string) (string, error)
	CreateSession(session model.Session) error
	GetSessionByHash(hash string) (model.Session, error)
	RevokeSession(id uuid.UUID) error
	RevokeAllUserSessions(userID uuid.UUID) error
	GetUserPermissions(userID uuid.UUID) ([]string, error)
}

type repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) Repo {
	return &repo{db: db}
}

func (r *repo) UserAuthProviders(userId uuid.UUID) []string {
	var authProviders []model.AuthProvider
	result := r.db.Where(generated.AuthProvider.UserID.Eq(userId)).Find(&authProviders)
	if result.Error != nil {
		return []string{}
	}
	providers := make([]string, 0, len(authProviders))
	for _, ap := range authProviders {
		providers = append(providers, ap.Provider)
	}
	return providers
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
	return gorm.G[model.EmailVerificationToken](r.db).
		Where(generated.EmailVerificationToken.TokenHash.Eq(tokenHash)).
		Where("used_at IS NULL").
		First(context.Background())
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
	return gorm.G[model.PasswordResetToken](r.db).
		Where(generated.PasswordResetToken.TokenHash.Eq(tokenHash)).
		Where("used_at IS NULL").
		First(context.Background())
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
func (r *repo) CompletePasswordReset(userID uuid.UUID, passwordHash string, al model.AuthLog) error {
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

// CreateOAuthUser crea un nuevo usuario con OAuth: user + authlog + authprovider + consumo de state
// Operación transaccional - rollback si algo falla, incluyendo la invalidación del state
func (r *repo) CreateOAuthUser(u model.User, al model.AuthLog, ap model.AuthProvider, state string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Consumir el state (atómico)
		if err := tx.Model(&model.OAuthState{}).
			Where("state = ? AND provider = ? AND deleted_at IS NULL", state, ap.Provider).
			Update("deleted_at", time.Now()).Error; err != nil {
			return err
		}

		// 2. Crear usuario
		if err := tx.Create(&u).Error; err != nil {
			return err
		}

		// 3. Crear auth log
		if err := tx.Create(&al).Error; err != nil {
			return err
		}

		// 4. Crear auth provider
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

// AddOAuthProviderToUser agrega un nuevo proveedor OAuth a un usuario existente y consume el state
// Transacción: consume state + agrega provider + crea log
func (r *repo) AddOAuthProviderToUser(userID uuid.UUID, ap model.AuthProvider, al model.AuthLog, state string, provider string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Consumir el state (atómico)
		if err := tx.Model(&model.OAuthState{}).
			Where("state = ? AND provider = ? AND deleted_at IS NULL", state, provider).
			Update("deleted_at", time.Now()).Error; err != nil {
			return err
		}

		// 2. Crear auth provider
		if err := tx.Create(&ap).Error; err != nil {
			return err
		}

		// 3. Crear auth log
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

// ValidateOAuthState validates an OAuth state and consumes it (one-time use)
// Returns nil if valid, error if invalid/expired/not found
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

// ConsumeOAuthStateAndLog atomically consumes an OAuth state and creates an auth log
// Used for the login-existing-provider case where user+provider already exist
func (r *repo) ConsumeOAuthStateAndLog(state, provider string, al model.AuthLog) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Consumir el state (soft-delete)
		if err := tx.Model(&model.OAuthState{}).
			Where("state = ? AND provider = ? AND deleted_at IS NULL", state, provider).
			Update("deleted_at", time.Now()).Error; err != nil {
			return err
		}

		// 2. Crear auth log
		if err := tx.Create(&al).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetOAuthProviderByState retorna el provider asociado a un state válido
func (r *repo) GetOAuthProviderByState(state string) (string, error) {
	oauthState, err := gorm.G[model.OAuthState](r.db).
		Where(generated.OAuthState.State.Eq(state)).
		Where("expires_at > ? AND deleted_at IS NULL", time.Now()).
		First(context.Background())
	if err != nil {
		return "", err
	}

	return oauthState.Provider, nil
}

func (r *repo) CreateSession(session model.Session) error {
	return r.db.Create(&session).Error
}

func (r *repo) GetSessionByHash(hash string) (model.Session, error) {
	return gorm.G[model.Session](r.db).
		Where(generated.Session.RefreshTokenHash.Eq(hash)).
		Where("revoked_at IS NULL").
		First(context.Background())
}

func (r *repo) RevokeSession(id uuid.UUID) error {
	_, err := gorm.G[model.Session](r.db).
		Where(generated.Session.ID.Eq(id)).
		Update(context.Background(), "revoked_at", time.Now())
	return err
}

func (r *repo) RevokeAllUserSessions(userID uuid.UUID) error {
	_, err := gorm.G[model.Session](r.db).
		Where(generated.Session.UserID.Eq(userID)).
		Where("revoked_at IS NULL").
		Update(context.Background(), "revoked_at", time.Now())
	return err
}
func (r *repo) GetUserPermissions(userID uuid.UUID) ([]string, error) {
	var permissions []string
	query := `
		SELECT roleeffectivepermissions.permission_id 
		FROM roleeffectivepermissions
		INNER JOIN users ON users.role_id = roleeffectivepermissions.role_id
		WHERE users.id = ?
		UNION
		SELECT permission_id 
		FROM userpermissions 
		WHERE user_id = ?
		`
	err := r.db.Raw(query, userID, userID).Scan(&permissions).Error
	return permissions, err
}