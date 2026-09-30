package route

import (
	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Dependencies struct {
	Pool           *pgxpool.Pool
	JWT            *helper.JWTManager
	Permissions    *helper.PermissionSet
	UserService    *service.UserService
	AuthService    *service.AuthService
	StudentService *service.StudentService
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	// publik 
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// autentikasi 
	auth := api.Group("/auth", middleware.RequireJSON())
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	// users 
	users := api.Group("/users",
		middleware.RequireJSON(),
		middleware.RequireAuth(deps.JWT))

	perms := deps.Permissions
	users.Get("/", middleware.RequirePermission(perms, "user:list"), deps.UserService.Get)
	users.Post("/", middleware.RequirePermission(perms, "user:update:any"), deps.UserService.Create)
	users.Delete("/:id", middleware.RequirePermission(perms, "user:delete"), deps.UserService.Delete)
	users.Patch("/:id/role", middleware.RequirePermission(perms, "role:assign"), deps.UserService.AssignRole)

	users.Get("/:id", deps.UserService.Get)
	users.Put("/:id", deps.UserService.Replace)
	users.Patch("/:id", deps.UserService.Patch)

	// students 
	students := api.Group("/students",
		middleware.RequireJSON(),
		middleware.RequireAuth(deps.JWT))

	students.Get("/", middleware.RequirePermission(perms, "student:list"), deps.StudentService.List)
	students.Post("/", middleware.RequirePermission(perms, "student:create"), deps.StudentService.Create)
	students.Get("/:id", deps.StudentService.Get)
	students.Put("/:id", deps.StudentService.Update)
	students.Delete("/:id", middleware.RequirePermission(perms, "student:delete"), deps.StudentService.Delete)
}
