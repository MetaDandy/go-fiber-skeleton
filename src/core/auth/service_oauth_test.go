package authentication

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockOAuthRepo is a mock for the Repo interface in oauth methods
type mockOAuthRepo struct {
	mock.Mock
	Repo // Embedded to satisfy the interface for methods we don't mock
}

func (m *mockOAuthRepo) SaveOAuthState(state, provider string) error {
	args := m.Called(state, provider)
	return args.Error(0)
}

func (m *mockOAuthRepo) GetOAuthProviderByState(state string) (string, error) {
	args := m.Called(state)
	return args.String(0), args.Error(1)
}

func TestOAuthLogin_InvalidProvider(t *testing.T) {
	mockRepo := new(mockOAuthRepo)
	service := NewOAuthService(mockRepo, nil, "http://localhost/callback")

	url, err := service.OAuthLogin("invalid_provider")

	assert.Empty(t, url)
	assert.NotNil(t, err)
	assert.Equal(t, 400, err.Status)
	assert.Equal(t, "Unsupported oauth provider", err.Message)
}

func TestOAuthCallback_InvalidState(t *testing.T) {
	mockRepo := new(mockOAuthRepo)
	mockRepo.On("GetOAuthProviderByState", "invalid_state").Return("", assert.AnError)
	service := NewOAuthService(mockRepo, nil, "http://localhost/callback")

	token, refresh, err := service.OAuthCallback("code", "invalid_state", "127.0.0.1", "agent")

	assert.Empty(t, token)
	assert.Empty(t, refresh)
	assert.NotNil(t, err)
	assert.Equal(t, 400, err.Status)
	assert.Equal(t, "Invalid or expired OAuth state", err.Message)
	mockRepo.AssertExpectations(t)
}
