package authentication

import (
	"testing"
	"time"

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

func TestRefreshToken_ReturnsNewTokens_WhenValidSession(t *testing.T) {
	// Arrange
	mockRepo := new(mockSessionRepo)
	mockURepo := new(mockSessionURepo)
	service := NewSessionService(mockRepo, mockURepo)

	sessionID := uuid.New()
	userID := uuid.New()
	roleID := uuid.New()

	// Set up a valid session that hasn't expired
	validSession := model.Session{
		ID:        sessionID,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		UserID:    userID,
	}

	testUser := model.User{
		ID:       userID,
		Email:    "test@example.com",
		RoleID:   roleID,
	}

	mockRepo.On("GetSessionByHash", mock.Anything).Return(validSession, nil)
	mockURepo.On("FindByID", userID.String()).Return(testUser, nil)
	mockRepo.On("GetUserPermissions", userID).Return([]string{"read", "write"}, nil)
	mockRepo.On("RevokeSession", sessionID).Return(nil)
	mockRepo.On("CreateSession", mock.Anything).Return(nil)

	// Act
	accessToken, refreshToken, err := service.RefreshToken("valid-refresh-token", "127.0.0.1", "user-agent")

	// Assert - verify new tokens are returned
	assert.Nil(t, err)
	assert.NotEmpty(t, accessToken, "access token should be generated")
	assert.NotEmpty(t, refreshToken, "refresh token should be generated")
	assert.NotEqual(t, "valid-refresh-token", refreshToken, "refresh token should be rotated")
	mockRepo.AssertExpectations(t)
	mockURepo.AssertExpectations(t)
}

func TestRefreshToken_Returns401_WhenSessionExpired(t *testing.T) {
	// Arrange
	mockRepo := new(mockSessionRepo)
	mockURepo := new(mockSessionURepo)
	service := NewSessionService(mockRepo, mockURepo)

	sessionID := uuid.New()
	userID := uuid.New()

	// Set up an EXPIRED session
	expiredSession := model.Session{
		ID:        sessionID,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		UserID:    userID,
	}

	mockRepo.On("GetSessionByHash", mock.Anything).Return(expiredSession, nil)
	mockRepo.On("RevokeSession", sessionID).Return(nil) // Should revoke the expired session

	// Act
	accessToken, refreshToken, err := service.RefreshToken("expired-refresh-token", "127.0.0.1", "user-agent")

	// Assert - verify 401 error is returned
	assert.NotNil(t, err)
	assert.Equal(t, 401, err.Status)
	assert.Equal(t, "Session expired", err.Message)
	assert.Empty(t, accessToken, "no access token should be returned")
	assert.Empty(t, refreshToken, "no refresh token should be returned")
	mockRepo.AssertExpectations(t)
}

func TestRefreshToken_Returns401_WhenUserNotFound(t *testing.T) {
	// Arrange
	mockRepo := new(mockSessionRepo)
	mockURepo := new(mockSessionURepo)
	service := NewSessionService(mockRepo, mockURepo)

	sessionID := uuid.New()
	userID := uuid.New()

	// Set up a valid (non-expired) session but user no longer exists
	validSession := model.Session{
		ID:        sessionID,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		UserID:    userID,
	}

	// User not found in database
	mockRepo.On("GetSessionByHash", mock.Anything).Return(validSession, nil)
	mockURepo.On("FindByID", userID.String()).Return(model.User{}, assert.AnError)

	// Act
	accessToken, refreshToken, err := service.RefreshToken("valid-refresh-token", "127.0.0.1", "user-agent")

	// Assert - verify 401 error is returned
	assert.NotNil(t, err)
	assert.Equal(t, 401, err.Status)
	assert.Equal(t, "User not found", err.Message)
	assert.Empty(t, accessToken, "no access token should be returned")
	assert.Empty(t, refreshToken, "no refresh token should be returned")
	mockRepo.AssertExpectations(t)
	mockURepo.AssertExpectations(t)
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

func TestLogout_ReturnsNil_WhenEmptyToken(t *testing.T) {
	// Arrange
	mockRepo := new(mockSessionRepo)
	mockURepo := new(mockSessionURepo)
	service := NewSessionService(mockRepo, mockURepo)

	// Act - empty token should return nil without calling any repo methods
	err := service.Logout("")

	// Assert - verify nil is returned (no error)
	assert.Nil(t, err, "empty token should return nil error")
	// No repo methods should be called for empty token
	mockRepo.AssertNotCalled(t, "GetSessionByHash")
	mockRepo.AssertNotCalled(t, "RevokeSession")
}
