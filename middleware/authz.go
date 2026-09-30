package middleware

import (
	"api-students/helper"

	"github.com/gofiber/fiber/v2"
)

func RequirePermission(perms *helper.PermissionSet, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		current, ok := helper.CurrentUser(c)
		if !ok {
			return helper.Unauthorized("belum terautentikasi")
		}
		if !perms.Can(current.Role, permission) {
			return helper.Forbidden("tidak memiliki izin untuk melakukan aksi ini")
		}
		return c.Next()
	}
}

func RequireRole(roles ...string) fiber.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		current, ok := helper.CurrentUser(c)
		if !ok {
			return helper.Unauthorized("belum terautentikasi")
		}
		if _, ok := allowed[current.Role]; !ok {
			return helper.Forbidden("role anda tidak diizinkan")
		}
		return c.Next()
	}
}
