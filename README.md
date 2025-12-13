# 🚀 Go Fiber Skeleton

Una plantilla profesional y modular para iniciar proyectos con **Go Fiber**, diseñada para ser forkeada y utilizada como base de desarrollo. Este esqueleto proporciona una arquitectura escalable, buenas prácticas y ejemplos funcionales listos para expandir.

---

## 📋 Características

✅ **Framework REST**: [Fiber v2](https://gofiber.io) - Framework web ultra rápido inspirado en Express.js
✅ **Base de datos**: PostgreSQL con [GORM](https://gorm.io) (ORM moderno)
✅ **Autenticación**: JWT (JSON Web Tokens) con middleware
✅ **Almacenamiento**: Integración con Supabase Storage
✅ **Hot Reload**: Desarrollo local con [Air](https://github.com/cosmtrek/air)
✅ **Docker**: Dockerfiles para desarrollo y producción
✅ **Seeding**: Sistema de seeders para datos iniciales
✅ **Paginación**: Helper para FindAll con soporte a búsqueda y ordenamiento
✅ **CORS**: Middleware CORS configurado
✅ **Logging**: Logger personalizado
✅ **Estructura modular**: Patrón de capas (Handler → Service → Repository)

---

## 📁 Estructura del Proyecto

```
go-fiber-skeleton/
├── cmd/                          # Punto de entrada de la aplicación
│   ├── main.go                   # main() del proyecto
│   └── api/
│       └── api.go               # Configuración y registro de rutas
│
├── src/                          # Lógica principal de la aplicación
│   ├── model/                   # Modelos GORM (Database schemas)
│   │   ├── user.go
│   │   └── task.go
│   ├── enum/                    # Enumeraciones
│   │   └── status.go
│   ├── response/                # DTOs de respuesta (Response Transfer Objects)
│   │   ├── user_response.go
│   │   └── task_response.go
│   ├── modules/                 # Módulos de negocio
│   │   ├── user/
│   │   │   ├── user_handler.go   # HTTP handlers
│   │   │   ├── user_service.go   # Lógica de negocio
│   │   │   ├── user_repo.go      # Acceso a datos
│   │   │   └── user_dto.go       # Data Transfer Objects (entrada)
│   │   └── task/                # Mismo patrón que user
│   │       ├── task_handler.go
│   │       ├── task_service.go
│   │       ├── task_repo.go
│   │       └── task_dto.go
│   └── container.go             # Inyección de dependencias (DI)
│
├── config/                       # Configuración e inicialización
│   ├── config.go                # Carga de .env y conexión BD
│   ├── migrate.go               # Auto-migraciones de GORM
│   └── seed/
│       ├── seeder.go            # Orquestador de seeders
│       └── user_seeder.go       # Seeder de datos de ejemplo
│
├── middleware/                   # Middlewares HTTP
│   ├── logger.go                # Logger personalizado
│   ├── jwt_middleware.go        # Validación de JWT
│   └── role_middleware.go       # Control de roles
│
├── helper/                       # Funciones auxiliares reutilizables
│   ├── jwt.go                   # Generación de JWT
│   ├── hash.go                  # Hash de contraseñas (bcrypt)
│   ├── findall.go               # Paginación y búsqueda
│   ├── response.go              # Estructuras de respuesta
│   └── supabase.go              # Cliente de Supabase Storage
│
├── Dockerfile                    # Imagen para producción (multi-stage)
├── Dockerfile.dev               # Imagen para desarrollo con Air
├── docker-compose.yaml          # Compose para producción
├── docker-compose.dev.yaml      # Compose para desarrollo
├── .air.toml                    # Configuración de Air (hot reload)
├── .env.example                 # Plantilla de variables de entorno
├── go.mod                       # Dependencias de Go
├── go.sum                       # Checksums de dependencias
└── README.md                    # Este archivo
```

---

## 🔧 Instalación y Setup

### Requisitos Previos

- **Go** 1.25.5 o superior
- **PostgreSQL** 15+ (local o en la nube como Supabase)
- **Docker** y **Docker Compose** (opcional, para desarrollo containerizado)

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

## 🏗️ Arquitectura

### Patrón de Capas (Layered Architecture)

Cada módulo (ej: `user`, `task`) sigue este patrón:

```
┌─────────────────────────────────────┐
│      Handler (HTTP Layer)           │  ← Recibe requests HTTP
│  - Parsea input (DTOs)              │
│  - Llama al Service                 │
│  - Retorna Response                 │
└────────────────┬────────────────────┘
                 │
┌────────────────▼────────────────────┐
│      Service (Business Logic)       │  ← Lógica de negocio
│  - Valida datos                     │
│  - Orquesta operaciones             │
│  - Llama al Repository              │
└────────────────┬────────────────────┘
                 │
┌────────────────▼────────────────────┐
│  Repository (Data Access Layer)     │  ← Acceso a BD
│  - Queries GORM                     │
│  - CRUD operations                  │
│  - Preload de relaciones            │
└─────────────────────────────────────┘
```

### Flujo de Datos

```
HTTP Request
    ↓
Handler.Create() → Parsea CreateUserDto
    ↓
Service.Create() → Valida y crea User
    ↓
Repo.Create() → Ejecuta INSERT en BD
    ↓
Service retorna UserResponse
    ↓
Handler retorna JSON response
    ↓
HTTP Response
```

---

## 📚 Modelos y Relaciones

### User (Usuario)

```go
type User struct {
    ID    uuid.UUID // Primary Key
    Name  string
    Email string    // Unique Index
}
```

- **Relación**: 1 → Many con Task
- **Endpoints**: 
  - `POST   /api/v1/users`
  - `GET    /api/v1/users`
  - `GET    /api/v1/users/:id`
  - `PUT    /api/v1/users/:id`
  - `DELETE /api/v1/users/:id`

### Task (Tarea)

```go
type Task struct {
    ID          uuid.UUID
    Title       string
    Description string
    Status      StatusEnum // pending | active | approved
    UserID      uuid.UUID  // Foreign Key
    User        User       // Relación (Preload automático)
}
```

- **Relación**: Many → 1 con User
- **Endpoints**: Mismos que User pero en `/tasks`

---

## 🔑 Variables de Entorno

```bash
# Base de datos (requerido)
DATABASE_URL=postgresql://user:password@localhost:5432/dbname

# Aplicación
PORT=8000                    # Puerto de escucha
ALLOW_ORIGINS=*             # CORS: * permite todos los orígenes
AUTO_MIGRATE=true           # Auto-migrar esquema BD al iniciar

# Autenticación (opcional, si usas JWT)
JWT_SECRET=tu_secret_muy_seguro_aqui

# Supabase (opcional, si usas Storage)
SUPABASE_PROJECT_URL=https://xxxxx.supabase.co
SUPABASE_API_KEY_SERVICE_ROLE=xxxxx
```

---

## 🚀 Endpoints Principales

### Health Check
```bash
GET  /
# Respuesta: "Aloha"
```

### Users
```bash
POST   /api/v1/users                    # Crear usuario
GET    /api/v1/users?limit=30&offset=0 # Listar (paginado)
GET    /api/v1/users/:id                # Obtener por ID
PUT    /api/v1/users/:id                # Actualizar
DELETE /api/v1/users/:id                # Eliminar
```

### Tasks
```bash
POST   /api/v1/tasks                    # Crear tarea
GET    /api/v1/tasks?limit=30&offset=0 # Listar (paginado)
GET    /api/v1/tasks/:id                # Obtener por ID (con User preloaded)
PUT    /api/v1/tasks/:id                # Actualizar
DELETE /api/v1/tasks/:id                # Eliminar
```

---

## 📦 Dependencias Principales

| Paquete | Versión | Propósito |
|---------|---------|-----------|
| `gofiber/fiber` | v2.52.10 | Framework web REST |
| `gorm.io/gorm` | v1.25.10 | ORM para base de datos |
| `gorm.io/driver/postgres` | v1.6.0 | Driver PostgreSQL para GORM |
| `golang-jwt/jwt` | v5.3.0 | Generación y validación JWT |
| `google/uuid` | v1.6.0 | UUIDs (usado como PK) |
| `joho/godotenv` | v1.5.1 | Carga de variables .env |
| `jinzhu/copier` | v0.4.0 | Copia entre structs (DTOs) |

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

1. Crear carpeta en `src/modules/newmodule/`
2. Crear archivos:
   - `newmodule_dto.go` → DTOs (input)
   - `newmodule_repo.go` → Repository (CRUD)
   - `newmodule_service.go` → Service (lógica)
   - `newmodule_handler.go` → Handler (HTTP)
3. Registrar en `src/container.go` (inyección de dependencias)
4. Registrar rutas en `cmd/api/api.go`

### Crear una Migración

Las migraciones son automáticas con GORM AutoMigrate. Solo define tu modelo en `src/model/` y agrega a `config/migrate.go`:

```go
func Migrate(db *gorm.DB) {
    err := db.AutoMigrate(
        &model.User{},
        &model.Task{},
        &model.YourNewModel{}, // ← Agregar aquí
    )
    // ...
}
```

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

Para rutas protegidas, agrega el middleware:

```go
api := app.Group("/api/v1")
api.Use(middleware.JwtMiddleware())

// Solo usuarios autenticados acceden aquí
api.Get("/protected", handler.Protected)
```

### Generar JWT

```go
import "github.com/MetaDandy/go-fiber-skeleton/helper"

token, err := helper.GenerateJwt(userID, email, role)
// Token válido por 24 horas
```

### Validar Contraseñas

```go
import "github.com/MetaDandy/go-fiber-skeleton/helper"

// Hash
hashedPassword, _ := helper.HashPassword("plainPassword")

// Verificar
isValid := helper.CheckPasswordHash("plainPassword", hashedPassword)
```

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

## 🌐 CORS

CORS está pre-configurado en `cmd/main.go`:

```go
app.Use(cors.New(cors.Config{
    AllowOrigins: os.Getenv("ALLOW_ORIGINS"), // "*" permite todos
    AllowMethods: "GET,POST,PATCH,DELETE,OPTIONS",
    AllowHeaders: "Origin, Content-Type, Accept, Authorization",
}))
```

Modifica según necesites.

---

## 📝 Logging

Logger personalizado en cada request:

```
📢 Ruta accedida: GET /api/v1/users
📢 Ruta accedida: POST /api/v1/tasks
```

Implementa tu propio logger en `middleware/logger.go` si necesitas más funcionalidad.

---

## 🧪 Testing (Future)

Este esqueleto no incluye tests aún, pero puedes agregar:

```bash
go get github.com/stretchr/testify
```

Estructura recomendada:

```
modules/user/
├── user_service.go
└── user_service_test.go  ← Tests aquí
```

---

## 🚢 Deployment (Render, Heroku, etc)

### Render.com

1. Push a GitHub
2. Crear nuevo "Web Service" en Render
3. Conectar repositorio
4. Build command: `go build -o app ./cmd`
5. Start command: `./app`
6. Agregar variables de entorno (.env)

### Railway.app

Instrucciones similares. Railway detecta Go automáticamente.

---

## 📚 Recursos Útiles

- [Fiber Docs](https://docs.gofiber.io)
- [GORM Docs](https://gorm.io)
- [JWT RFC 7519](https://tools.ietf.org/html/rfc7519)
- [PostgreSQL Docs](https://www.postgresql.org/docs)

---

## 🤝 Contribuciones

Este es un esqueleto para tu uso personal. Si lo mejoras y quieres compartir, considera hacer un PR al original.

---

## 📄 Licencia

MIT - Libre para usar en proyectos personales y comerciales.

---

## 🎯 Próximos Pasos

1. ✅ Clonar/Forkar este repositorio
2. ✅ Configurar `.env` con tu BD
3. ✅ Ejecutar `air` o `docker compose up`
4. ✅ Agregar tus módulos en `src/modules/`
5. ✅ Expandir con tu lógica de negocio

¡Feliz desarrollo! 🚀

---

**Última actualización**: Diciembre 2025
