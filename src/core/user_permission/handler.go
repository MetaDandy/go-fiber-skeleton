package user_permission

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
	userPermissionGroup := router.Group("/user-permission")
	// PATCH /user-permission/:user_id specifies updating permissions (add/remove arrays)
	userPermissionGroup.Patch("/:user_id", h.UpdateDetails)
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
