package middleware

import (
	"strings"
	"sync"
	"time"

	"api-students/helper"

	"github.com/gofiber/fiber/v2"
)

func RequireAuth(jwtManager *helper.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get(fiber.HeaderAuthorization)
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return helper.Unauthorized("header Authorization tidak ada atau salah bentuk")
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		user, err := jwtManager.VerifyAccessToken(tokenStr)
		if err != nil {
			return helper.Unauthorized("access token tidak valid")
		}

		c.Locals("user", user)
		return c.Next()
	}
}

func RequireJSON() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() == fiber.MethodPost || c.Method() == fiber.MethodPut || c.Method() == fiber.MethodPatch {
			ct := c.Get(fiber.HeaderContentType)
			if !strings.HasPrefix(ct, "application/json") {
				return helper.UnsupportedMediaType("Content-Type harus application/json")
			}
		}
		return c.Next()
	}
}

func LoginRateLimiter() fiber.Handler {
	var (
		mu       sync.Mutex
		attempts = make(map[string][]time.Time)
	)

	return func(c *fiber.Ctx) error {
		ip := c.IP()
		now := time.Now()

		mu.Lock()
		defer mu.Unlock()

		window := time.Minute
		maxAttempts := 5

		var validAttempts []time.Time
		for _, t := range attempts[ip] {
			if now.Sub(t) < window {
				validAttempts = append(validAttempts, t)
			}
		}

		if len(validAttempts) >= maxAttempts {
			return helper.TooManyRequests("terlalu banyak percobaan login, coba lagi dalam satu menit")
		}

		attempts[ip] = append(validAttempts, now)
		return c.Next()
	}
}