package mail

import (
	"fmt"
	"os"
)

// NewEmailService crea una instancia de EmailService basada en la variable de entorno EMAIL_PROVIDER
// Agnóstico: permite cambiar de proveedor solo cambiando la variable de entorno
// Proveedores soportados: "mailpit", "resend"
// Default: "mailpit" (seguro para desarrollo)
func NewEmailService() (EmailService, error) {
	provider := os.Getenv("EMAIL_PROVIDER")
	if provider == "" {
		provider = "mailpit" // Default para desarrollo
	}

	config := EmailConfig{
		FromEmail: os.Getenv("EMAIL_FROM"),
		FromName:  os.Getenv("EMAIL_FROM_NAME"),
		AppURL:    os.Getenv("APP_URL"),
	}

	// Validar que config tenga valores
	if config.FromEmail == "" || config.FromName == "" || config.AppURL == "" {
		return nil, fmt.Errorf("email service: missing required environment variables (EMAIL_FROM, EMAIL_FROM_NAME, APP_URL)")
	}

	switch provider {
	case "mailpit":
		return NewMailpitService(config), nil

	case "resend":
		return NewResendService(config)

	default:
		return nil, fmt.Errorf("email service: unknown provider '%s', must be 'mailpit' or 'resend'", provider)
	}
}
