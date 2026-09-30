package helper

import (
	"api-students/app/model"

	"github.com/gofiber/fiber/v2"
)

func Success(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SuccessCursor(c *fiber.Ctx, message string, data any, meta *model.CursorMeta) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
		"meta":    meta,
	})
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}