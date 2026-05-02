package user

import (
	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/gofiber/fiber/v3"
)

type Handler interface {
	RegisterRoutes(router fiber.Router)
	Create(c fiber.Ctx) error
	FindAll(c fiber.Ctx) error
	FindByID(c fiber.Ctx) error
	Update(c fiber.Ctx) error
	Delete(c fiber.Ctx) error
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
	users := router.Group("/users", h.jwtMiddle)
	users.Post("/", h.Create)
	users.Get("/", h.FindAll)
	users.Get("/:id", h.FindByID)
	users.Patch("/:id", h.Update)
	users.Delete("/:id", h.Delete)
}

func (h *handler) Create(c fiber.Ctx) error {
	input := new(Create)
	if err := c.Bind().Body(&input); err != nil {
		return api_error.BadRequest("Invalid input")
	}

	if err := h.service.Create(*input); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusCreated)
}

func (h *handler) FindAll(c fiber.Ctx) error {
	opts := helper.NewFindAllOptionsFromQuery(c)
	finded, err := h.service.FindAll(opts)
	if err != nil {
		return err
	}
	return c.JSON(finded)
}

func (h *handler) FindByID(c fiber.Ctx) error {
	id := c.Params("id")
	finded, err := h.service.FindByID(id)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(finded)
}

func (h *handler) Update(c fiber.Ctx) error {
	var input Update

	if err := c.Bind().Body(&input); err != nil {
		return api_error.BadRequest("Invalid request body")
	}

	id := c.Params("id")

	if err := h.service.Update(id, input); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *handler) Delete(c fiber.Ctx) error {
	id := c.Params("id")

	if err := h.service.Delete(id); err != nil {
		return err
	}
	return c.Status(fiber.StatusNoContent).Send(nil)
}
