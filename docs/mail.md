# Mail Service - Guía Completa

## 🎯 Visión General

Este proyecto tiene un servicio de email **agnóstico** que funciona con múltiples proveedores:
- **Mailpit** (desarrollo local) - Gratis, sin dependencias externas
- **Resend** (producción) - API profesional para emails reales

El cambio entre proveedores es **automático** según la variable de entorno `EMAIL_PROVIDER`.

---

## 📁 Estructura del Servicio

```
src/service/mail/
├── main.go          ← Interface EmailService (agnóstica)
├── config.go        ← Configuración común
├── types.go         ← Tipos compartidos
├── token.go         ← Generación de tokens
├── mailpit.go       ← Implementación SMTP (Mailpit)
├── resend.go        ← Implementación API HTTP (Resend)
└── factory.go       ← Factory pattern (elige proveedor)
```

---

## 🚀 Paso 1: Configurar Variables de Entorno (.env)

Abre o crea tu archivo `.env` en la raíz del proyecto:

```bash
# ===== MAIL SERVICE =====
EMAIL_PROVIDER=mailpit              # mailpit o resend
EMAIL_FROM=noreply@myapp.local      # Email del remitente
EMAIL_FROM_NAME=My App              # Nombre del remitente
APP_URL=http://localhost:3000       # URL base de tu app

# ===== MAILPIT CONFIG (solo si usas Mailpit) =====
MAILPIT_HOST=localhost              # Host donde está Mailpit
MAILPIT_PORT=1025                   # Puerto SMTP de Mailpit
MAILPIT_WEB_URL=http://localhost:8025  # Web UI de Mailpit

# ===== RESEND CONFIG (solo si usas Resend) =====
# RESEND_API_KEY=re_tu_api_key_aqui  # Descomenta para producción
```

### Variables Requeridas
- ✅ `EMAIL_PROVIDER` - "mailpit" o "resend"
- ✅ `EMAIL_FROM` - Email válido del remitente
- ✅ `EMAIL_FROM_NAME` - Nombre que verá el usuario
- ✅ `APP_URL` - URL base (para links en emails)

---

## 🐳 Paso 2: Levantar Mailpit con Docker

### Requisito Previo: Permisos Docker

Si ves este error:
```
permission denied while trying to connect to the docker API
```

Ejecuta esto UNA VEZ:
```bash
sudo usermod -aG docker $USER
newgrp docker
```

Luego verifica:
```bash
docker ps
```

### Iniciar Mailpit

```bash
# Levantar en background
docker-compose -f docker-compose.mail.yaml up -d

# Ver si está corriendo
docker-compose -f docker-compose.mail.yaml ps

# Ver logs
docker-compose -f docker-compose.mail.yaml logs -f mailpit

# Detener
docker-compose -f docker-compose.mail.yaml down
```

**Resultado esperado:**
```
NAME            STATUS              PORTS
mailpit         Up 2 seconds        0.0.0.0:1025->1025/tcp, 0.0.0.0:8025->8025/tcp
```

### Puertos

- **1025** - SMTP server (tu app conecta aquí)
- **8025** - Web UI (accedes en navegador)

---

## ✅ Paso 3: Verificar que Mailpit Está Corriendo

### Opción 1: Web UI
```
http://localhost:8025
```

Deberías ver:
- Interfaz clara con "Mailbox" vacío
- Buscador de emails
- Opción de taggear mensajes

### Opción 2: SMTP Connection
```bash
# Verificar que puerto 1025 está abierto
nc -zv localhost 1025
```

Resultado esperado:
```
Connection to localhost 1025 port [tcp/*] succeeded!
```

---

## 🔧 Paso 4: Configurar la App

La app **YA ESTÁ CONFIGURADA** automáticamente, pero aquí te muestro cómo funciona:

### En `src/container.go`
```go
// El container crea el mail service automáticamente
mailService, err := mail.NewEmailService()
if err != nil {
    log.Fatalf("Failed to initialize mail service: %v", err)
}

// Lo inyecta en auth service
authService := authentication.NewService(authRepo, userRepo, mailService)
```

### En `src/core/auth/service.go`
```go
// El service recibe el mail service inyectado
type service struct {
    repo        Repo
    uRepo       uRepo
    mailService mail.EmailService  // ← Agnóstico
}

// En signup, envía email automáticamente
func (s *service) SignUpPassword(input SignUpPassword) error {
    // ... crear usuario ...
    
    // Generar token
    token, _ := mail.GenerateVerificationToken()
    
    // Enviar email (funciona con Mailpit o Resend)
    s.mailService.SendVerificationEmail(ctx, user.Email, user.Name, token)
    
    return nil
}
```

---

## ▶️ Paso 5: Ejecutar la App

```bash
# Cargar variables de .env
export $(cat .env | xargs)

# Ejecutar normalmente
go run cmd/main.go

# O con air (hot reload)
air
```

Output esperado:
```
2026/03/21 11:30:45 Server is running on http://localhost:3000
```

---

## 📧 Paso 6: Enviar Emails de Prueba

### Opción 1: Endpoint de Test (Recomendado)

**Endpoint especial para testing:**
```bash
curl -X POST http://localhost:3000/auth/send-test-email \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "name": "Test User"
  }'
```

**Respuesta:**
```json
{
  "message": "test email sent successfully",
  "email": "test@example.com"
}
```

**En logs verás:**
```
Test email sent to test@example.com
```

### Opción 2: Signup (con verificación)

```bash
curl -X POST http://localhost:3000/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "ip": "127.0.0.1",
    "userAgent": "curl"
  }'
```

**Qué pasa:**
1. Usuario se crea en BD
2. Token de verificación se genera
3. Email de verificación se envía automáticamente
4. Email aparece en Mailpit 2025

---

## 🌐 Paso 7: Ver Emails en Mailpit

### Web UI (http://localhost:8025)

Cuando envías un email:

1. **Aparece en Mailbox**
   - Asunto del email
   - From/To
   - Fecha/hora

2. **Haz click para ver detalles**
   - HTML renderizado
   - Links funcionales (click para verificar)
   - Headers MIME
   - Código fuente

3. **Características extras**
   - Buscar por asunto/email
   - Taggear mensajes
   - Test HTML
   - Screenshot HTML

### Verificación de Email

Si haces signup:
1. Email de verificación llega a Mailpit
2. Verás un link: `http://localhost:3000/auth/verify-email?token=...`
3. Haz click en el link (directamente desde Mailpit)
4. Con una migración de BD futura, el email se marcará verificado ✓

---

## 🔄 Cómo Funciona la Magia (Agnóstico)

### El Flujo Técnico

```
Tu Código (sin saber qué provider usa)
             ↓
Interface EmailService
             ↓
    s.mailService.SendVerificationEmail(...)
             ↓
        Factory Pattern
             ↓
    ¿EMAIL_PROVIDER?
    ├─ "mailpit" → MailpitService.SendVerificationEmail()
    └─ "resend"  → ResendService.SendVerificationEmail()
```

### Cambiar de Mailpit a Resend

**Opción A: Cambiar ENV (recomendado)**
```env
EMAIL_PROVIDER=resend
RESEND_API_KEY=re_tu_api_key_aqui
```

**Listo. El código NO cambia.**

**Opción B: En Producción**
```bash
# En tu servidor, cambiar .env:
EMAIL_PROVIDER=resend
RESEND_API_KEY=re_api_key_productiva
APP_URL=https://miapp.com
EMAIL_FROM=noreply@miapp.com

# Reiniciar app
systemctl restart mi-app
```

Los emails ahora van a Resend → usuarios reales.

---

## 🎁 Métodos Disponibles

### Interface EmailService

```go
// Enviar email de verificación (con token)
SendVerificationEmail(ctx context.Context, to, name, token string) error

// Enviar email de reset de contraseña
SendPasswordReset(ctx context.Context, to, name, resetLink string) error

// Enviar email de bienvenida
SendWelcome(ctx context.Context, to, name string) error
```

### Funciones Utilitarias (en src/service/mail/)

```go
// Generar token aleatorio (32 bytes hex)
token, err := mail.GenerateVerificationToken()

// Hashear token para guardar en BD (SHA256)
tokenHash := mail.HashToken(token)

// Generar token de reset (igual a verification)
resetToken, err := mail.GeneratePasswordResetToken()
```

---

## 🧪 Workflow Día a Día

### Mañana: Setup Inicial

```bash
# 1. Permisos Docker (si no lo hiciste)
sudo usermod -aG docker $USER
newgrp docker

# 2. Levantar Mailpit
docker-compose -f docker-compose.mail.yaml up -d

# 3. Verificar que corre
docker-compose -f docker-compose.mail.yaml ps

# 4. Ejecutar app
export $(cat .env | xargs)
go run cmd/main.go

# 5. Abrir web UI
open http://localhost:8025  # o tu navegador
```

### Durante Desarrollo

```bash
# Test endpoint en otra terminal
curl -X POST http://localhost:3000/auth/send-test-email \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "name": "Test"}'

# Ver email en Mailpit UI
# http://localhost:8025
```

### Cambios en Código

Si modificas templates de email (en `src/service/mail/mailpit.go` o `resend.go`):
1. App se reinicia (con air)
2. Los emails siguientes usan templates nuevos
3. Prueba en Mailpit Web UI

---

## 🐛 Troubleshooting

### "connection refused localhost:1025"
```
Problema: Mailpit no está corriendo
Solución: docker-compose -f docker-compose.mail.yaml up -d
```

### "Email service: missing required environment variables"
```
Problema: .env no tiene EMAIL_FROM, EMAIL_FROM_NAME o APP_URL
Solución: Verificar .env y cargar: export $(cat .env | xargs)
```

### "RESEND_API_KEY environment variable is required"
```
Problema: EMAIL_PROVIDER=resend pero RESEND_API_KEY no se set
Solución: Agregar RESEND_API_KEY a .env
```

### Email no aparece en Mailpit UI
```
Verificar:
1. ¿Mailpit está corriendo? → docker-compose ps
2. ¿MAILPIT_HOST=localhost? (no 127.0.0.1)
3. ¿Puerto 1025 correcto?
4. ¿Logs de app muestran error? → check STDOUT/STDERR
```

---

## 📚 Archivos Relacionados

- **Plan Completo:** `internal_docs/MAIL_IMPLEMENTATION_PLAN.md`
- **Setup Guide:** `internal_docs/MAIL_SETUP_GUIDE.md`
- **Test Email Guide:** `internal_docs/MAIL_TEST_EMAIL_GUIDE.md`
- **Email Verification:** `internal_docs/EMAIL_VERIFICATION.md`

---

## 📋 Resumen Rápido

| Componente | Para Desarrollo | Para Producción |
|-----------|-----------------|-----------------|
| Provider | Mailpit | Resend |
| Config | `EMAIL_PROVIDER=mailpit` | `EMAIL_PROVIDER=resend` |
| API Key | No necesita | `RESEND_API_KEY=...` |
| Dependencias | net/smtp (built-in) | resend-go v3 |
| Docker | `docker-compose.mail.yaml` | No necesita |
| Emails reales | No (capturados) | Sí (usuarios reciben) |
| Tracking | Web UI visual | Webhooks Resend |

---

## ✨ Próximas Mejoras (Futuro)

- [ ] Implementar `/auth/verify-email` handler para confirmar emails
- [ ] Implementar `/auth/resend-verification` endpoint
- [ ] Agregar webhooks de Resend para tracking
- [ ] Tests unitarios con mocks
- [ ] Logging y monitoring de emails
- [ ] Rate limiting en endpoints de email
- [ ] Soporte para attachments

---

**¿Preguntas? Mira la sección de Troubleshooting o revisa los archivos en `internal_docs/`**
