package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/MetaDandy/go-fiber-skeleton/src/enum"
	"golang.org/x/oauth2"
)

// UserInfo contiene la información genérica del usuario
type UserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

// GetUserInfo obtiene la información del usuario basado en el proveedor y token
func GetUserInfo(ctx context.Context, provider string, token *oauth2.Token) (*UserInfo, error) {
	switch provider {
	case enum.Google.String():
		return getGoogleUserInfo(ctx, token)
	case enum.Github.String():
		return getGitHubUserInfo(ctx, token)
	case enum.Discord.String():
		return getDiscordUserInfo(ctx, token)
	default:
		return nil, fmt.Errorf("provider '%s' not supported", provider)
	}
}

// getGoogleUserInfo obtiene datos del usuario de Google
func getGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	resp, err := fetchUserData(ctx, token, "https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, err
	}

	var data struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.Unmarshal(resp, &data); err != nil {
		return nil, fmt.Errorf("failed to parse google response: %w", err)
	}

	return &UserInfo{
		ID:    data.ID,
		Email: data.Email,
		Name:  data.Name,
		Image: data.Picture,
	}, nil
}

// getGitHubUserInfo obtiene datos del usuario de GitHub
func getGitHubUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	resp, err := fetchUserData(ctx, token, "https://api.github.com/user")
	if err != nil {
		return nil, err
	}

	var data struct {
		ID     int    `json:"id"`
		Login  string `json:"login"`
		Name   string `json:"name"`
		Email  string `json:"email"`
		Avatar string `json:"avatar_url"`
	}

	if err := json.Unmarshal(resp, &data); err != nil {
		return nil, fmt.Errorf("failed to parse github response: %w", err)
	}

	// Si no tiene email público, obtenerlo desde /user/emails
	if data.Email == "" {
		emailResp, err := fetchUserData(ctx, token, "https://api.github.com/user/emails")
		if err == nil {
			var emails []struct {
				Email   string `json:"email"`
				Primary bool   `json:"primary"`
			}
			if err := json.Unmarshal(emailResp, &emails); err == nil {
				// Buscar el email primario
				for _, e := range emails {
					if e.Primary {
						data.Email = e.Email
						break
					}
				}
				// Si no hay primario, tomar el primero
				if data.Email == "" && len(emails) > 0 {
					data.Email = emails[0].Email
				}
			}
		}
	}

	return &UserInfo{
		ID:    fmt.Sprintf("%d", data.ID),
		Email: data.Email,
		Name:  data.Name,
		Image: data.Avatar,
	}, nil
}

// getDiscordUserInfo obtiene datos del usuario de Discord
func getDiscordUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	resp, err := fetchUserData(ctx, token, "https://discord.com/api/users/@me")
	if err != nil {
		return nil, err
	}

	var data struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		Username string `json:"username"`
		Avatar   string `json:"avatar"`
	}

	if err := json.Unmarshal(resp, &data); err != nil {
		return nil, fmt.Errorf("failed to parse discord response: %w", err)
	}

	return &UserInfo{
		ID:    data.ID,
		Email: data.Email,
		Name:  data.Username,
		Image: fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", data.ID, data.Avatar),
	}, nil
}

// fetchUserData hace una petición GET autenticada con el token
func fetchUserData(ctx context.Context, token *oauth2.Token, url string) ([]byte, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}
