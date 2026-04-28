package mail

import (
	"context"
	"fmt"
	"os"

	"github.com/resend/resend-go/v3"
)

// ResendService implementa EmailService usando Resend API HTTP
// Para producción con emails reales
type ResendService struct {
	config EmailConfig
	client *resend.Client
}

// NewResendService crea una nueva instancia de Resend service
// Lee la API key desde la variable de entorno RESEND_API_KEY
func NewResendService(config EmailConfig) (*ResendService, error) {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("resend: RESEND_API_KEY environment variable is required")
	}

	return &ResendService{
		config: config,
		client: resend.NewClient(apiKey),
	}, nil
}

// SendVerificationEmail envía email de verificación a través de Resend API
func (r *ResendService) SendVerificationEmail(ctx context.Context, to, name, token string) error {
	verifyLink := fmt.Sprintf("%s/api/v1/auth/verify-email/%s", r.config.AppURL, token)

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
	<div style="background: #f5f5f5; padding: 20px; border-radius: 8px;">
		<h2 style="color: #333; margin-bottom: 16px;">Hola %s,</h2>
		<p style="color: #666; line-height: 1.6; margin-bottom: 20px;">
			¡Bienvenido a %s! Para completar tu registro, verifica tu email.
		</p>
		<div style="text-align: center; margin: 30px 0;">
			<a href="%s" style="background: #007bff; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block; font-weight: bold;">
				Verificar Email
			</a>
		</div>
		<p style="color: #999; font-size: 12px; text-align: center;">
			Si no solicitaste este email, puedes ignorarlo.
		</p>
		<p style="color: #999; font-size: 11px; text-align: center;">
			Este enlace expira en 24 horas.
		</p>
	</div>
</body>
</html>
	`, name, r.config.FromName, verifyLink)

	return r.sendEmail(to, "Verifica tu email - "+r.config.FromName, htmlBody)
}

// SendPasswordReset envía email de reset de contraseña a través de Resend API
func (r *ResendService) SendPasswordReset(ctx context.Context, to, name, resetLink string) error {
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
	<div style="background: #f5f5f5; padding: 20px; border-radius: 8px;">
		<h2 style="color: #333; margin-bottom: 16px;">Hola %s,</h2>
		<p style="color: #666; line-height: 1.6; margin-bottom: 20px;">
			Recibimos una solicitud para resetear tu contraseña.
		</p>
		<div style="text-align: center; margin: 30px 0;">
			<a href="%s" style="background: #dc3545; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block; font-weight: bold;">
				Resetear Contraseña
			</a>
		</div>
		<p style="color: #999; font-size: 12px; text-align: center;">
			Si no solicitaste este email, puedes ignorarlo y tu contraseña seguirá siendo segura.
		</p>
		<p style="color: #999; font-size: 11px; text-align: center;">
			Este enlace expira en 1 hora.
		</p>
	</div>
</body>
</html>
	`, name, resetLink)

	return r.sendEmail(to, "Reset tu contraseña - "+r.config.FromName, htmlBody)
}

// SendWelcome envía email de bienvenida a través de Resend API
func (r *ResendService) SendWelcome(ctx context.Context, to, name string) error {
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
	<div style="background: #f5f5f5; padding: 20px; border-radius: 8px;">
		<h2 style="color: #333; margin-bottom: 16px;">¡Bienvenido %s!</h2>
		<p style="color: #666; line-height: 1.6; margin-bottom: 20px;">
			Tu cuenta ha sido creada exitosamente. Estamos emocionados de tenerte aquí.
		</p>
		<div style="text-align: center; margin: 30px 0;">
			<a href="%s" style="background: #28a745; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block; font-weight: bold;">
				Ir a Mi Cuenta
			</a>
		</div>
		<p style="color: #999; font-size: 12px; text-align: center;">
			Si tienes preguntas, contáctanos en support@example.com
		</p>
	</div>
</body>
</html>
	`, name, r.config.AppURL)

	return r.sendEmail(to, "Bienvenido a "+r.config.FromName, htmlBody)
}

// sendEmail envía un email genérico a través de Resend API (método privado)
func (r *ResendService) sendEmail(to, subject, htmlBody string) error {
	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", r.config.FromName, r.config.FromEmail),
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
	}

	sent, err := r.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("resend: failed to send email to %s: %w", to, err)
	}

	if sent.Id == "" {
		return fmt.Errorf("resend: empty response ID from API")
	}

	return nil
}
