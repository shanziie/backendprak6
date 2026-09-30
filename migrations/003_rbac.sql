CREATE TABLE IF NOT EXISTS roles (
    name        VARCHAR(20)  PRIMARY KEY,
    description VARCHAR(150) NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
 
INSERT INTO roles (name, description) VALUES
    ('admin', 'Akses penuh terhadap seluruh data dan pengaturan'),
    ('staff', 'Boleh melihat data seluruh user, tetapi tidak boleh mengubah'),
    ('user',  'Hanya boleh mengelola datanya sendiri')
ON CONFLICT (name) DO NOTHING;
 
CREATE TABLE IF NOT EXISTS permissions (
    name        VARCHAR(50)  PRIMARY KEY,
    description VARCHAR(150) NOT NULL
);
 
INSERT INTO permissions (name, description) VALUES
    ('user:list',       'Melihat daftar seluruh user'),
    ('user:read:any',   'Melihat data user mana pun'),
    ('user:update:any', 'Mengubah data user mana pun'),
    ('user:delete',     'Menghapus user'),
    ('role:assign',     'Mengubah role milik user lain')
ON CONFLICT (name) DO NOTHING;
 
CREATE TABLE IF NOT EXISTS role_permissions (
    role_name       VARCHAR(20) NOT NULL
        REFERENCES roles(name)       ON DELETE CASCADE,
    permission_name VARCHAR(50) NOT NULL
        REFERENCES permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role_name, permission_name)
);
 
INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'user:list'),
    ('admin', 'user:read:any'),
    ('admin', 'user:update:any'),
    ('admin', 'user:delete'),
    ('admin', 'role:assign'),
    ('staff', 'user:list'),
    ('staff', 'user:read:any')
ON CONFLICT DO NOTHING;
 
UPDATE users SET role = 'user' WHERE role NOT IN (SELECT name FROM roles);
 
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_fkey;
ALTER TABLE users
    ADD CONSTRAINT users_role_fkey
    FOREIGN KEY (role) REFERENCES roles(name) ON UPDATE CASCADE;
 
CREATE INDEX IF NOT EXISTS users_role_idx ON users (role);
```[cite: 2]

#### 4. `migrations/004_student_permissions.sql`
```sql
INSERT INTO permissions (name, description) VALUES
    ('student:list',       'Melihat daftar seluruh mahasiswa'),
    ('student:read:any',   'Melihat data mahasiswa mana pun'),
    ('student:create',     'Menambahkan data mahasiswa'),
    ('student:update:any', 'Mengubah data mahasiswa mana pun'),
    ('student:delete',     'Menghapus data mahasiswa')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'student:list'),
    ('admin', 'student:read:any'),
    ('admin', 'student:create'),
    ('admin', 'student:update:any'),
    ('admin', 'student:delete'),
    ('staff', 'student:list'),
    ('staff', 'student:read:any'),
    ('staff', 'student:create')
ON CONFLICT DO NOTHING;

ALTER TABLE students ADD COLUMN IF NOT EXISTS owner_id INT;

UPDATE students SET owner_id = (SELECT id FROM users LIMIT 1) WHERE owner_id IS NULL;

ALTER TABLE students ALTER COLUMN owner_id SET NOT NULL;

ALTER TABLE students DROP CONSTRAINT IF EXISTS students_owner_id_fkey;
ALTER TABLE students
    ADD CONSTRAINT students_owner_id_fkey
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS students_owner_idx ON students (owner_id);
```[cite: 2]

#### 5. `migrations/005_cursor_index.sql`
```sql
CREATE INDEX IF NOT EXISTS users_created_at_id_desc_idx
    ON users (created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS students_created_at_id_desc_idx
    ON students (created_at DESC, id DESC);
```[cite: 14]

---

### III. Folder `route`

#### 1. `route/route.go`
```go
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

	// --- publik ---
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// --- autentikasi ---
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	// --- users ---
	users := api.Group("/users",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))

	perms := deps.Permissions
	users.Get("/", middleware.RequirePermission(perms, "user:list"), deps.UserService.Get)
	users.Post("/", middleware.RequirePermission(perms, "user:update:any"), deps.UserService.Create)
	users.Delete("/:id", middleware.RequirePermission(perms, "user:delete"), deps.UserService.Delete)
	users.Patch("/:id/role", middleware.RequirePermission(perms, "role:assign"), deps.UserService.AssignRole)

	users.Get("/:id", deps.UserService.Get)
	users.Put("/:id", deps.UserService.Replace)
	users.Patch("/:id", deps.UserService.Patch)

	// --- students ---
	students := api.Group("/students",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))

	students.Get("/", middleware.RequirePermission(perms, "student:list"), deps.StudentService.List)
	students.Post("/", middleware.RequirePermission(perms, "student:create"), deps.StudentService.Create)
	students.Get("/:id", deps.StudentService.Get)
	students.Put("/:id", deps.StudentService.Update)
	students.Delete("/:id", middleware.RequirePermission(perms, "student:delete"), deps.StudentService.Delete)
}