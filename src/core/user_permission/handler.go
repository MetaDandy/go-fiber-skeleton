package user_permission

import (
	"github.com/MetaDandy/go-fiber-skeleton/middleware"
	"github.com/MetaDandy/go-fiber-skeleton/src/enum"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	service   Service
	jwtMiddle fiber.Handler
}

func NewHandler(service Service, jwtMiddle fiber.Handler) *Handler {
	return &Handler{
		service:   service,
		jwtMiddle: jwtMiddle,
	}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	userPermissionGroup := router.Group("/user-permission", h.jwtMiddle)
	// PATCH /user-permission/:user_id specifies updating permissions (add/remove arrays)
	userPermissionGroup.Patch(
		"/:user_id",
		middleware.RequirePermission(enum.UserPermissionCreate.String()), // <--- ¡AQUÍ ESTÁ LA MÁGIA!
		h.UpdateDetails,
	)
}

func (h *Handler) UpdateDetails(c fiber.Ctx) error {
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
