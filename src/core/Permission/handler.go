package permission

import (
	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/gofiber/fiber/v3"
)

type Handler interface {
	RegisterRoutes(router fiber.Router)
	FindAll(c fiber.Ctx) error
	FindByID(c fiber.Ctx) error
}

type handler struct {
	service   Service
	jwtMiddle fiber.Handler
}

func NewHandler(service Service, jwtMiddle fiber.Handler) Handler {
	return &handler{
		service:   service,
		jwtMiddle: jwtMiddle,
	}
}

func (h *handler) RegisterRoutes(router fiber.Router) {
	permissions := router.Group("/permissions", h.jwtMiddle)
	permissions.Get("/", h.FindAll)
	permissions.Get("/:id", h.FindByID)
}

func (h *handler) FindAll(c fiber.Ctx) error {
	opts := helper.NewFindAllOptionsFromQuery(c)
	finded, err := h.service.FindAll(opts)
	if err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		return api_error.InternalServerError("Could not retrieve permissions").WithErr(err)
	}
	return c.JSON(finded)
}

func (h *handler) FindByID(c fiber.Ctx) error {
	id := c.Params("id")
	finded, err := h.service.FindByID(id)
	if err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		return api_error.NotFound("Permission not found").WithErr(err)
	}

	return c.Status(fiber.StatusOK).JSON(finded)
}
