package middleware

import (
	"log/slog"

	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/gofiber/fiber/v3"
)

// ErrorHandler es el middleware global que captura y formatea errores
func ErrorHandler(c fiber.Ctx) error {
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

		// Retornar la respuesta formateada
		return c.Status(apiErr.Status).JSON(apiErr.ToResponse())
	}

	// Si es un error de Fiber (ej: route not found)
	if fiberErr, ok := err.(*fiber.Error); ok {
		switch fiberErr.Code {
		case fiber.StatusNotFound:
			return c.Status(fiber.StatusNotFound).JSON(
				api_error.NotFound("Ruta no encontrada").ToResponse(),
			)
		case fiber.StatusMethodNotAllowed:
			return c.Status(fiber.StatusMethodNotAllowed).JSON(
				api_error.Forbidden("Método no permitido").ToResponse(),
			)
		default:
			return c.Status(fiberErr.Code).JSON(
				api_error.InternalServerError("Error interno del servidor").ToResponse(),
			)
		}
	}

	// Cualquier otro error no identificado
	slog.Error("Unhandled error", "error", err.Error())
	return c.Status(fiber.StatusInternalServerError).JSON(
		api_error.InternalServerError("Error interno del servidor").ToResponse(),
	)
}
