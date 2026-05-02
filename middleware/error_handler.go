package middleware

import (
	"log/slog"

	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/gofiber/fiber/v3"
)

// ErrorHandler es el middleware global que captura y formatea errores
func ErrorHandler(c fiber.Ctx) error {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Panic recovered", "error", r)
			_ = c.Status(fiber.StatusInternalServerError).JSON(api_error.InternalServerError("Internal Server Error"))
		}
	}()

	// Proceder con el siguiente handler
	err := c.Next()

	if err == nil {
		return nil
	}

	// Intentar convertir a *api_error.Error
	if apiErr, ok := err.(*api_error.Error); ok {
		// Si es un error interno, loguear los detalles
		if apiErr.Err != nil {
			slog.Error("API error with internal cause",
				"code", apiErr.Code.String(),
				"message", apiErr.Message,
				"status", apiErr.Status,
				"error", apiErr.Err.Error(),
			)
		}

		// Retornar la respuesta formateada directamente (el struct ya tiene los tags json correctos)
		return c.Status(apiErr.Status).JSON(apiErr)
	}

	// Si es un error de Fiber (ej: route not found)
	if fiberErr, ok := err.(*fiber.Error); ok {
		var apiErr *api_error.Error
		switch fiberErr.Code {
		case fiber.StatusNotFound:
			apiErr = api_error.NotFound("Resource not found")
		case fiber.StatusMethodNotAllowed:
			apiErr = api_error.Forbidden("Method not allowed")
		case fiber.StatusUnauthorized:
			apiErr = api_error.Unauthorized("Unauthorized")
		default:
			apiErr = api_error.InternalServerError(fiberErr.Message)
		}
		return c.Status(fiberErr.Code).JSON(apiErr)
	}

	// Cualquier otro error no identificado
	slog.Error("Unhandled error", "error", err.Error())
	return c.Status(fiber.StatusInternalServerError).JSON(
		api_error.InternalServerError("Internal Server Error"),
	)
}
