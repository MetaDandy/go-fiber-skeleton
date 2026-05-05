# 🚀 Go Fiber Skeleton

Una plantilla profesional y modular para iniciar proyectos con **Go Fiber**, diseñada para ser forkeada y utilizada como base de desarrollo. Este esqueleto proporciona una arquitectura escalable, buenas prácticas y ejemplos funcionales listos para expandir.

---

## 📋 Características

✅ **Framework REST**: [Fiber v3](https://gofiber.io) - Framework web ultra rápido inspirado en Express.js
✅ **Base de datos**: PostgreSQL con [GORM](https://gorm.io) (ORM moderno)
✅ **Migraciones**: Pressly [Goose](https://github.com/pressly/goose) para versionado de schema
✅ **Autenticación**: JWT (JSON Web Tokens) con middleware validación + context
✅ **Almacenamiento**: Integración con Supabase Storage (helper incluido)
✅ **Hot Reload**: Desarrollo local con [Air](https://github.com/cosmtrek/air)
✅ **Docker**: Multi-stage Dockerfile para producción, Dockerfile.dev para desarrollo
✅ **Seeding**: Sistema automático de seeders con validación (no duplicados)
✅ **Paginación**: FindAllOptions con search, sort, limit, offset, soft-delete filters
✅ **Testing API**: Colección Bruno para tests de endpoints (REST client integrado)
✅ **CORS**: Middleware CORS configurado con `ALLOW_ORIGINS`
✅ **Logging**: Logger personalizado + context en cada request
✅ **Error Handling**: Gestión explícita de errores con Fiber NewError
✅ **Arquitectura modular**: Patrón de capas (Handler → Service → Repository)
✅ **Inyección de dependencias**: Container pattern centralizado
✅ **Soft Deletes**: Soporte nativo en modelos (deleted_at field)
✅ **Validación**: Struct tags para validación de entrada
✅ **Retry Logic**: Reintentos automáticos en conexión BD (10 intentos, 2s delay)
✅ **GORM CLI**: Field helpers generados en `src/generated/` - ejecutar `make generate` tras cambios en modelos

---

## 🔧� GORM CLI Field Helpers

Este proyecto usa [GORM CLI](https://gorm.io/cli/gorm) para generar field helpers tipados en `src/generated/`. Estos helpers permiten escribir consultas type-safe sin strings SQL crudos.

### Comando de Generación

```bash
make generate
# O directamente:
gorm generate -i ./src/model -o ./src/generated
```

### Uso en Repos

En lugar de:
```go
// ❌ Antiguo (string-based)
db.Where("name ILIKE ?", searchPattern)
db.Where("user_id = ?", userID)
```

Usar:
```go
// ✅ Nuevo (field helpers tipados)
db.Where(generated.User.Name.ILike(searchPattern))
db.Where(generated.AuthProvider.UserID.Eq(userID))
```

### Archivos Generados

Los helpers se encuentran en `src/generated/`:
- `User.go`, `Role.go`, `Permission.go` - Modelos principales
- `Session.go`, `Auth_Provider.go`, etc. - Modelos de autenticación
- **NO editar manualmente** - se regeneran con `make generate`

---

## 📁 Estructura del Proyecto

```
go-fiber-skeleton/
├── cmd/                          # Punto de entrada de la aplicación
│   ├── main.go                   # Inicialización de app, middlewares, listener
│   └── api/
│       └── api.go               # Registro de rutas y handlers
│
├── src/                          # Lógica principal de la aplicación
│   ├── container.go             # Inyección de dependencias (DI container)
│   ├── enum/
│   │   └── status.go            # Enum Task.Status (pendiente, en_progreso, hecho)
│   ├── model/                   # Modelos GORM (DB schemas con relaciones)
│   │   ├── user.go              # Modelo Usuario (UUID, name, email)
│   │   └── task.go              # Modelo Tarea (UUID, title, description, status, user_id)
│   ├── response/                # DTOs de respuesta + convertidores
│   │   ├── paginated.go         # Wrapper genérico Paginated[T]
│   │   ├── user.go              # User response DTO + UserToDto()
│   │   └── task.go              # Task response DTO + TaskToDto(), TaskToListDto()
│   └── modules/                 # Módulos de negocio (User, Task)
│       ├── user/
│       │   ├── handler.go       # HTTP handlers (interface + impl)
│       │   ├── services.go      # Lógica de negocio (interface + impl)
│       │   ├── repo.go          # Acceso a datos GORM (interface + impl)
│       │   └── dto.go           # DTOs de entrada (Create, Update)
│       └── task/                # Mismo patrón que user
│           ├── handler.go       # Handlers CRUD + FindAll filtrado por user
│           ├── services.go      # Lógica + validación usuario existe
│           ├── repo.go          # Repo GORM + filtro user_id
│           └── dto.go           # DTOs con validación UUID
│
├── config/                       # Configuración e inicialización
│   ├── config.go                # Carga .env, conexión BD con reintentos
│   ├── migrate.go               # Corredor de migraciones Goose
│   └── seed/
│       ├── seeder.go            # Orquestador de seeders (entry point)
│       └── user.go              # Seed de usuarios de ejemplo
│
├── middleware/                   # Middlewares HTTP (Fiber)
│   ├── logger.go                # Logger de requests
│   ├── jwt.go                   # Validación JWT + extracción de claims
│   └── role.go                  # RequireRole() para control de acceso
│
├── migration/                    # Migraciones SQL (Goose)
│   └── 001_basic.sql            # Creación de tablas users, task, enum status
│
├── helper/                       # Funciones auxiliares reutilizables
│   ├── jwt.go                   # GenerateJwt() con 24h expiry
│   ├── hash.go                  # HashPassword(), CheckPasswordHash() (bcrypt)
│   ├── findall.go               # FindAllOptions, ApplyFindAllOptions()
│   └── supabase.go              # Upload() a Supabase Storage
│
├── rest/                         # Colección de tests API (Bruno)
│   ├── bruno.json               # Metadata de colección
│   ├── collection.bru           # Root collection
│   ├── aloha.bru                # Test de health check
│   ├── environments/            # Vars de env (local, prod, example)
│   ├── user/                    # Tests CRUD usuario
│   │   ├── create.bru
│   │   ├── findall.bru
│   │   ├── findbyid.bru
│   │   ├── update.bru
│   │   └── delete.bru
│   └── task/                    # Tests CRUD tareas
│       ├── create.bru
│       ├── findall.bru
│       ├── update.bru
│       └── delete.bru
│
├── Dockerfile                    # Build multi-stage para producción
├── Dockerfile.dev               # Build con Air para desarrollo
├── docker-compose.yaml          # Compose producción
├── docker-compose.dev.yaml      # Compose desarrollo (hot reload)
├── .air.toml                    # Configuración Air (watch, rebuild, output)
├── .env.example                 # Plantilla de variables de entorno
├── .env                         # Variables de entorno (local)
├── go.mod                       # Definición del módulo Go
├── go.sum                       # Checksums de dependencias
└── README.md                    # Este archivo
```

---

## 🔧 Instalación y Setup

### Requisitos Previos

- **Go** 1.20 o superior (se recomienda 1.25.5+)
- **PostgreSQL** 15+ (local o en la nube como Supabase)
- **Docker** y **Docker Compose** (opcional, para desarrollo/producción containerizado)
- **Posix shell** (bash, zsh, etc.) - para scripts de inicialización

### 1️⃣ Clonar/Forkar el Repositorio

```bash
# Forka este repositorio en GitHub, luego clona tu fork
git clone https://github.com/TU_USUARIO/go-fiber-skeleton.git
cd go-fiber-skeleton
```

### 2️⃣ Configurar Variables de Entorno

```bash
# Copia el archivo de ejemplo
cp .env.example .env

# Edita .env con tus valores
# Las variables necesarias son:
# - DATABASE_URL: Conexión a PostgreSQL (ej: postgresql://user:pass@host:5432/db)
# - PORT: Puerto de la aplicación (default: 8000)
# - ALLOW_ORIGINS: CORS origins (default: *)
# - AUTO_MIGRATE: Auto-migración de BD (true/false)
```

### 3️⃣ Instalar Dependencias

```bash
go mod download
```

### 4️⃣ Opción A: Ejecutar Localmente

```bash
# Instalar Air para hot reload
go install github.com/air-verse/air@latest

# Ejecutar con hot reload
air

# O ejecutar directamente
go run ./cmd
```

### 4️⃣ Opción B: Ejecutar con Docker

```bash
# Desarrollo con hot reload
docker compose -f docker-compose.dev.yaml up --build

# Producción
docker compose -f docker-compose.yaml up --build
```

---

## 🧬 Convenciones de Código y Patrones

### Nomenclatura de Archivos

| Layer | Patrón | Ejemplo |
|-------|--------|---------|
| **Modelos** | `singular.go` | `user.go`, `task.go` |
| **Enums** | `singular.go` | `status.go` |
| **Response DTOs** | `singular.go` | `user.go`, `task.go` |
| **Input DTOs** | `dto.go` (en módulo) | `user/dto.go`, `task/dto.go` |
| **Handlers** | `handler.go` | `user/handler.go`, `task/handler.go` |
| **Services** | `services.go` | `user/services.go`, `task/services.go` |
| **Repositories** | `repo.go` | `user/repo.go`, `task/repo.go` |
| **Middleware** | `nombreMiddleware.go` | `jwt.go`, `logger.go`, `role.go` |

**⚠️ Nota Importante**: Los archivos NO llevan sufijo de layer (no es `user_handler.go`, es solo `handler.go`). Cada paquete contiene múltiples archivos (`handler.go`, `services.go`, `repo.go`, `dto.go`) que implementan las interfaces de esa capa.

### Interfaz → Struct Pattern

Cada capa (Handler, Service, Repo) define una interface pública e implementa con un struct privado:

```go
// interface pública
type Handler interface {
    Create(c *fiber.Ctx) error
    FindAll(c *fiber.Ctx) error
    // ...
}

// struct privada que implementa
type handler struct {
    service Service
}

// constructor que retorna la interface
func NewHandler(service Service) Handler {
    return &handler{service: service}
}
```

### Converso de DTOs

Todas las conversiones entre capas usan funciones converter nombradas `{Type}ToDto()`:

```go
// response/user.go
func UserToDto(u *model.User) User { ... }
func UserToListDto(users []model.User) []User { ... }

// response/task.go
func TaskToDto(t *model.Task) Task { ... }
func TaskToListDto(tasks []model.Task) []Task { ... }
```

### Manejo de Errores Explícito

Se retornan errores explícitamente en cada nivel:

```go
// Repository layer
func (r *repo) FindByID(id string) (model.Task, error) {
    var task model.Task
    err := r.db.First(&task, "id = ?", id).Error
    return task, err  // ← error explícito
}

// Service layer
func (s *service) FindByID(id string) (*response.Task, error) {
    task, err := s.repo.FindByID(id)
    if err != nil {
        return nil, err  // ← se propaga
    }
    // ...
}

// Handler layer
func (h *handler) FindByID(c *fiber.Ctx) error {
    id := c.Params("id")
    finded, err := h.service.FindByID(id)
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": err.Error(),
        })
    }
    return c.Status(fiber.StatusOK).JSON(finded)
}
```

---

## 📚 Modelos y Relaciones

### User (Usuario)

```go
type User struct {
    ID        uuid.UUID           // Primary Key
    Name      string              // Nombre del usuario
    Email     string              // Unique Index
    Tasks     []Task              // Relación 1 → Many
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt      // Soft delete
}
// TableName: "users" (plural)
```

- **Relación**: 1 Usuario → Many Tareas
- **Endpoints**: 
  - `POST   /api/v1/users` → Create
  - `GET    /api/v1/users?limit=10&offset=0&search=...` → FindAll (paginado)
  - `GET    /api/v1/users/:id` → FindByID
  - `PATCH  /api/v1/users/:id` → Update (campos parciales)
  - `DELETE /api/v1/users/:id` → Delete (soft delete)

### Task (Tarea)

```go
type Status enum{
    "pendiente"    // initial
    "en_progreso"  // in progress
    "hecho"        // done
}

type Task struct {
    ID          uuid.UUID           // Primary Key
    Title       string              // Título de la tarea
    Description string              // Descripción
    Status      enum.Status         // PostgreSQL ENUM
    UserID      uuid.UUID           // Foreign Key → users (CASCADE)
    User        User                // Relación (Preload automático)
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   gorm.DeletedAt      // Soft delete
}
// TableName: "task" (singular) ⚠️
```

- **Relación**: Many Tareas → 1 Usuario
- **Endpoints**: 
  - `POST   /api/v1/tasks` → Create (requiere user_id válido)
  - `GET    /api/v1/tasks?limit=10&offset=0` → FindAll (filtrado por usuario autenticado)
  - `GET    /api/v1/tasks/:id` → FindByID (con User preloaded)
  - `PATCH  /api/v1/tasks/:id` → Update (campos parciales)
  - `DELETE /api/v1/tasks/:id` → Delete (soft delete)

---

## 🔑 Variables de Entorno

Todas las variables se cargan desde `.env` (gitignored) usando [godotenv](https://github.com/joho/godotenv).

### Variables Requeridas

```bash
# Base de datos PostgreSQL (requerido)
DATABASE_URL=postgresql://user:password@localhost:5432/dbname
```

### Variables Opcionales

```bash
# Aplicación
PORT=8001                        # Puerto de escucha (default: 8001)
ALLOW_ORIGINS=*                  # CORS origins (default: *)
AUTO_MIGRATE=true                # Auto-migrar con Goose (default: true)
GO_ENV=development               # development|production (default: development)

# Autenticación JWT
JWT_SECRET=tu_secret_muy_seguro_aqui  # Secret para firmar JWT (requerido si usas JWT)
```

### Cargar Archivo .env

1. Copiar plantilla:
   ```bash
   cp .env.example .env
   ```

2. Editar `.env` con tus valores

3. Valores se cargan automáticamente al iniciar app (en [config/config.go](config/config.go))

**⚠️ Nota Seguridad**: Nunca commitees `.env` a git. Usa `.env.example` como plantilla.

---

## 🚀 Endpoints Principales

### Health Check
```bash
GET  /
# Respuesta: "Aloha"
```

### Users (CRUD)
```bash
POST   /api/v1/users                          # Crear usuario
       # Body: { "name": "Juan", "email": "juan@example.com" }

GET    /api/v1/users                          # Listar usuarios (paginado)
       # Query params: ?limit=10&offset=0&search=...&order_by=...&sort=asc

GET    /api/v1/users/:id                      # Obtener usuario por ID

PATCH  /api/v1/users/:id                      # Actualizar usuario (campos parciales)
       # Body: { "name": "Juan Updated" }

DELETE /api/v1/users/:id                      # Eliminar usuario (soft delete)
```

### Tasks (CRUD) - Filtrado por Usuario Autenticado
```bash
POST   /api/v1/tasks                          # Crear tarea
       # Body: { "title": "...", "description": "...", "user_id": "..." }
       # ⚠️ Requiere JWT válido

GET    /api/v1/tasks                          # Listar tareas del usuario autenticado (paginado)
       # Query params: ?limit=10&offset=0&search=...
       # ⚠️ Requiere JWT - solo tareas del usuario

GET    /api/v1/tasks/:id                      # Obtener tarea por ID

PATCH  /api/v1/tasks/:id                      # Actualizar tarea (campos parciales)
       # Body: { "title": "...", "status": "hecho" }

DELETE /api/v1/tasks/:id                      # Eliminar tarea (soft delete)
```

### Parámetros de Paginación (Todos los FindAll)

```bash
# Query string params disponibles:
limit=10          # Registros por página (default: 10, max: 30)
offset=0          # Desplazamiento (default: 0)
order_by=created_at  # Campo para ordenar (default: created_at)
sort=desc         # asc o desc (default: desc)
search=query      # Búsqueda full-text (ILIKE) en múltiples campos
show_deleted=false   # Mostrar soft-deleted (default: false)
only_deleted=false   # Solo soft-deleted (default: false)

# Ejemplo:
GET /api/v1/users?limit=20&offset=0&search=juan&sort=asc&order_by=name
```

### Response Paginado

```json
{
  "data": [
    { "id": "550e8400-e29b-41d4-a716-446655440000", "name": "Juan", "email": "juan@example.com" },
    { "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8", "name": "María", "email": "maria@example.com" }
  ],
  "total": 2,
  "limit": 20,
  "offset": 0,
  "pages": 1
}
```

---

## 📦 Dependencias Principales

| Paquete | Uso |
|---------|-----|
| `gofiber/fiber/v2` | Framework web REST |
| `gorm.io/gorm` | ORM para base de datos |
| `gorm.io/driver/postgres` | Driver PostgreSQL para GORM |
| `golang-jwt/jwt/v5` | Generación y validación JWT |
| `google/uuid` | Generación UUIDs (usado como PK) |
| `joho/godotenv` | Carga de variables `.env` |
| `jinzhu/copier` | Copia entre structs (DTOs) |
| `pressly/goose/v3` | Migration runner versionado |
| `golang.org/x/crypto` | Bcrypt para hash de contraseñas |
| `lib/pq` | Driver nativo PostgreSQL |

Ver [go.mod](go.mod) para versiones exactas.

---

## 🛠️ Desarrollo

### Hot Reload Local

```bash
# Instalar Air (solo primera vez)
go install github.com/air-verse/air@latest

# Ejecutar con watch automático
air

# Air compilará y reiniciará automáticamente al guardar cambios
```

### Hot Reload con Docker

```bash
docker compose -f docker-compose.dev.yaml up

# Los cambios se reflejan automáticamente en el contenedor
```

### Agregar un Nuevo Módulo

*Ejemplo: Crear módulo `post` con CRUD completo*

**Paso 1**: Crear estructura de carpetas
```bash
mkdir -p src/modules/post
```

**Paso 2**: Crear archivos del módulo

- **dto.go** - Data Transfer Objects de entrada:
```go
package post

type Create struct {
    Title   string `validate:"required"`
    Content string
    AuthorID string `validate:"required,uuid"`
}

type Update struct {
    Title    *string
    Content  *string
}
```

- **repo.go** - Repository (CRUD en BD):
```go
package post

import (
    "github.com/MetaDandy/go-fiber-skeleton/src/model"
    "gorm.io/gorm"
)

type Repo interface {
    Create(m model.Post) error
    FindByID(id string) (model.Post, error)
    FindAll(opts *helper.FindAllOptions) ([]model.Post, int64, error)
    Update(m model.Post) error
    Delete(id string) error
}

type repo struct {
    db *gorm.DB
}

func NewRepo(db *gorm.DB) Repo {
    return &repo{db: db}
}

func (r *repo) Create(m model.Post) error {
    return r.db.Create(&m).Error
}

// ... resto de métodos
```

- **services.go** - Business Logic:
```go
package post

type Service interface {
    Create(input Create) error
    FindByID(id string) (*response.Post, error)
    FindAll(opts *helper.FindAllOptions) (*response.Paginated[response.Post], error)
    Update(id string, input Update) error
    Delete(id string) error
}

type service struct {
    repo Repo
}

func NewService(repo Repo) Service {
    return &service{repo: repo}
}

// ... implementar métodos de interface
```

- **handler.go** - HTTP Handlers:
```go
package post

import "github.com/gofiber/fiber/v2"

type Handler interface {
    RegisterRoutes(router fiber.Router)
    Create(c *fiber.Ctx) error
    FindAll(c *fiber.Ctx) error
    FindByID(c *fiber.Ctx) error
    Update(c *fiber.Ctx) error
    Delete(c *fiber.Ctx) error
}

type handler struct {
    service Service
}

func NewHandler(service Service) Handler {
    return &handler{service: service}
}

func (h *handler) RegisterRoutes(router fiber.Router) {
    posts := router.Group("/posts")
    posts.Post("/", h.Create)
    posts.Get("/", h.FindAll)
    posts.Get("/:id", h.FindByID)
    posts.Patch("/:id", h.Update)
    posts.Delete("/:id", h.Delete)
}

// ... implementar handlers
```

**Paso 3**: Crear modelo en `src/model/post.go`:
```go
package model

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type Post struct {
    ID        uuid.UUID
    Title     string
    Content   string
    AuthorID  uuid.UUID
    Author    User
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Post) TableName() string {
    return "posts"
}
```

**Paso 4**: Crear response DTO en `src/response/post.go`:
```go
package response

import "github.com/MetaDandy/go-fiber-skeleton/src/model"

type Post struct {
    ID        string
    Title     string
    Content   string
    AuthorID  string
    CreatedAt string
}

func PostToDto(p *model.Post) Post {
    return Post{
        ID:        p.ID.String(),
        Title:     p.Title,
        Content:   p.Content,
        AuthorID:  p.AuthorID.String(),
        CreatedAt: p.CreatedAt.String(),
    }
}
```

**Paso 5**: Registrar en DI Container - [src/container.go](src/container.go):
```go
// Agregar imports
import "github.com/MetaDandy/go-fiber-skeleton/src/modules/post"

// En estructura Container, agregar:
type Container struct {
    UserRepo   user.Repo
    UserSvc    user.Service
    UserHandler user.Handler
    // ... agregar:
    PostRepo    post.Repo
    PostSvc     post.Service
    PostHandler post.Handler
}

// En NewContainer():
c.PostRepo = post.NewRepo(db)
c.PostSvc = post.NewService(c.PostRepo)
c.PostHandler = post.NewHandler(c.PostSvc)

return &c
```

**Paso 6**: Registrar rutas en [cmd/api/api.go](cmd/api/api.go):
```go
func Setup(app *fiber.App, container *src.Container) {
    api := app.Group("/api/v1")
    // ... rutas existentes
    
    // Agregar:
    container.PostHandler.RegisterRoutes(api)
}
```

**Paso 7**: Crear migración SQL:
```bash
cat > migration/002_create_posts.sql << 'EOF'
-- +goose Up
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    content TEXT,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_posts_author_id ON posts(author_id);

-- +goose Down
DROP TABLE posts;
EOF
```

¡Listo! Tu nuevo módulo `post` está lista para usar.

### Crear una Migración

Este proyecto usa [Pressly Goose](https://github.com/pressly/goose) para gestión de migraciones SQL versionadas.

**Ubicación**: `migration/` directorio con formato: `NNN_description.sql`

**Para agregar una migración**:

1. Crear archivo numerado: `migration/002_add_columns.sql`
2. Escribir SQL con comentarios de Goose:

```sql
-- +goose Up
ALTER TABLE users ADD COLUMN phone VARCHAR(20);

-- +goose Down
ALTER TABLE users DROP COLUMN phone;
```

3. Goose ejecutará automáticamente al iniciar la app (vía `config/migrate.go`)

**Nota**: Goose maneja las versiones automáticamente. No uses AutoMigrate de GORM.

### Crear un Seeder

1. Crear archivo `src/model/whatever_seeder.go` en `config/seed/`
2. Implementar función `SeedWhatevers(db *gorm.DB) error`
3. Llamar desde `config/seed/seeder.go`

---

## 🐳 Docker

### Producción

```bash
# Build y run
docker compose -f docker-compose.yaml up --build

# La app escucha en el puerto especificado en .env
```

**Dockerfile**: Multi-stage build que:
1. Compila la app en Alpine con Go
2. Copia solo el binario a una imagen mínima de Alpine
3. Resultado: ~50-60 MB vs 400+ MB con la imagen base de Go

### Desarrollo

```bash
# Build con Air para hot reload
docker compose -f docker-compose.dev.yaml up --build

# Cambios se reflejan automáticamente
```

**Dockerfile.dev**: Imagen con Air incluido para desarrollo rápido

---

## 🔐 Autenticación y Autorización

### JWT Middleware

Para rutas protegidas por JWT, agrega el middleware a la ruta:

```go
// En cmd/api/api.go o donde registres rutas:
api := app.Group("/api/v1")

// Rutas públicas (sin middleware)
api.Post("/auth/login", handler.Login)

// Rutas protegidas con JWT
api.Use(middleware.Jwt())
api.Get("/tasks", handler.ListTasks)      // Solo usuarios autenticados
api.Post("/tasks", handler.CreateTask)
```

### Cómo funciona JWT Middleware

1. Valida token en header: `Authorization: Bearer <token>`
2. Extrae claims del JWT
3. Almacena en contexto Fiber:
   - `c.Locals("user_id")` → ID del usuario
   - `c.Locals("email")` → Email del usuario
   - `c.Locals("role")` → Rol del usuario (si existe)
4. Si hay error, retorna 401 Unauthorized

### Generar JWT Token

```go
import "github.com/MetaDandy/go-fiber-skeleton/helper"

userID := "550e8400-e29b-41d4-a716-446655440000"
email := "user@example.com"
role := "admin"

token, err := helper.GenerateJwt(userID, email, role)
if err != nil {
    log.Fatal(err)
}
// token es válido por 24 horas
// Usar en header: Authorization: Bearer {token}
```

### Role-Based Access Control (RBAC)

```go
// Proteger ruta que requiere rol específico:
api.Use(middleware.RequireRole("admin"))
api.Delete("/users/:id", handler.DeleteUser)

// Sistema actualmente soporta:
// - "admin": Acceso total
// - "user": Acceso limitado
// - Personalizable en middleware/role.go
```

### Validar Contraseñas (Bcrypt)

```go
import "github.com/MetaDandy/go-fiber-skeleton/helper"

// Hash: generar hash seguro de contraseña
hashedPassword, err := helper.HashPassword("miContraseña123")

// Verificar: comparar contraseña con hash
isValid := helper.CheckPasswordHash("miContraseña123", hashedPassword)
if !isValid {
    return fiber.NewError(fiber.StatusUnauthorized, "Contraseña incorrecta")
}

---

## 📊 Paginación y Búsqueda

El helper `FindAll` proporciona paginación out-of-the-box:

```bash
# Ejemplo de request
GET /api/v1/users?limit=10&offset=0&order_by=created_at&sort=desc&search=juan

# Parámetros disponibles:
# - limit: Registros por página (max 30, default 30)
# - offset: Desplazamiento (default 0)
# - order_by: Campo para ordenar (default created_at)
# - sort: asc o desc (default desc)
# - search: Búsqueda por nombre (ILIKE)
# - show_deleted: Mostrar soft-deleted (default false)
# - only_deleted: Solo soft-deleted (default false)
```

**Response paginado**:

```json
{
  "data": [
    { "id": "...", "name": "Juan", "email": "..." },
    { "id": "...", "name": "María", "email": "..." }
  ],
  "total": 2,
  "limit": 10,
  "offset": 0,
  "pages": 1
}
```

---

## 🌐 CORS Configuration

CORS está configurado en [cmd/main.go](cmd/main.go) usando Fiber middleware:

```go
import "github.com/gofiber/fiber/v2/middleware/cors"

app.Use(cors.New(cors.Config{
    AllowOrigins: os.Getenv("ALLOW_ORIGINS"), // "*" permite todos
    AllowMethods: "GET,POST,PATCH,DELETE,OPTIONS",
    AllowHeaders: "Origin, Content-Type, Accept, Authorization",
    AllowCredentials: true,
}))
```

**Variables de entorno relacionadas**:
```bash
ALLOW_ORIGINS=*                    # "*" = cualquier origen
# o específico:
ALLOW_ORIGINS=http://localhost:3000,https://app.example.com
```

Para modificar CORS, edita [cmd/main.go](cmd/main.go) directamente.

---

## 📝 Logging

Logger personalizado se aplica en cada request mediante middleware en [middleware/logger.go](middleware/logger.go).

**Formato de log actual**:
```
📢 GET /api/v1/users
📢 POST /api/v1/tasks (body preview)
```

**Cómo personalizar logs**:

Edita [middleware/logger.go](middleware/logger.go):

```go
func Logger() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Aquí puedes:
        // - Log con library externa (logrus, slog, zap)
        // - Parsear body y loguear datos sensibles
        // - Agregar request ID
        // - Enviar a servicio de logs (Datadog, ELK, etc)
        
        fmt.Printf("📢 %s %s\n", c.Method(), c.Path())
        return c.Next()
    }
}
```

**Para producción recomendamos**:
- [sirupsen/logrus](https://github.com/sirupsen/logrus) - Structured logging
- [uber-go/zap](https://github.com/uber-go/zap) - Performance logging
- [Datadog](https://docs.datadoghq.com/logs/) - Centralized logging

---

## 🧪 Testing - API Testing con Bruno

Este proyecto incluye una **colección Bruno** pre-configurada para testing de endpoints: [rest/](rest/)

### Instalar Bruno

```bash
# macOS (Homebrew)
brew install bruno

# Linux
curl -fsSL https://app.usebruno.com/install/linux.sh | bash

# O descargar desde: https://app.usebruno.com/downloads
```

### Estructura de Tests

```
rest/
├── collection.bru              # Root collection
├── bruno.json                  # Metadata
├── aloha.bru                   # Health check test
├── environments/               # Variable environments
│   ├── carpyen.local.bru       # Local dev env
│   ├── carpyen.prod.bru        # Production env
├── user/                       # User module tests
│   ├── create.bru
│   ├── findall.bru
│   ├── findbyid.bru
│   ├── update.bru
│   └── delete.bru
└── task/                       # Task module tests
    ├── create.bru
    ├── findall.bru
    ├── update.bru
    └── delete.bru
```

### Usar Bruno

1. Abrir Bruno
2. Importar: `File → Open Collection → Seleccionar carpeta rest/`
3. Seleccionar environment: `carpyen.local` (o la que uses)
4. Ejecutar tests: Click en cada request o Run Collection

### Agregar Tests Nuevos

1. En Bruno: `+ New Request`
2. Configurar: Method, URL, Headers, Body
3. Guardar como: `rest/module/operation.bru`
4. Usar variables env: `{{BASE_URL}}/api/v1/users`

### Unit Testing (Go)

Para agregar tests en Go (actualmente vacío):

```bash
# Crear tests
mkdir -p src/modules/user
cat > src/modules/user/services_test.go << 'EOF'
package user

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
    // Test aquí
    assert.Equal(t, true, true)
}
EOF

# Ejecutar tests
go test ./...

# Con coverage
go test -cover ./...
```



---

## 🚢 Deployment a Producción

### Opción 1: Railway.app (Recomendado - Simple)

1. Push código a GitHub
2. Ir a [railway.app](https://railway.app), crear cuenta
3. Create New Project → GitHub Repo
4. Conectar repositorio
5. Railway detecta Go automáticamente
6. Agregar variables en Variables tab:
   ```
   DATABASE_URL=postgresql://...
   JWT_SECRET=tu_secret
   ALLOW_ORIGINS=https://tudominio.com
   ```
7. Deploy - Railway corre `go run ./cmd`

### Opción 2: Render.com (Alternativa)

1. Push a GitHub
2. Crear nuevo Web Service en [render.com](https://render.com)
3. Conectar repositorio
4. Configurar:
   - **Build Command**: `go build -o app ./cmd`
   - **Start Command**: `./app`
5. Agregar environment variables en Environment
6. Deploy

### Opción 3: Docker en VPS/Cloud

```bash
# Build imagen
docker build -f Dockerfile -t go-fiber-app:latest .

# Crear .env en VPS con vars producción
# Luego:
docker run -p 8000:8000 --env-file .env go-fiber-app:latest

# O con docker-compose:
docker compose -f docker-compose.yaml up -d
```

### Pre-Flight Checklist

Antes de deployar a producción:

- [ ] Cambiar `JWT_SECRET` a valor seguro (generate con: `openssl rand -base64 32`)
- [ ] Configurar `ALLOW_ORIGINS` con dominios reales (no `*`)
- [ ] Verificar `DATABASE_URL` apunta a BD producción
- [ ] Revisar logs de error con `docker logs` o plataforma
- [ ] Usar HTTPS (certificado SSL/TLS)
- [ ] Configurar backups automáticos de BD
- [ ] Monitoreo: Datadog, New Relic, o similar
- [ ] Rate limiting (opcional, agregar middleware)
- [ ] CORS headers según necesidad

---

## � Troubleshooting & Common Issues

### Error: `could not connect to database`

**Causa**: Variable `DATABASE_URL` inválida o BD no accesible

**Solución**:
```bash
# Verificar .env existe y tiene DATABASE_URL
cat .env | grep DATABASE_URL

# Testear conexión PostgreSQL directo
psql postgresql://user:pass@host:5432/dbname

# Si usas Supabase, obtener connection string desde dashboard
# Railway: revisar variables en tab Environment
```

### Error: `migration: migration not found`

**Causa**: Archivo migration no encontrado o nombre inválido

**Solución**:
```bash
# Archivos deben estar en migration/ con formato NNN_description.sql
ls migration/
# Output: 001_basic.sql ✓

# Si agregaste migración: revisar nombre
# El número debe ser secuencial y válido para Goose

# Ejecutar migraciones manualmente:
cd migration && goose postgres $DATABASE_URL up
```

### Error: `port already in use`

**Causa**: Otra app escucha en puerto 8001 (o el configurado)

**Solución**:
```bash
# Cambiar puerto en .env
PORT=8002

# O matar proceso existente:
lsof -i :8001          # Ver qué usa el puerto
kill -9 <PID>          # Matar proceso
```

### Error: `JWT token invalid or expired`

**Causa**: Token expirado (24h) o `JWT_SECRET` diferente

**Solución**:
```bash
# Generar token nuevo desde login
POST /api/v1/auth/login

# Si cambió JWT_SECRET, tokens antiguos serán inválidos
# Regenerar token o usar el mismo SECRET en servidor

# Verificar SECRET en .env:
cat .env | grep JWT_SECRET
```

### Error: `CORS request blocked`

**Causa**: `ALLOW_ORIGINS` no incluye cliente

**Solución**:
```bash
# En .env, agregar origen cliente:
ALLOW_ORIGINS=http://localhost:3000,https://app.example.com

# Si en desarrollo: usa *
ALLOW_ORIGINS=*

# Luego reiniciar app
```

### Hot Reload (Air) no funciona

**Causa**: Air no reinstalado o .air.toml inválido

**Solución**:
```bash
# Reinstalar Air
go install github.com/air-verse/air@latest

# Verificar .air.toml existe en raíz
file .air.toml

# Ejecutar de nuevo
air
```

### Tests con Bruno fallan

**Causa**: Variables environment no configuradas

**Solución**:
1. Abrir Bruno
2. Seleccionar environment (carpyen.local, etc)
3. Configurar URL base y variables
4. Re-ejecutar test

---

## 📖 Ressources & Links

### Official Docs
- [Fiber Documentation](https://docs.gofiber.io)
- [GORM Documentation](https://gorm.io)
- [PostgreSQL Docs](https://www.postgresql.org/docs)
- [JWT RFC 7519](https://tools.ietf.org/html/rfc7519)

### Tools
- [Goose Migrations](https://github.com/pressly/goose)
- [Bruno API Client](https://www.usebruno.com/)
- [Railway.app Hosting](https://railway.app)
- [Supabase Postgres Hosting](https://supabase.com)

### Best Practices
- [Clean Code Go](https://github.com/golang-standards/project-layout)
- [12-Factor App](https://12factor.net/)
- [REST API Best Practices](https://restfulapi.net/)

---

## 🤝 Contribuciones

Este es un esqueleto para tu uso personal. Si lo mejoras y quieres compartir, considera hacer un PR al original.

---

## 📄 Licencia

MIT - Libre para usar en proyectos personales y comerciales.

---

## 🎯 Próximos Pasos

### Para comenzar desarrollo inmediato:

1. ✅ Clonar/Forkar este repositorio
2. ✅ Copiar `.env.example` → `.env` y configurar `DATABASE_URL`
3. ✅ Ejecutar `air` (o `docker compose -f docker-compose.dev.yaml up`)
4. ✅ Verificar health check: `curl http://localhost:8001/` → "Aloha"
5. ✅ Importar colección Bruno [rest/](rest/) para testear endpoints
6. ✅ Crear variantes del módulo `user` o `task` según necesidad
7. ✅ Deployar a Railway.app o servicio de tu preferencia

---

## 📄 Licencia

MIT - Libre para usar en proyectos personales y comerciales.

---

**Última actualización**: Marzo 2026  
**Versión Go**: 1.25.5+  
**Versión Fiber**: v2.52.10+
