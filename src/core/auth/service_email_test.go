package authentication

import (
	"context"
	"testing"
	"time"

	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock for Repo
type MockEmailRepo struct {
	mock.Mock
}

func (m *MockEmailRepo) UserAuthProviders(userId uuid.UUID) []string {
	args := m.Called(userId)
	if args.Get(0) != nil {
		return args.Get(0).([]string)
	}
	return []string{}
}

func (m *MockEmailRepo) Create(u model.User, al model.AuthLog, ap *model.AuthProvider) error {
	args := m.Called(u, al, ap)
	return args.Error(0)
}

func (m *MockEmailRepo) SaveEmailVerificationToken(token model.EmailVerificationToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockEmailRepo) GetEmailVerificationTokenByHash(tokenHash string) (model.EmailVerificationToken, error) {
	args := m.Called(tokenHash)
	return args.Get(0).(model.EmailVerificationToken), args.Error(1)
}

func (m *MockEmailRepo) MarkEmailAsVerified(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockEmailRepo) InvalidateOldEmailTokens(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockEmailRepo) GetPasswordResetTokenByHash(tokenHash string) (model.PasswordResetToken, error) {
	args := m.Called(tokenHash)
	return args.Get(0).(model.PasswordResetToken), args.Error(1)
}

func (m *MockEmailRepo) CreateAuthLog(al model.AuthLog) error {
	args := m.Called(al)
	return args.Error(0)
}

func (m *MockEmailRepo) SavePasswordResetTokenWithLog(prt model.PasswordResetToken, al model.AuthLog) error {
	args := m.Called(prt, al)
	return args.Error(0)
}

func (m *MockEmailRepo) CompletePasswordReset(userID uuid.UUID, passwordHash string, al model.AuthLog) error {
	args := m.Called(userID, passwordHash, al)
	return args.Error(0)
}

func (m *MockEmailRepo) CreateOAuthUser(u model.User, al model.AuthLog, ap model.AuthProvider, state string) error {
	args := m.Called(u, al, ap, state)
	return args.Error(0)
}

func (m *MockEmailRepo) GetOAuthProvider(userID uuid.UUID, provider string) error {
	args := m.Called(userID, provider)
	return args.Error(0)
}

func (m *MockEmailRepo) AddOAuthProviderToUser(userID uuid.UUID, ap model.AuthProvider, al model.AuthLog, state string, provider string) error {
	args := m.Called(userID, ap, al, state, provider)
	return args.Error(0)
}

func (m *MockEmailRepo) SaveOAuthState(state, provider string) error {
	args := m.Called(state, provider)
	return args.Error(0)
}

func (m *MockEmailRepo) ValidateOAuthState(state, provider string) error {
	args := m.Called(state, provider)
	return args.Error(0)
}

func (m *MockEmailRepo) ConsumeOAuthStateAndLog(state, provider string, al model.AuthLog) error {
	args := m.Called(state, provider, al)
	return args.Error(0)
}

func (m *MockEmailRepo) GetOAuthProviderByState(state string) (string, error) {
	args := m.Called(state)
	return args.String(0), args.Error(1)
}

func (m *MockEmailRepo) CreateSession(session model.Session) error {
	args := m.Called(session)
	return args.Error(0)
}

func (m *MockEmailRepo) GetSessionByHash(hash string) (model.Session, error) {
	args := m.Called(hash)
	return args.Get(0).(model.Session), args.Error(1)
}

func (m *MockEmailRepo) RevokeSession(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockEmailRepo) RevokeAllUserSessions(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockEmailRepo) GetUserPermissions(userID uuid.UUID) ([]string, error) {
	args := m.Called(userID)
	if args.Get(0) != nil {
		return args.Get(0).([]string), args.Error(1)
	}
	return nil, args.Error(1)
}

// Mock for emailURepo
type MockEmailURepo struct {
	mock.Mock
}

func (m *MockEmailURepo) FindByEmail(email string) (model.User, error) {
	args := m.Called(email)
	return args.Get(0).(model.User), args.Error(1)
}

// Mock for MailService
type MockMailService struct {
	mock.Mock
}

func (m *MockMailService) SendVerificationEmail(ctx context.Context, to, name, token string) error {
	args := m.Called(ctx, to, name, token)
	return args.Error(0)
}

func (m *MockMailService) SendPasswordResetEmail(ctx context.Context, to, name, resetLink string) error {
	args := m.Called(ctx, to, name, resetLink)
	return args.Error(0)
}

func (m *MockMailService) SendPasswordReset(ctx context.Context, to string, name string, resetLink string) error {
	args := m.Called(ctx, to, name, resetLink)
	return args.Error(0)
}

func (m *MockMailService) SendWelcome(ctx context.Context, to string, name string) error {
	args := m.Called(ctx, to, name)
	return args.Error(0)
}

func TestVerifyEmail_NoToken(t *testing.T) {
	repo := new(MockEmailRepo)
	uRepo := new(MockEmailURepo)
	mailService := new(MockMailService)

	s := NewEmailService(repo, uRepo, mailService)

	err := s.VerifyEmail("")
	assert.NotNil(t, err)
	assert.Equal(t, 400, err.Status)
}

func TestResendVerificationEmail_NoEmail(t *testing.T) {
	repo := new(MockEmailRepo)
	uRepo := new(MockEmailURepo)
	mailService := new(MockMailService)

	s := NewEmailService(repo, uRepo, mailService)

	err := s.ResendVerificationEmail("")
	assert.NotNil(t, err)
	assert.Equal(t, 400, err.Status)
}

func TestVerifyEmail_Success(t *testing.T) {
	// Arrange
	repo := new(MockEmailRepo)
	uRepo := new(MockEmailURepo)
	mailService := new(MockMailService)
	s := NewEmailService(repo, uRepo, mailService)

	userID := uuid.New()
	validToken := model.EmailVerificationToken{
		ID:        uuid.New(),
		TokenHash: "hashed_valid_token",
		UserID:    userID,
		CreatedAt: time.Now().Add(-1 * time.Hour), // Within 24 hours
	}

	repo.On("GetEmailVerificationTokenByHash", mock.Anything).Return(validToken, nil)
	repo.On("MarkEmailAsVerified", userID).Return(nil)
	repo.On("InvalidateOldEmailTokens", userID).Return(nil)

	// Act
	err := s.VerifyEmail("valid_token")

	// Assert
	assert.Nil(t, err)
	repo.AssertExpectations(t)
}

func TestVerifyEmail_ReturnsError_WhenTokenExpired(t *testing.T) {
	// Arrange
	repo := new(MockEmailRepo)
	uRepo := new(MockEmailURepo)
	mailService := new(MockMailService)
	s := NewEmailService(repo, uRepo, mailService)

	userID := uuid.New()
	expiredToken := model.EmailVerificationToken{
		ID:        uuid.New(),
		TokenHash: "hashed_expired_token",
		UserID:    userID,
		CreatedAt: time.Now().Add(-48 * time.Hour), // Older than 24 hours
	}

	repo.On("GetEmailVerificationTokenByHash", mock.Anything).Return(expiredToken, nil)

	// Act
	err := s.VerifyEmail("expired_token")

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, 401, err.Status)
	assert.Contains(t, err.Message, "expired")
}

func TestResendVerificationEmail_SendsNewToken(t *testing.T) {
	// Arrange
	repo := new(MockEmailRepo)
	uRepo := new(MockEmailURepo)
	mailService := new(MockMailService)
	s := NewEmailService(repo, uRepo, mailService)

	email := "test@example.com"
	userID := uuid.New()
	unverifiedUser := model.User{
		ID:            userID,
		Email:         email,
		Name:          "Test User",
		EmailVerified: false,
	}

	uRepo.On("FindByEmail", email).Return(unverifiedUser, nil)
	repo.On("InvalidateOldEmailTokens", userID).Return(nil)
	repo.On("SaveEmailVerificationToken", mock.Anything).Return(nil)
	mailService.On("SendVerificationEmail", mock.Anything, email, "Test User", mock.Anything).Return(nil)

	// Act
	err := s.ResendVerificationEmail(email)

	// Assert
	assert.Nil(t, err)
	uRepo.AssertExpectations(t)
	repo.AssertExpectations(t)
	mailService.AssertExpectations(t)
}

func TestResendVerificationEmail_ReturnsError_WhenAlreadyVerified(t *testing.T) {
	// Arrange
	repo := new(MockEmailRepo)
	uRepo := new(MockEmailURepo)
	mailService := new(MockMailService)
	s := NewEmailService(repo, uRepo, mailService)

	email := "alreadyverified@example.com"
	userID := uuid.New()
	alreadyVerifiedUser := model.User{
		ID:            userID,
		Email:         email,
		Name:          "Test User",
		EmailVerified: true, // Already verified
	}

	uRepo.On("FindByEmail", email).Return(alreadyVerifiedUser, nil)

	// Act
	err := s.ResendVerificationEmail(email)

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, 400, err.Status)
	assert.Contains(t, err.Message, "already verified")
}
