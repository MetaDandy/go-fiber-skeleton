package authentication

import "github.com/gofiber/fiber/v2"

type Handler interface {
	RegisterRoutes(router fiber.Router)
	UserAuthProviders(c *fiber.Ctx) error
}

type handler struct {
	service Service
}

func NewHandler(service Service) Handler {
	return &handler{
		service: service,
	}
}

func (h *handler) RegisterRoutes(router fiber.Router) {
	auth := router.Group("/auth")
	auth.Get("/providers/:email", h.UserAuthProviders)
}

func (h *handler) UserAuthProviders(c *fiber.Ctx) error {
	email := c.Params("email")
	if email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email es requerido",
		})
	}

	providers, err := h.service.UserAuthProviders(email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error al obtener proveedores de autenticación",
		})
	}

	return c.JSON(fiber.Map{
		"providers": providers,
	})
}
