package role

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
	UpdateHeader(c fiber.Ctx) error
	UpdateDetails(c fiber.Ctx) error
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
	roles := router.Group("/roles")
	roles.Post("/", h.Create)
	roles.Get("/", h.FindAll)
	roles.Get("/:id", h.FindByID)
	roles.Patch("/:id/header", h.UpdateHeader)
	roles.Patch("/:id/details", h.UpdateDetails)
}

func (h *handler) Create(c fiber.Ctx) error {
	var dto Create

	if err := c.Bind().Body(&dto); err != nil {
		return api_error.BadRequest("Invalid body").WithErr(err)
	}

	if err := h.service.Create(dto); err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		return api_error.InternalServerError("Could not create role").WithErr(err)
	}

	return c.SendStatus(fiber.StatusCreated)
}

func (h *handler) FindAll(c fiber.Ctx) error {
	opts := helper.NewFindAllOptionsFromQuery(c)

	finded, err := h.service.FindAll(opts)
	if err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		return api_error.InternalServerError("Could not retrieve roles").WithErr(err)
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
		return api_error.NotFound("Role not found").WithErr(err)
	}

	return c.Status(fiber.StatusOK).JSON(finded)
}

func (h *handler) UpdateHeader(c fiber.Ctx) error {
	var input UpdateHeader

	if err := c.Bind().Body(&input); err != nil {
		return api_error.BadRequest("Invalid request body").WithErr(err)
	}

	id := c.Params("id")

	if err := h.service.UpdateHeader(id, input); err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		return api_error.InternalServerError("Could not update role header").WithErr(err)
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *handler) UpdateDetails(c fiber.Ctx) error {
	var input UpdateDetails

	if err := c.Bind().Body(&input); err != nil {
		return api_error.BadRequest("Invalid request body").WithErr(err)
	}

	id := c.Params("id")

	if err := h.service.UpdateDetails(id, input); err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		return api_error.InternalServerError("Could not update role details").WithErr(err)
	}

	return c.SendStatus(fiber.StatusOK)
}
