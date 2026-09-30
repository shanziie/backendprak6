package service

import (
	"encoding/csv"
	"errors"
	"strconv"
	"strings"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"

	"github.com/gofiber/fiber/v2"
)

type StudentService struct {
	repo  repository.StudentRepository
	perms *helper.PermissionSet
}

func NewStudentService(repo repository.StudentRepository, perms *helper.PermissionSet) *StudentService {
	return &StudentService{repo: repo, perms: perms}
}

func (s *StudentService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	format, err := helper.Negotiate(c, helper.FormatJSON, helper.FormatCSV)
	if err != nil {
		return err
	}

	q, err := helper.ParseCursorQuery(c)
	if err != nil {
		return err
	}

	students, err := s.repo.FindAfterCursor(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	if format == helper.FormatCSV {
		return s.WriteStudentsCSV(c, students)
	}

	hasMore := len(students) > q.Limit
	if hasMore {
		students = students[:q.Limit]
	}

	meta := &model.CursorMeta{Limit: q.Limit, HasMore: hasMore}
	if hasMore && len(students) > 0 {
		last := students[len(students)-1]
		meta.NextCursor = helper.EncodeCursor(last.CreatedAt, last.ID)
	}

	return helper.SuccessCursor(c, "daftar mahasiswa berhasil diambil", students, meta)
}

func (s *StudentService) WriteStudentsCSV(c *fiber.Ctx, students []model.Student) error {
	c.Set(fiber.HeaderContentType, helper.FormatCSV+"; charset=utf-8")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="students.csv"`)

	var buffer strings.Builder
	writer := csv.NewWriter(&buffer)

	header := []string{"id", "nim", "name", "grade", "owner_id", "created_at"}
	if err := writer.Write(header); err != nil {
		return helper.Internal(err)
	}

	for _, st := range students {
		row := []string{
			strconv.Itoa(st.ID), st.Nim, st.Name,
			strconv.FormatFloat(st.Grade, 'f', 2, 64),
			strconv.Itoa(st.OwnerID),
			st.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		}
		if err := writer.Write(row); err != nil {
			return helper.Internal(err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return helper.Internal(err)
	}

	return c.SendString(buffer.String())
}

func (s *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	student := model.Student{
		Nim:     strings.TrimSpace(req.Nim),
		Name:    strings.TrimSpace(req.Name),
		Grade:   req.Grade,
		OwnerID: current.UserID,
	}

	created, err := s.repo.Create(ctx, student)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusCreated, "mahasiswa berhasil didaftarkan", created)
}

func (s *StudentService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrStudentNotFound) {
			return helper.NotFound("mahasiswa tidak ditemukan")
		}
		return helper.Internal(err)
	}

	if !CanAccessStudent(current, student.OwnerID, s.perms, "student:read:any") {
		return helper.Forbidden("tidak berhak mengakses data mahasiswa ini")
	}

	return helper.Success(c, fiber.StatusOK, "mahasiswa ditemukan", student)
}

func (s *StudentService) Update(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrStudentNotFound) {
			return helper.NotFound("mahasiswa tidak ditemukan")
		}
		return helper.Internal(err)
	}

	if !CanAccessStudent(current, student.OwnerID, s.perms, "student:update:any") {
		return helper.Forbidden("tidak berhak mengubah data mahasiswa ini")
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	if IsEmptyStudentPatch(req) {
		return helper.BadRequest("body tidak boleh kosong")
	}

	updatedStudent := ApplyStudentPatch(student, req)
	result, err := s.repo.Update(ctx, id, updatedStudent)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "data mahasiswa berhasil diubah", result)
}

func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	if current.UserID == id {
		
	}

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrStudentNotFound) {
			return helper.NotFound("mahasiswa tidak ditemukan")
		}
		return helper.Internal(err)
	}

	if !CanAccessStudent(current, student.OwnerID, s.perms, "student:delete") {
		return helper.Forbidden("tidak berhak menghapus data mahasiswa ini")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrStudentNotFound) {
			return helper.NotFound("mahasiswa tidak ditemukan")
		}
		return helper.Internal(err)
	}

	return helper.NoContent(c)
}
