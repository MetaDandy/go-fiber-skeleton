package authentication

import (
	"testing"

	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSessionRepo struct {
	mock.Mock
	Repo // Embedded to satisfy the interface for methods we don't mock
}

func (m *mockSessionRepo) GetSessionByHash(hash string) (model.Session, error) {
	args := m.Called(hash)
	return args.Get(0).(model.Session), args.Error(1)
}

func (m *mockSessionRepo) RevokeSession(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockSessionRepo) GetUserPermissions(userID uuid.UUID) ([]string, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockSessionRepo) CreateSession(session model.Session) error {
	args := m.Called(session)
	return args.Error(0)
}

type mockSessionURepo struct {
	mock.Mock
}

func (m *mockSessionURepo) FindByID(id string) (model.User, error) {
	args := m.Called(id)
	return args.Get(0).(model.User), args.Error(1)
}

func TestRefreshToken_InvalidSession(t *testing.T) {
	// Arrange
	mockRepo := new(mockSessionRepo)
	mockURepo := new(mockSessionURepo)
	service := NewSessionService(mockRepo, mockURepo)

	mockRepo.On("GetSessionByHash", mock.Anything).Return(model.Session{}, assert.AnError)

	// Act
	_, _, err := service.RefreshToken("fake-token", "127.0.0.1", "user-agent")

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, 401, err.Status)
	assert.Equal(t, "Invalid or expired session", err.Message)
	mockRepo.AssertExpectations(t)
}

func TestLogout_ValidSession(t *testing.T) {
	// Arrange
	mockRepo := new(mockSessionRepo)
	mockURepo := new(mockSessionURepo)
	service := NewSessionService(mockRepo, mockURepo)

	sessionID := uuid.New()
	mockRepo.On("GetSessionByHash", mock.Anything).Return(model.Session{ID: sessionID}, nil)
	mockRepo.On("RevokeSession", sessionID).Return(nil)

	// Act
	err := service.Logout("fake-token")

	// Assert
	assert.Nil(t, err)
	mockRepo.AssertExpectations(t)
}
