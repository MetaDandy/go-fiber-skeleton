# 🤖 AGENTS.md - Go Fiber Skeleton

Guía de supervivencia para agentes OpenCode trabajando en este repositorio.

## 🏗️ Arquitectura y Estructura

- **Framework**: [Fiber v3](https://gofiber.io) (Note: code uses `v3`, README says `v2`).
- **Patrón Modular**: Ubicado en `src/core/` (Handler → Service → Repo).
- **Inyección de Dependencias**: Centralizada en `src/container.go`. Agregá nuevos módulos ahí.
- **Modelos**: GORM en `src/model/`.
- **DTOs**: Entrada en `src/core/{module}/dto.go`, Salida en `src/response/`.
- **Manejo de Errores**: Usá `api_error` package (`api_error.BadRequest`, etc.).

### Convenciones Críticas
- **Nomenclatura**: Archivos dentro de módulos son genéricos (`handler.go`, `service.go`, `repo.go`).
- **Interfaces**: Cada capa define una `interface` pública e implementa con un `struct` privado.
- **GORM**: Usamos UUIDs como Primary Keys.
- **Soft Deletes**: Soportado vía `gorm.DeletedAt`.

## 🛠️ Comandos de Desarrollo

- **Hot Reload**: `air` (usa `.air.toml`).
- **Docker Dev**: `docker compose -f docker-compose.dev.yaml up --build`.
- **Database**: PostgreSQL. No uses `AutoMigrate`. Usá **Goose**.

### Migraciones (Goose)
- **Ubicación**: `migration/`. Formato: `NNN_description.sql`.
- **Ejecución**: Se corren automáticamente al iniciar la app si `AUTO_MIGRATE=true`.
- **Manual**: `cd migration && goose postgres $DATABASE_URL up`.

## 🔐 Autenticación y RBAC

- **JWT**: Validado vía `middleware.Jwt`. Claims extraídos a `c.Locals("user_id")`, etc.
- **RBAC**: Complejo sistema de Roles y Permisos en `src/core/role` y `src/core/permission`.
- **OAuth**: Soporta múltiples providers (Google, etc.) en `src/service/auth`.

## 🧪 Testing

- **API Testing**: Colección **Bruno** en `rest/`. Usala para verificar endpoints manualmente.
- **Unit Testing**: Correr con `go test ./...`.

## ⚠️ Gotchas y Errores Comunes

- **README desactualizado**: El `README.md` menciona `src/modules/` y `fiber v2`, pero el código real usa `src/core/` y `fiber v3`. Confiá en el código.
- **Mailpit vs Resend**: El sistema de mail (`src/service/mail`) es polimórfico. Mailpit para dev (puerto 1025/8025), Resend para prod.
- **DATABASE_URL**: Asegurate de que el scheme sea `postgres://` o `postgresql://` según requiera el driver.
