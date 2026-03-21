package mail

// EmailConfig es la configuración compartida para todos los proveedores
type EmailConfig struct {
	FromEmail string // ej: noreply@myapp.com
	FromName  string // ej: My App
	AppURL    string // ej: http://localhost:3000 o https://myapp.com
}
