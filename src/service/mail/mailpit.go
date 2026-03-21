package mail

import (
	"context"
	"fmt"
	"net/smtp"
	"os"
)

// MailpitService implementa EmailService usando SMTP directo (Mailpit)
// No requiere dependencias externas, usa net/smtp built-in
type MailpitService struct {
	config EmailConfig
	host   string
	port   string
}

// NewMailpitService crea una nueva instancia de Mailpit service
func NewMailpitService(config EmailConfig) *MailpitService {
	host := os.Getenv("MAILPIT_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("MAILPIT_PORT")
	if port == "" {
		port = "1025"
	}

	return &MailpitService{
		config: config,
		host:   host,
		port:   port,
	}
}

// SendVerificationEmail envía email de verificación a través de Mailpit SMTP
func (m *MailpitService) SendVerificationEmail(ctx context.Context, to, name, token string) error {
	verifyLink := fmt.Sprintf("%s/auth/verify-email/%s", m.config.AppURL, token)

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
	`, name, m.config.FromName, verifyLink)

	return m.sendEmail(to, "Verifica tu email - "+m.config.FromName, htmlBody)
}

// SendPasswordReset envía email de reset de contraseña a través de Mailpit SMTP
func (m *MailpitService) SendPasswordReset(ctx context.Context, to, name, resetLink string) error {
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

	return m.sendEmail(to, "Reset tu contraseña - "+m.config.FromName, htmlBody)
}

// SendWelcome envía email de bienvenida a través de Mailpit SMTP
func (m *MailpitService) SendWelcome(ctx context.Context, to, name string) error {
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
	`, name, m.config.AppURL)

	return m.sendEmail(to, "Bienvenido a "+m.config.FromName, htmlBody)
}

// sendEmail envía un email genérico a través de SMTP (método privado)
// Crea headers MIME válidos para compatibilidad con todos los clientes
func (m *MailpitService) sendEmail(to, subject, htmlBody string) error {
	// Headers MIME (formato CRLF es crítico para SMTP)
	headers := fmt.Sprintf(
		"From: %s <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=\"UTF-8\"\r\n"+
			"\r\n",
		m.config.FromName,
		m.config.FromEmail,
		to,
		subject,
	)

	fullMessage := headers + htmlBody

	// Conectar a Mailpit SMTP
	addr := fmt.Sprintf("%s:%s", m.host, m.port)

	// Mailpit no requiere autenticación en desarrollo
	err := smtp.SendMail(
		addr,
		nil, // Sin autenticación
		m.config.FromEmail,
		[]string{to},
		[]byte(fullMessage),
	)

	if err != nil {
		return fmt.Errorf("mailpit: failed to send email to %s: %w", to, err)
	}

	return nil
}
