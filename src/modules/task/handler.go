package task

import (
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/gofiber/fiber/v2"
)

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
	return &handler{
		service: service,
	}
}

func (h *handler) RegisterRoutes(router fiber.Router) {
	tasks := router.Group("/tasks")
	tasks.Post("/", h.Create)
	tasks.Get("/:id/f", h.FindAll)
	tasks.Get("/:id", h.FindByID)
	tasks.Patch("/:id", h.Update)
	tasks.Delete("/:id", h.Delete)
}

func (h *handler) Create(c *fiber.Ctx) error {
	var input Create
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid input")
	}

	err := h.service.Create(input)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Could not create task")
	}

	return c.SendStatus(fiber.StatusCreated)
}

func (h *handler) FindAll(c *fiber.Ctx) error {
	// ? Deberia ser c.Locals pero hasta que no implemente auth sera con params
	userID := c.Params("id")
	opts := helper.NewFindAllOptionsFromQuery(c)
	finded, err := h.service.FindAll(userID, opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(finded)
}

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

func (h *handler) Update(c *fiber.Ctx) error {
	var input Update

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	id := c.Params("id")

	err := h.service.Update(id, input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *handler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.service.Delete(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)
}
