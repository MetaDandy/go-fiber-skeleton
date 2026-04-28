---
name: scaffold-module
description: "Use when: the user asks to create, generate, or scaffold a new core module. This skill creates the standard folder structure (handler, service, repo, dto), initializes dependencies in src/container.go, and registers new permissions in config/seed/permission.go."
---

# Scaffold New Core Module

This skill generates a new module following the project's standard Go Fiber + GORM architecture, sets up dependency injection, and seeds its base permissions.

## Execution Steps

When generating a new module, perform ALL of the following steps:
1. Ask the user for the name of the module (e.g., `user_permission`).
2. Create `src/core/<module>/dto.go`, `repo.go`, `service.go`, and `handler.go` following the Code Templates below.
3. Create `src/response/<module>.go` following the Code Templates below.
4. Update `src/container.go`: instantiate the Repo, Service, and Handler, then append the handler to the `Handlers` array returned by `SetupContainer()`.
5. Update `src/enum/permission.go` (if it exists) to declare the new permission enums (e.g., `enum.<Module>Create`, `enum.<Module>Read`).
6. Update `config/seed/permission.go` by appending the generated permissions to the `permissionModules` slice.

## Code Templates

### 1. Input DTO (`src/core/<module>/dto.go`)
```go
package <module>

type Create struct {
// e.g. Name string `json:"name" validate:"required"`
}

type Update struct {
// e.g. Name string `json:"name"`
}
```

### 2. Output Response (`src/response/<module>.go`)
```go
package response

import (
"time"
"github.com/MetaDandy/go-fiber-skeleton/src/model"
"github.com/jinzhu/copier"
)

type <Entity> struct {
ID        string `json:"id"`
CreatedAt string `json:"created_at"`
UpdatedAt string `json:"updated_at"`
}

func <Entity>ToDto(m *model.<Entity>) <Entity> {
var dto <Entity>
copier.Copy(&dto, m)
dto.ID = m.ID.String() // Adjust if ID is not UUID
dto.CreatedAt = m.CreatedAt.Format(time.RFC3339)
dto.UpdatedAt = m.UpdatedAt.Format(time.RFC3339)
return dto
}

func <Entity>ToListDto(m []model.<Entity>) []<Entity> {
out := make([]<Entity>, len(m))
for i := range m {
out[i] = <Entity>ToDto(&m[i])
}
return out
}
```

### 3. Repository (`src/core/<module>/repo.go`)
```go
package <module>

import (
"github.com/MetaDandy/go-fiber-skeleton/src/model"
"gorm.io/gorm"
)

type Repo interface {
BeginTx() *gorm.DB
Create(m model.<Entity>) error
FindByID(id string) (model.<Entity>, error)
}

type repo struct {
db *gorm.DB
}

func NewRepo(db *gorm.DB) Repo {
return &repo{db: db}
}

func (r *repo) BeginTx() *gorm.DB {
return r.db.Begin()
}

func (r *repo) Create(m model.<Entity>) error {
return r.db.Create(&m).Error
}

func (r *repo) FindByID(id string) (model.<Entity>, error) {
var m model.<Entity>
err := r.db.First(&m, "id = ?", id).Error
return m, err
}
```

### 4. Service (`src/core/<module>/service.go`)
```go
package <module>

import (
"github.com/MetaDandy/go-fiber-skeleton/src/response"
)

type Service interface {
Create(input Create) error
FindByID(id string) (*response.<Entity>, error)
}

type service struct {
repo Repo
}

func NewService(repo Repo) Service {
return &service{repo: repo}
}
// Implement logic methods here
```

### 5. Handler (`src/core/<module>/handler.go`)
```go
package <module>

import (
"github.com/gofiber/fiber/v3"
)

type Handler struct {
service Service
}

func NewHandler(service Service) *Handler {
return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
// e.g. router.Post("/<module>", h.Create)
}
// Implement Fiber endpoints here
```
