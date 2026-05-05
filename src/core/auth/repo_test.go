package authentication

import (
	"fmt"
	"testing"
	"time"

	"github.com/MetaDandy/go-fiber-skeleton/src/enum"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Task 7.1: TestAuthRepo_Create (con provider)
func TestAuthRepo_Create_WithProvider(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create a new user via the repo
	role := createTestRole(t, db)
	newUser := model.User{
		ID:        uuid.New(),
		Name:      "OAuth User",
		Email:     "oauth@example.com",
		RoleID:    role.ID,
		Password:  nil, // OAuth users might not have password
	}

	authLog := model.AuthLog{
		ID:    uuid.New(),
		Event: enum.OAuthLogin,
		UserID: newUser.ID,
	}

	authProvider := &model.AuthProvider{
		ID:             uuid.New(),
		Provider:       "google",
		ProviderUserID: "google-123",
		UserID:         newUser.ID,
	}

	err := repo.Create(newUser, authLog, authProvider)

	assert.NoError(t, err)

	// Verify user was persisted
	var found model.User
	err = db.Where("id = ?", newUser.ID).First(&found).Error
	assert.NoError(t, err)
	assert.Equal(t, "OAuth User", found.Name)
	assert.Equal(t, "oauth@example.com", found.Email)

	// Verify auth provider was persisted
	var foundProvider model.AuthProvider
	err = db.Where("user_id = ? AND provider = ?", newUser.ID, "google").First(&foundProvider).Error
	assert.NoError(t, err)
	assert.Equal(t, "google-123", foundProvider.ProviderUserID)

	// Verify auth log was persisted
	var foundLog model.AuthLog
	err = db.Where("user_id = ? AND event = ?", newUser.ID, enum.OAuthLogin).First(&foundLog).Error
	assert.NoError(t, err)
}

// Task 7.2: TestAuthRepo_Create (sin provider)
func TestAuthRepo_Create_WithoutProvider(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	role := createTestRole(t, db)
	user := model.User{
		ID:        uuid.New(),
		Name:      "Local User",
		Email:     "local@example.com",
		RoleID:    role.ID,
		Password:  strPtr("hashedpassword"),
	}

	authLog := model.AuthLog{
		ID:    uuid.New(),
		Event: enum.LoginSuccess,
		UserID: user.ID,
	}

	err := repo.Create(user, authLog, nil)

	assert.NoError(t, err)

	// Verify user was persisted
	var found model.User
	err = db.Where("id = ?", user.ID).First(&found).Error
	assert.NoError(t, err)
	assert.Equal(t, "Local User", found.Name)

	// Verify no auth provider was created
	var count int64
	db.Model(&model.AuthProvider{}).Where("user_id = ?", user.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

// Task 7.3: TestAuthRepo_SaveEmailVerificationToken
func TestAuthRepo_SaveEmailVerificationToken(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	user := createTestUser(t, db)

	token := model.EmailVerificationToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: "hash-123",
	}

	err := repo.SaveEmailVerificationToken(token)

	assert.NoError(t, err)

	// Verify token was persisted
	var found model.EmailVerificationToken
	err = db.Where("token_hash = ?", "hash-123").First(&found).Error
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.UserID)
	assert.Nil(t, found.UsedAt)
}

// Task 7.4: TestAuthRepo_GetEmailVerificationToken (válido)
func TestAuthRepo_GetEmailVerificationToken_Valid(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	user := createTestUser(t, db)

	// Create a valid token
	token := model.EmailVerificationToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: "valid-hash",
		UsedAt:    nil,
	}
	db.Create(&token)

	// Retrieve it
	found, err := repo.GetEmailVerificationTokenByHash("valid-hash")

	assert.NoError(t, err)
	assert.Equal(t, token.ID, found.ID)
	assert.Equal(t, user.ID, found.UserID)
}

// Task 7.5: TestAuthRepo_GetEmailVerificationToken (usado - error)
func TestAuthRepo_GetEmailVerificationToken_Used(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	user := createTestUser(t, db)

	// Create a used token
	now := time.Now()
	token := model.EmailVerificationToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: "used-hash",
		UsedAt:    &now,
	}
	db.Create(&token)

	// Try to retrieve it - should fail because UsedAt is not null
	_, err := repo.GetEmailVerificationTokenByHash("used-hash")

	assert.Error(t, err)
}

// Task 7.6: TestAuthRepo_MarkEmailAsVerified
func TestAuthRepo_MarkEmailAsVerified(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	user := createTestUser(t, db)

	// Verify email is not verified initially
	var fresh model.User
	db.Where("id = ?", user.ID).First(&fresh)
	assert.False(t, fresh.EmailVerified)
	assert.Nil(t, fresh.EmailVerifiedAt)

	// Mark as verified
	err := repo.MarkEmailAsVerified(user.ID)

	assert.NoError(t, err)

	// Verify email is now verified
	db.Where("id = ?", user.ID).First(&fresh)
	assert.True(t, fresh.EmailVerified)
	assert.NotNil(t, fresh.EmailVerifiedAt)
}

// Task 7.7: TestAuthRepo_InvalidateOldEmailTokens
func TestAuthRepo_InvalidateOldEmailTokens(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	user := createTestUser(t, db)

	// Create multiple tokens - some used, some not
	now := time.Now()
	token1 := model.EmailVerificationToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: "hash-1",
		UsedAt:    nil,
	}
	token2 := model.EmailVerificationToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: "hash-2",
		UsedAt:    nil,
	}
	token3 := model.EmailVerificationToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: "hash-3",
		UsedAt:    &now, // already used
	}
	db.Create(&token1)
	db.Create(&token2)
	db.Create(&token3)

	// Invalidate old tokens
	err := repo.InvalidateOldEmailTokens(user.ID)

	assert.NoError(t, err)

	// Verify tokens 1 and 2 are now used
	var t1, t2, t3 model.EmailVerificationToken
	db.Where("id = ?", token1.ID).First(&t1)
	db.Where("id = ?", token2.ID).First(&t2)
	db.Where("id = ?", token3.ID).First(&t3)

	assert.NotNil(t, t1.UsedAt)
	assert.NotNil(t, t2.UsedAt)
	assert.NotNil(t, t3.UsedAt) // was already used
}

// Task 7.8: TestAuthRepo_GetPasswordResetToken (válido)
func TestAuthRepo_GetPasswordResetToken_Valid(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	user := createTestUser(t, db)

	// Create a valid password reset token
	token := model.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: "reset-hash-valid",
		UsedAt:    nil,
	}
	db.Create(&token)

	// Retrieve it
	found, err := repo.GetPasswordResetTokenByHash("reset-hash-valid")

	assert.NoError(t, err)
	assert.Equal(t, token.ID, found.ID)
	assert.Equal(t, user.ID, found.UserID)
}

// Task 7.9: TestAuthRepo_SavePasswordResetTokenWithLog
func TestAuthRepo_SavePasswordResetTokenWithLog(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	user := createTestUser(t, db)

	// Create token and log
	token := model.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: "reset-hash-new",
	}

	authLog := model.AuthLog{
		ID:    uuid.New(),
		Event: enum.PasswordReset,
		UserID: user.ID,
	}

	// Invalidate old tokens and save new one with log
	err := repo.SavePasswordResetTokenWithLog(token, authLog)

	assert.NoError(t, err)

	// Verify token was persisted
	var foundToken model.PasswordResetToken
	err = db.Where("token_hash = ?", "reset-hash-new").First(&foundToken).Error
	assert.NoError(t, err)
	assert.Equal(t, user.ID, foundToken.UserID)

	// Verify log was persisted
	var foundLog model.AuthLog
	err = db.Where("user_id = ? AND event = ?", user.ID, enum.PasswordReset).First(&foundLog).Error
	assert.NoError(t, err)
}

// Task 7.10: TestAuthRepo_CompletePasswordReset
func TestAuthRepo_CompletePasswordReset(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	user := createTestUser(t, db)

	// Create a reset token first
	token := model.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: "reset-hash-complete",
		UsedAt:    nil,
	}
	db.Create(&token)

	// Complete password reset
	newPasswordHash := "new-hashed-password"
	authLog := model.AuthLog{
		ID:    uuid.New(),
		Event: enum.PasswordResetSuccess,
		UserID: user.ID,
	}

	err := repo.CompletePasswordReset(user.ID, newPasswordHash, authLog)

	assert.NoError(t, err)

	// Verify password was updated
	var fresh model.User
	db.Where("id = ?", user.ID).First(&fresh)
	assert.Equal(t, newPasswordHash, *fresh.Password)

	// Verify token is now used
	var freshToken model.PasswordResetToken
	db.Where("id = ?", token.ID).First(&freshToken)
	assert.NotNil(t, freshToken.UsedAt)

	// Verify log was created
	var count int64
	db.Model(&model.AuthLog{}).Where("user_id = ? AND event = ?", user.ID, enum.PasswordResetSuccess).Count(&count)
	assert.Equal(t, int64(1), count)
}

// Task 7.11: TestAuthRepo_CreateOAuthUser
func TestAuthRepo_CreateOAuthUser(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	role := createTestRole(t, db)

	// Create OAuth state first
	state := createOAuthState(t, db, "google")

	user := model.User{
		ID:        uuid.New(),
		Name:      "OAuth User",
		Email:     "oauth2@example.com",
		RoleID:    role.ID,
		Password:  nil,
	}

	authLog := model.AuthLog{
		ID:    uuid.New(),
		Event: enum.OAuthLogin,
		UserID: user.ID,
	}

	authProvider := model.AuthProvider{
		ID:             uuid.New(),
		Provider:       "google",
		ProviderUserID: "google-456",
		UserID:         user.ID,
	}

	err := repo.CreateOAuthUser(user, authLog, authProvider, state)

	assert.NoError(t, err)

	// Verify user was created
	var found model.User
	err = db.Where("id = ?", user.ID).First(&found).Error
	assert.NoError(t, err)
	assert.Equal(t, "OAuth User", found.Name)

	// Verify auth provider was created
	var foundProvider model.AuthProvider
	err = db.Where("user_id = ? AND provider = ?", user.ID, "google").First(&foundProvider).Error
	assert.NoError(t, err)
	assert.Equal(t, "google-456", foundProvider.ProviderUserID)

	// Verify state was consumed (soft deleted)
	var oauthState model.OAuthState
	db.Where("state = ?", state).First(&oauthState)
	assert.NotNil(t, oauthState.DeletedAt)
}

// Task 7.12: TestAuthRepo_AddOAuthProviderToUser
func TestAuthRepo_AddOAuthProviderToUser(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	user := createTestUser(t, db)

	// Create OAuth state first
	state := createOAuthState(t, db, "github")

	authProvider := model.AuthProvider{
		ID:             uuid.New(),
		Provider:       "github",
		ProviderUserID: "github-789",
		UserID:         user.ID,
	}

	authLog := model.AuthLog{
		ID:    uuid.New(),
		Event: enum.OAuthLogin,
		UserID: user.ID,
	}

	err := repo.AddOAuthProviderToUser(user.ID, authProvider, authLog, state, "github")

	assert.NoError(t, err)

	// Verify provider was added
	var foundProvider model.AuthProvider
	err = db.Where("user_id = ? AND provider = ?", user.ID, "github").First(&foundProvider).Error
	assert.NoError(t, err)
	assert.Equal(t, "github-789", foundProvider.ProviderUserID)

	// Verify state was consumed
	var oauthState model.OAuthState
	db.Where("state = ?", state).First(&oauthState)
	assert.NotNil(t, oauthState.DeletedAt)
}

// Task 7.13: TestAuthRepo_ValidateOAuthState_Success
func TestAuthRepo_ValidateOAuthState_Success(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create a valid OAuth state
	state := "valid-state-123"
	oauthState := model.OAuthState{
		ID:        uuid.New(),
		State:     state,
		Provider:  "google",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	db.Create(&oauthState)

	// Validate the state
	err := repo.ValidateOAuthState(state, "google")

	assert.NoError(t, err)

	// Verify state was consumed (soft deleted)
	var fresh model.OAuthState
	db.Where("id = ?", oauthState.ID).First(&fresh)
	assert.NotNil(t, fresh.DeletedAt)
}

// Task 7.14: TestAuthRepo_ValidateOAuthState_Expired
func TestAuthRepo_ValidateOAuthState_Expired(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create an expired OAuth state
	state := "expired-state-456"
	oauthState := model.OAuthState{
		ID:        uuid.New(),
		State:     state,
		Provider:  "google",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // expired
	}
	db.Create(&oauthState)

	// Try to validate - should fail
	err := repo.ValidateOAuthState(state, "google")

	assert.Error(t, err)
}

// Task 7.15: TestAuthRepo_ValidateOAuthState_AlreadyConsumed
func TestAuthRepo_ValidateOAuthState_AlreadyConsumed(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create a consumed OAuth state
	now := time.Now()
	state := "consumed-state-789"
	oauthState := model.OAuthState{
		ID:        uuid.New(),
		State:     state,
		Provider:  "google",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		DeletedAt: &now, // already consumed
	}
	db.Create(&oauthState)

	// Try to validate - should fail because deleted_at is not null
	err := repo.ValidateOAuthState(state, "google")

	assert.Error(t, err)
}

// Task 7.16: TestAuthRepo_GetOAuthProviderByState
func TestAuthRepo_GetOAuthProviderByState(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create a valid OAuth state
	state := "provider-state-123"
	oauthState := model.OAuthState{
		ID:        uuid.New(),
		State:     state,
		Provider:  "github",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	db.Create(&oauthState)

	// Get provider by state
	provider, err := repo.GetOAuthProviderByState(state)

	assert.NoError(t, err)
	assert.Equal(t, "github", provider)
}

// Task 7.17: TestAuthRepo_ConsumeOAuthStateAndLog
func TestAuthRepo_ConsumeOAuthStateAndLog(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	user := createTestUser(t, db)

	// Create a valid OAuth state
	state := "consume-state-456"
	oauthState := model.OAuthState{
		ID:        uuid.New(),
		State:     state,
		Provider:  "google",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	db.Create(&oauthState)

	authLog := model.AuthLog{
		ID:    uuid.New(),
		Event: enum.OAuthLogin,
		UserID: user.ID,
	}

	// Consume state and create log
	err := repo.ConsumeOAuthStateAndLog(state, "google", authLog)

	assert.NoError(t, err)

	// Verify state was consumed
	var fresh model.OAuthState
	db.Where("id = ?", oauthState.ID).First(&fresh)
	assert.NotNil(t, fresh.DeletedAt)

	// Verify log was created
	var foundLog model.AuthLog
	err = db.Where("id = ?", authLog.ID).First(&foundLog).Error
	assert.NoError(t, err)
}

// Task 7.18: TestAuthRepo_CreateSession
func TestAuthRepo_CreateSession(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	user := createTestUser(t, db)

	session := model.Session{
		ID:               uuid.New(),
		UserID:           user.ID,
		Provider:         "local",
		RefreshTokenHash: "refresh-hash-123",
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		Ip:               "127.0.0.1",
		UserAgent:        "test-agent",
	}

	err := repo.CreateSession(session)

	assert.NoError(t, err)

	// Verify session was persisted
	var found model.Session
	err = db.Where("id = ?", session.ID).First(&found).Error
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.UserID)
	assert.Equal(t, "refresh-hash-123", found.RefreshTokenHash)
	assert.Nil(t, found.RevokedAt)
}

// Task 7.19: TestAuthRepo_GetUserPermissions (UNION role + user)
func TestAuthRepo_GetUserPermissions(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	// Create a role and assign permissions to it
	role := createTestRole(t, db)

	perm1 := createTestPermission(t, db, "read:users")
	perm2 := createTestPermission(t, db, "write:users")

	// Insert into RolePermissions
	rp1 := model.RolePermission{
		ID:           uuid.New(),
		RoleID:       role.ID,
		PermissionID: perm1.ID,
	}
	rp2 := model.RolePermission{
		ID:           uuid.New(),
		RoleID:       role.ID,
		PermissionID: perm2.ID,
	}
	db.Create(&rp1)
	db.Create(&rp2)

	// Insert into RoleEffectivePermissions (source is the same role)
	rep1 := model.RoleEffectivePermission{
		ID:           uuid.New(),
		RoleID:       role.ID,
		SourceRoleID: role.ID,
		PermissionID: perm1.ID,
	}
	rep2 := model.RoleEffectivePermission{
		ID:           uuid.New(),
		RoleID:       role.ID,
		SourceRoleID: role.ID,
		PermissionID: perm2.ID,
	}
	db.Create(&rep1)
	db.Create(&rep2)

	// Create a user WITH this role
	pwd := "hashedpassword"
	user := &model.User{
		ID:        uuid.New(),
		Name:      "Test User",
		Email:     fmt.Sprintf("test-%s@example.com", uuid.New().String()),
		Password:  &pwd,
		RoleID:    role.ID,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Assign direct user permissions
	perm3 := createTestPermission(t, db, "delete:users")
	assignUserPermission(t, db, user.ID, perm3.ID)

	// Get user permissions (UNION of role + user)
	permissions, err := repo.GetUserPermissions(user.ID)

	assert.NoError(t, err)
	assert.Len(t, permissions, 3)
	assert.Contains(t, permissions, "read:users")
	assert.Contains(t, permissions, "write:users")
	assert.Contains(t, permissions, "delete:users")
}

// Task 7.20: TestAuthRepo_RevokeAllUserSessions
func TestAuthRepo_RevokeAllUserSessions(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	user := createTestUser(t, db)

	// Create multiple sessions for the user
	session1 := createTestSession(t, db, user.ID)
	session2 := createTestSession(t, db, user.ID)
	_ = session2 // keep reference

	// Revoke all sessions
	err := repo.RevokeAllUserSessions(user.ID)

	assert.NoError(t, err)

	// Verify all sessions are revoked
	var s1, s2 model.Session
	db.Where("id = ?", session1.ID).First(&s1)
	db.Where("id = ?", session2.ID).First(&s2)

	assert.NotNil(t, s1.RevokedAt)
	assert.NotNil(t, s2.RevokedAt)
}

// Additional test: TestUserAuthProviders
func TestAuthRepo_UserAuthProviders(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	user := createTestUser(t, db)

	// Initially no providers
	providers := repo.UserAuthProviders(user.ID)
	assert.Len(t, providers, 0)

	// Add a provider
	ap := model.AuthProvider{
		ID:             uuid.New(),
		UserID:         user.ID,
		Provider:       "google",
		ProviderUserID: "google-123",
	}
	db.Create(&ap)

	providers = repo.UserAuthProviders(user.ID)
	assert.Len(t, providers, 1)
	assert.Equal(t, "google", providers[0])

	// Add another provider
	ap2 := model.AuthProvider{
		ID:             uuid.New(),
		UserID:         user.ID,
		Provider:       "github",
		ProviderUserID: "github-456",
	}
	db.Create(&ap2)

	providers = repo.UserAuthProviders(user.ID)
	assert.Len(t, providers, 2)
}

// Additional test: TestRevokeSession
func TestAuthRepo_RevokeSession(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	user := createTestUser(t, db)
	session := createTestSession(t, db, user.ID)

	// Verify session is not revoked
	var fresh model.Session
	db.Where("id = ?", session.ID).First(&fresh)
	assert.Nil(t, fresh.RevokedAt)

	// Revoke the session
	err := repo.RevokeSession(session.ID)

	assert.NoError(t, err)

	// Verify session is now revoked
	db.Where("id = ?", session.ID).First(&fresh)
	assert.NotNil(t, fresh.RevokedAt)
}

// Additional test: TestGetSessionByHash
func TestAuthRepo_GetSessionByHash(t *testing.T) {
	db := setupTestContainer(t)
	repo := NewRepo(db)

	user := createTestUser(t, db)
	session := createTestSession(t, db, user.ID)

	// Get session by hash
	found, err := repo.GetSessionByHash(session.RefreshTokenHash)

	assert.NoError(t, err)
	assert.Equal(t, session.ID, found.ID)
	assert.Equal(t, user.ID, found.UserID)

	// Try with wrong hash
	_, err = repo.GetSessionByHash("wrong-hash")
	assert.Error(t, err)
}

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}
