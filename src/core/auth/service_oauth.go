package authentication

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/MetaDandy/go-fiber-skeleton/constant"
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/enum"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/MetaDandy/go-fiber-skeleton/src/service/auth"
	"github.com/MetaDandy/go-fiber-skeleton/src/service/mail"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// oauthURepo defines user repository methods needed by oauthService
type oauthURepo interface {
	FindByEmail(email string) (model.User, error)
}

// OAuthService interface
type OAuthService interface {
	OAuthLogin(provider string) (string, *api_error.Error)
	OAuthCallback(code, state, ip, userAgent string) (string, string, *api_error.Error)
}

type oauthService struct {
	repo        Repo
	uRepo       oauthURepo
	redirectURL string
}

func NewOAuthService(repo Repo, uRepo oauthURepo, redirectURL string) OAuthService {
	return &oauthService{
		repo:        repo,
		uRepo:       uRepo,
		redirectURL: redirectURL,
	}
}

// OAuthLogin implements the OAuth login flow
func (s *oauthService) OAuthLogin(provider string) (string, *api_error.Error) {
	if !enum.IsValidAuthProvider(provider) {
		return "", api_error.BadRequest("Unsupported oauth provider")
	}
	creds, err := auth.LoadCredentials(provider)
	if err != nil {
		return "", api_error.InternalServerError("Could not load provider credentials").WithErr(err)
	}
	state, err := generateState()
	if err != nil {
		return "", api_error.InternalServerError("Failed to generate OAuth state").WithErr(err)
	}
	if err := s.repo.SaveOAuthState(state, provider); err != nil {
		return "", api_error.InternalServerError("Failed to save OAuth state").WithErr(err)
	}
	authURL, err := auth.GetAuthURL(provider, creds, s.redirectURL, state)
	if err != nil {
		return "", api_error.InternalServerError("Could not generate auth URL").WithErr(err)
	}
	return authURL, nil
}

// OAuthCallback handles the OAuth callback
func (s *oauthService) OAuthCallback(code, state, ip, userAgent string) (string, string, *api_error.Error) {
	// Validate state and get provider
	provider, err := s.repo.GetOAuthProviderByState(state)
	if err != nil {
		return "", "", api_error.BadRequest("Invalid or expired OAuth state")
	}
	creds, err := auth.LoadCredentials(provider)
	if err != nil {
		return "", "", api_error.InternalServerError("Could not load provider credentials").WithErr(err)
	}
	token, err := auth.ExchangeCode(provider, creds, s.redirectURL, code)
	if err != nil {
		return "", "", api_error.InternalServerError("Failed to exchange authorization code").WithErr(err)
	}
	userInfo, err := auth.GetUserInfo(context.Background(), provider, token)
	if err != nil {
		return "", "", api_error.InternalServerError("Failed to retrieve user information").WithErr(err)
	}
	// Build internal input
	input := oAuthCallbackInternal{
		Provider: provider,
		State:    state,
		UserInfo: OAuthUserInfo{
			ID:    userInfo.ID,
			Email: userInfo.Email,
			Name:  userInfo.Name,
			Image: userInfo.Image,
		},
		Ip:        ip,
		UserAgent: userAgent,
	}
	// Call internal logic
	return s.oauthCreateOrLogin(input)
}

// oAuthCallbackInternal is the internal DTO after processing the callback
type oAuthCallbackInternal struct {
	Provider  string
	State     string
	UserInfo  OAuthUserInfo
	Ip        string
	UserAgent string
}

// oauthCreateOrLogin handles the OAuth flow: creates user if not exists or logs in if exists
func (s *oauthService) oauthCreateOrLogin(input oAuthCallbackInternal) (string, string, *api_error.Error) {
	if input.UserInfo.Email == "" {
		return "", "", api_error.BadRequest("No email provided by OAuth provider")
	}
	if !enum.IsValidAuthProvider(input.Provider) {
		return "", "", api_error.BadRequest("Invalid OAuth provider")
	}
	// Buscar si el usuario existe por email
	user, err := s.uRepo.FindByEmail(input.UserInfo.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		return "", "", api_error.InternalServerError("Could not verify user").WithErr(err)
	}
	// CASO A: Email NO existe - SIGNUP con OAuth
	if err == gorm.ErrRecordNotFound {
		return s.oauthSignUp(input)
	}
	// CASO B: Email EXISTS - LOGIN con OAuth
	return s.oauthLogin(input, user)
}

// oauthSignUp creates a new user with OAuth
func (s *oauthService) oauthSignUp(input oAuthCallbackInternal) (string, string, *api_error.Error) {
	user := model.User{
		ID:              uuid.New(),
		Email:           input.UserInfo.Email,
		Name:            input.UserInfo.Name,
		Picture:         input.UserInfo.Image,
		Password:        nil,
		EmailVerified:   true,
		EmailVerifiedAt: ptrTime(time.Now()),
		RoleID:          constant.GenericID,
	}
	authLog := model.AuthLog{
		ID:        uuid.New(),
		Event:     enum.OAuthLogin,
		UserID:    user.ID,
		Ip:        input.Ip,
		UserAgent: input.UserAgent,
	}
	authProvider := model.AuthProvider{
		ID:             uuid.New(),
		Provider:       input.Provider,
		ProviderUserID: input.UserInfo.ID,
		UserID:         user.ID,
	}
	// Transacción: crear user + authlog + authprovider + consumir state
	if err := s.repo.CreateOAuthUser(user, authLog, authProvider, input.State); err != nil {
		log.Printf("failed to create OAuth user: %v", err)
		return "", "", api_error.InternalServerError("Failed to create user").WithErr(err)
	}
	return s.generateAndSaveSession(user, input.Provider, input.Ip, input.UserAgent)
}

// oauthLogin handles login of an existing user with OAuth
func (s *oauthService) oauthLogin(input oAuthCallbackInternal, user model.User) (string, string, *api_error.Error) {
	// Verificar si el usuario ya tiene este proveedor
	err := s.repo.GetOAuthProvider(user.ID, input.Provider)
	if err != nil && err != gorm.ErrRecordNotFound {
		return "", "", api_error.InternalServerError("Could not verify auth provider").WithErr(err)
	}
	// CASO B1: El proveedor ya existe
	if err == nil {
		authLog := model.AuthLog{
			ID:        uuid.New(),
			Event:     enum.OAuthLogin,
			UserID:    user.ID,
			Ip:        input.Ip,
			UserAgent: input.UserAgent,
		}
		// Transacción atómica: consumir state + crear log
		if err := s.repo.ConsumeOAuthStateAndLog(input.State, input.Provider, authLog); err != nil {
			log.Printf("failed to consume OAuth state and create log: %v", err)
		}
	} else {
		// CASO B2: El proveedor NO existe - Agregar nuevo proveedor
		authProvider := model.AuthProvider{
			ID:             uuid.New(),
			Provider:       input.Provider,
			ProviderUserID: input.UserInfo.ID,
			UserID:         user.ID,
		}
		authLog := model.AuthLog{
			ID:        uuid.New(),
			Event:     enum.OAuthLogin,
			UserID:    user.ID,
			Ip:        input.Ip,
			UserAgent: input.UserAgent,
		}
		if err := s.repo.AddOAuthProviderToUser(user.ID, authProvider, authLog, input.State, input.Provider); err != nil {
			log.Printf("failed to add OAuth provider to user: %v", err)
			return "", "", api_error.InternalServerError("Failed to authenticate").WithErr(err)
		}
	}
	return s.generateAndSaveSession(user, input.Provider, input.Ip, input.UserAgent)
}

// generateAndSaveSession creates tokens and session
func (s *oauthService) generateAndSaveSession(user model.User, provider, ip, userAgent string) (string, string, *api_error.Error) {
	permissions, err := s.repo.GetUserPermissions(user.ID)
	if err != nil {
		log.Printf("failed to get user permissions: %v", err)
		return "", "", api_error.InternalServerError("Failed to get user permissions").WithErr(err)
	}
	var roleIDStr string
	if user.RoleID != uuid.Nil {
		roleIDStr = user.RoleID.String()
	}
	accessToken, err := helper.GenerateJwt(user.ID.String(), user.Email, roleIDStr, permissions)
	if err != nil {
		log.Printf("failed to generate access token: %v", err)
		return "", "", api_error.InternalServerError("Failed to generate access token").WithErr(err)
	}
	refreshToken, err := helper.GenerateRefreshToken()
	if err != nil {
		log.Printf("failed to generate refresh token: %v", err)
		return "", "", api_error.InternalServerError("Failed to generate refresh token").WithErr(err)
	}
	session := model.Session{
		ID:               uuid.New(),
		Provider:         provider,
		RefreshTokenHash: mail.HashToken(refreshToken),
		ExpiresAt:        time.Now().Add(7 * 24 * time.Hour),
		Ip:               ip,
		UserAgent:        userAgent,
		UserID:           user.ID,
	}
	// Revocar sesiones anteriores y guardar la nueva
	_ = s.repo.RevokeAllUserSessions(user.ID)
	if err := s.repo.CreateSession(session); err != nil {
		log.Printf("failed to create session: %v", err)
		return "", "", api_error.InternalServerError("Failed to register session").WithErr(err)
	}
	return accessToken, refreshToken, nil
}

// generateState generates a random state for CSRF protection
func generateState() (string, error) {
	state := make([]byte, 32)
	if _, err := rand.Read(state); err != nil {
		return "", fmt.Errorf("failed to generate OAuth state: %w", err)
	}
	return string(state), nil
}

// ptrTime creates a pointer to time.Time
func ptrTime(t time.Time) *time.Time {
	return &t
}
