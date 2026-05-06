package authentication

import (
	"context"
	"testing"

	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock for Repo
type MockEmailRepo struct {
	mock.Mock
}

func (m *MockEmailRepo) GetEmailVerificationTokenByHash(hash string) (*model.EmailVerificationToken, error) {
	args := m.Called(hash)
	if args.Get(0) != nil {
		return args.Get(0).(*model.EmailVerificationToken), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockEmailRepo) MarkEmailAsVerified(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockEmailRepo) InvalidateOldEmailTokens(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockEmailRepo) SaveEmailVerificationToken(token model.EmailVerificationToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockEmailRepo) CreateUser(user model.User) error { return m.Called(user).Error(0) }
func (m *MockEmailRepo) GetUserByEmail(email string) (model.User, error) { 
	args := m.Called(email)
	return args.Get(0).(model.User), args.Error(1)
}
func (m *MockEmailRepo) FindByEmail(email string) (model.User, error) {
	args := m.Called(email)
	return args.Get(0).(model.User), args.Error(1)
}
func (m *MockEmailRepo) SaveSession(session model.Session) error { return m.Called(session).Error(0) }
func (m *MockEmailRepo) GetSessionByID(id string) (model.Session, error) { 
	args := m.Called(id)
	return args.Get(0).(model.Session), args.Error(1)
}
func (m *MockEmailRepo) GetSessionByHash(hash string) (model.Session, error) {
	args := m.Called(hash)
	return args.Get(0).(model.Session), args.Error(1)
}
func (m *MockEmailRepo) DeleteSession(id string) error { return m.Called(id).Error(0) }
func (m *MockEmailRepo) RevokeSession(id uuid.UUID) error { return m.Called(id).Error(0) }
func (m *MockEmailRepo) RevokeAllUserSessions(userID uuid.UUID) error { return m.Called(userID).Error(0) }

func (m *MockEmailRepo) SavePasswordResetToken(token model.PasswordResetToken) error { return m.Called(token).Error(0) }
func (m *MockEmailRepo) GetPasswordResetTokenByHash(hash string) (*model.PasswordResetToken, error) { 
	args := m.Called(hash)
	if args.Get(0) != nil {
		return args.Get(0).(*model.PasswordResetToken), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockEmailRepo) InvalidateOldResetTokens(userID uuid.UUID) error { return m.Called(userID).Error(0) }
func (m *MockEmailRepo) UpdatePassword(userID uuid.UUID, newPassword string) error { return m.Called(userID, newPassword).Error(0) }
func (m *MockEmailRepo) CompletePasswordReset(userID uuid.UUID, newPasswordHash string, log model.AuthLog) error { return m.Called(userID, newPasswordHash, log).Error(0) }

func (m *MockEmailRepo) GetOAuthState(state string) (*model.OAuthState, error) { 
	args := m.Called(state)
	if args.Get(0) != nil {
		return args.Get(0).(*model.OAuthState), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockEmailRepo) SaveOAuthState(state model.OAuthState) error { return m.Called(state).Error(0) }
func (m *MockEmailRepo) DeleteOAuthState(state string) error { return m.Called(state).Error(0) }
func (m *MockEmailRepo) AddOAuthProviderToUser(userID uuid.UUID, provider model.AuthProvider, log model.AuthLog, email, ip string) error { return m.Called(userID, provider, log, email, ip).Error(0) }
func (m *MockEmailRepo) GetUserAuthProviders(userID uuid.UUID) ([]string, error) {
	args := m.Called(userID)
	if args.Get(0) != nil {
		return args.Get(0).([]string), args.Error(1)
	}
	return nil, args.Error(1)
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
