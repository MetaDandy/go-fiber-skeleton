package user

import (
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/gofiber/fiber/v2"
)

type UserHandler interface {
	RegisterRoutes(router fiber.Router)
	Create(c *fiber.Ctx) error
	FindAll(c *fiber.Ctx) error
	FindByID(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
}

type Handler struct {
	service UserService
}

func NewUserHandler(service UserService) UserHandler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	users := router.Group("/users")
	users.Post("/", h.Create)
	users.Get("/", h.FindAll)
	users.Get("/:id", h.FindByID)
	users.Put("/:id", h.Update)
	users.Delete("/:id", h.Delete)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var input CreateUserDto
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid input")
	}

	user, err := h.service.Create(input)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Could not create user")
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

func (h *Handler) FindAll(c *fiber.Ctx) error {
	opts := helper.NewFindAllOptionsFromQuery(c)
	finded, err := h.service.FindAll(opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(finded)
}

func (h *Handler) FindByID(c *fiber.Ctx) error {
	id := c.Params("id")
	finded, err := h.service.FindByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(finded)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	var input UpdateUserDto

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	id := c.Params("id")

	updated, err := h.service.Update(id, input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(updated)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.service.Delete(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}
