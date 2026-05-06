package user_permission

import (
	"github.com/MetaDandy/go-fiber-skeleton/middleware"
	"github.com/MetaDandy/go-fiber-skeleton/src/enum"
	"github.com/gofiber/fiber/v3"
)

// Handler defines the interface for user permission HTTP handlers.
// It provides endpoints for managing user-permission assignments.
type Handler interface {
	RegisterRoutes(router fiber.Router)

	// UpdateDetails handles PATCH /user-permission/:user_id
	// Requires permission: user_permission.create
	// Request body: UpdateDetails (Add []string, Remove []string)
	// Response: 200 OK with success message, or error with appropriate status code
	UpdateDetails(c fiber.Ctx) error
}

type handler struct {
	service   Service
	jwtMiddle fiber.Handler
}

// NewHandler creates a new Handler instance.
//
// Parameters:
//   - service: The user permission service for business logic operations
//   - jwtMiddle: Fiber middleware for JWT authentication
//
// Returns:
//   - Handler: The configured handler implementing the Handler interface
func NewHandler(service Service, jwtMiddle fiber.Handler) Handler {
	return &handler{
		service:   service,
		jwtMiddle: jwtMiddle,
	}
}

func (h *handler) RegisterRoutes(router fiber.Router) {
	userPermissionGroup := router.Group("/user-permission", h.jwtMiddle)
	// PATCH /user-permission/:user_id specifies updating permissions (add/remove arrays)
	userPermissionGroup.Patch(
		"/:user_id",
		middleware.RequirePermission(enum.UserPermissionCreate.String()), // <--- ¡AQUÍ ESTÁ LA MÁGIA!
		h.UpdateDetails,
	)
}

func (h *handler) UpdateDetails(c fiber.Ctx) error {
	userID := c.Params("user_id")

	var req UpdateDetails
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	if err := h.service.UpdateDetails(userID, req); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"message": "User permissions updated successfully",
	})
}
