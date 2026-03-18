package auth

import (
	"context"
	"fmt"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// Credentials contiene las credenciales de un proveedor desde .env
type Credentials struct {
	ClientID     string
	ClientSecret string
}

// getProviderConfig retorna la configuración OAuth2 según el proveedor
func getProviderConfig(provider string, creds Credentials, redirectURL string) *oauth2.Config {
	switch strings.ToLower(provider) {
	case "google":
		return &oauth2.Config{
			ClientID:     creds.ClientID,
			ClientSecret: creds.ClientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		}

	case "github":
		return &oauth2.Config{
			ClientID:     creds.ClientID,
			ClientSecret: creds.ClientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				"user:email",
			},
			Endpoint: github.Endpoint,
		}

	case "discord":
		return &oauth2.Config{
			ClientID:     creds.ClientID,
			ClientSecret: creds.ClientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				"identify",
				"email",
			},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://discord.com/api/oauth2/authorize",
				TokenURL: "https://discord.com/api/oauth2/token",
			},
		}

	default:
		return nil
	}
}

// GetAuthURL genera la URL de login del proveedor OAuth
func GetAuthURL(provider string, creds Credentials, redirectURL string, state string) (string, error) {
	config := getProviderConfig(provider, creds, redirectURL)
	if config == nil {
		return "", fmt.Errorf("provider '%s' not supported", provider)
	}

	return config.AuthCodeURL(state), nil
}

// ExchangeCode intercambia el código de autorización por un token
func ExchangeCode(provider string, creds Credentials, redirectURL string, code string) (*oauth2.Token, error) {
	config := getProviderConfig(provider, creds, redirectURL)
	if config == nil {
		return nil, fmt.Errorf("provider '%s' not supported", provider)
	}

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	return token, nil
}

// LoadCredentials carga las credenciales de un proveedor desde variables de entorno
func LoadCredentials(provider string) (Credentials, error) {
	clientIDKey := fmt.Sprintf("%s_CLIENT_ID", strings.ToUpper(provider))
	clientSecretKey := fmt.Sprintf("%s_CLIENT_SECRET", strings.ToUpper(provider))

	clientID := os.Getenv(clientIDKey)
	clientSecret := os.Getenv(clientSecretKey)

	if clientID == "" || clientSecret == "" {
		return Credentials{}, fmt.Errorf("missing credentials for %s. set %s and %s in .env", provider, clientIDKey, clientSecretKey)
	}

	return Credentials{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}, nil
}
