package presenter

import "github.com/gofiber/fiber/v2"

type ErrorResponse struct {
	Message string `json:"message"`
}

func JSON(c *fiber.Ctx, status int, v any) error {
	return c.Status(status).JSON(v)
}

func Error(c *fiber.Ctx, status int, message string) error {
	return JSON(c, status, ErrorResponse{Message: message})
}
