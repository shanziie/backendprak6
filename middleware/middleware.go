package middleware

import (
	"errors"
	"log/slog"
	"time"

	"api-students/helper"

	"github.com/google/uuid"
	"github.com/gofiber/fiber/v2"
)

func RequestLogger(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Locals("requestid", requestID)
		c.Set("X-Request-ID", requestID)

		start := time.Now()
		err := c.Next()

		status := c.Response().StatusCode()
		if err != nil {
			var appErr *helper.AppError
			if errors.As(err, &appErr) {
				status = appErr.Status
			} else {
				status = fiber.StatusInternalServerError
			}
		}

		attrs := []any{
			slog.String("request_id", requestID),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", status),
			slog.Duration("duration", time.Since(start)),
			slog.String("ip", c.IP()),
		}

		if user, ok := helper.CurrentUser(c); ok {
			attrs = append(attrs,
				slog.Int("user_id", user.UserID),
				slog.String("role", user.Role))
		}

		logger.Info("http_request", attrs...)
		return err
	}
}