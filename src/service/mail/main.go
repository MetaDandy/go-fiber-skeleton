package mail

import "context"

// EmailService es la interface agnóstica que implementan todos los proveedores
// Permite cambiar entre Mailpit, Resend, SendGrid, etc. sin cambiar el código
type EmailService interface {
	SendVerificationEmail(ctx context.Context, to, name, token string) error
	SendPasswordReset(ctx context.Context, to, name, resetLink string) error
	SendWelcome(ctx context.Context, to, name string) error
}
