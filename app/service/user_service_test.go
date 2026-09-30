package service

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/config"
	"api-students/helper"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type fakeUserRepo struct {
	created model.User
	updated model.User
	deleted bool
}

func (f *fakeUserRepo) Create(ctx context.Context, u model.User) (model.User, error) {
	f.created = u
	return u, nil
}

func (f *fakeUserRepo) FindByUsername(ctx context.Context, username string) (model.User, error) {
	return model.User{}, repository.ErrNotFound
}

func (f *fakeUserRepo) FindByID(ctx context.Context, id int) (model.User, error) {
	return model.User{ID: id, Username: "alice", Email: "alice@example.com", Role: "user", IsActive: true, CreatedAt: time.Now()}, nil
}

func (f *fakeUserRepo) Update(ctx context.Context, id int, u model.User) (model.User, error) {
	f.updated = u
	return u, nil
}

func (f *fakeUserRepo) UpdateRole(ctx context.Context, id int, role string) (model.User, error) {
	return model.User{ID: id, Role: role}, nil
}

func (f *fakeUserRepo) Delete(ctx context.Context, id int) error {
	f.deleted = true
	return nil
}

type fakeStudentRepo struct {
	student model.Student
	deleted bool
}

func (f *fakeStudentRepo) FindAfterCursor(ctx context.Context, q model.CursorQuery) ([]model.Student, error) {
	return nil, nil
}

func (f *fakeStudentRepo) Create(ctx context.Context, s model.Student) (model.Student, error) {
	return s, nil
}

func (f *fakeStudentRepo) FindByID(ctx context.Context, id int) (model.Student, error) {
	return f.student, nil
}

func (f *fakeStudentRepo) Update(ctx context.Context, id int, s model.Student) (model.Student, error) {
	return s, nil
}

func (f *fakeStudentRepo) Delete(ctx context.Context, id int) error {
	f.deleted = true
	return nil
}

func TestUserServiceCreateUsesRequestPassword(t *testing.T) {
	password := "StrongPass123"
	repo := &fakeUserRepo{}
	service := NewUserService(repo, helper.NewPermissionSet(map[string][]string{"admin": {"user:update:any"}}))

	app := config.NewApp(slog.New(slog.NewTextHandler(io.Discard, nil)))
	app.Post("/users", service.Create)

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"username":"alice","email":"alice@example.com","password":"StrongPass123"}`))
	req.Header.Set("Content-Type", fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d but got %d", http.StatusCreated, resp.StatusCode)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repo.created.Password), []byte(password)); err != nil {
		t.Fatalf("password hash did not match input: %v", err)
	}
}

func TestStudentServiceDeleteRequiresPermission(t *testing.T) {
	repo := &fakeStudentRepo{student: model.Student{ID: 2, OwnerID: 11, Nim: "20240101", Name: "Budi", Grade: 88.5}}
	service := NewStudentService(repo, helper.NewPermissionSet(map[string][]string{"user": {}}))

	app := config.NewApp(slog.New(slog.NewTextHandler(io.Discard, nil)))
	app.Delete("/students/:id", func(c *fiber.Ctx) error {
		c.Locals("user", model.AuthUser{UserID: 7, Role: "user"})
		return service.Delete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/students/2", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected status %d but got %d", fiber.StatusForbidden, resp.StatusCode)
	}
	if repo.deleted {
		t.Fatal("delete should not execute for unauthorized user")
	}
}
