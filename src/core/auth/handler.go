package authentication

import (
	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/gofiber/fiber/v3"
)

type Handler interface {
	RegisterRoutes(router fiber.Router)
	UserAuthProviders(c fiber.Ctx) error
	SignUpPassword(c fiber.Ctx) error
	SendTestEmail(c fiber.Ctx) error
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
	auth.Post("/signup", h.SignUpPassword)
	auth.Post("/send-test-email", h.SendTestEmail)
}

func (h *handler) UserAuthProviders(c fiber.Ctx) error {
	email := c.Params("email")
	if email == "" {
		return api_error.BadRequest("Email parameter is required")
	}

	providers, err := h.service.UserAuthProviders(email)
	if err != nil {
		// Si el error es un *api_error.Error, retornarlo directamente
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		// Si es otro tipo de error, envolverlo
		return api_error.InternalServerError("Could not retrieve authentication providers").WithErr(err)
	}

	return c.JSON(fiber.Map{
		"providers": providers,
	})
}

func (h *handler) SignUpPassword(c fiber.Ctx) error {
	var input SignUpPassword
	if err := c.Bind().Body(&input); err != nil {
		return api_error.BadRequest("Invalid request body")
	}

	if err := input.Validate(); err != nil {
		return err
	}

	// Si el error es un *api_error.Error, retornarlo directamente
	if err := h.service.SignUpPassword(input); err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		// Si es otro tipo de error, envolverlo
		return api_error.InternalServerError(err.Error()).WithErr(err)
	}

	return c.JSON(fiber.Map{
		"message": "user created successfully",
	})
}

func (h *handler) SendTestEmail(c fiber.Ctx) error {
	var input struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := c.Bind().Body(&input); err != nil {
		return api_error.BadRequest("Invalid request body")
	}

	if input.Email == "" {
		return api_error.BadRequest("Email is required")
	}

	if err := h.service.SendTestEmail(input.Email, input.Name); err != nil {
		if apiErr, ok := err.(*api_error.Error); ok {
			return apiErr
		}
		return api_error.InternalServerError("Failed to send test email").WithErr(err)
	}

	return c.JSON(fiber.Map{
		"message": "test email sent successfully",
		"email":   input.Email,
	})
}
