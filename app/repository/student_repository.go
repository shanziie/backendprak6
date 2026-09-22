package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
)

var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("data duplikat")
)

type StudentRepository interface {
	FindAll(ctx context.Context, q model.ListQuery) ([]model.Student, int, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	Create(ctx context.Context, s model.Student) (model.Student, error)
	Update(ctx context.Context, s model.Student) (model.Student, error)
	Delete(ctx context.Context, id int) error
}

type studentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentPostgresRepository{pool: pool}
}

const studentColumns = "id, owner_id, nim, name, grade, is_active"

func scanStudent(row pgx.Row) (model.Student, error) {
	var s model.Student
	err := row.Scan(&s.ID, &s.OwnerID, &s.NIM, &s.Name, &s.Grade, &s.IsActive)
	return s, err
}

func (r *studentPostgresRepository) FindByID(ctx context.Context, id int) (model.Student, error) {
	query := `SELECT ` + studentColumns + ` FROM students WHERE id = $1`
	student, err := scanStudent(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, err
	}
	return student, nil
}

func (r *studentPostgresRepository) Create(ctx context.Context, s model.Student) (model.Student, error) {
	query := `INSERT INTO students (owner_id, nim, name, grade, is_active)
	          VALUES ($1, $2, $3, $4, $5)
	          RETURNING ` + studentColumns

	created, err := scanStudent(r.pool.QueryRow(ctx, query, s.OwnerID, s.NIM, s.Name, s.Grade, s.IsActive))
	if err != nil {
		return model.Student{}, fmt.Errorf("gagal membuat student: %w", err)
	}
	return created, nil
}

func (r *studentPostgresRepository) Update(ctx context.Context, s model.Student) (model.Student, error) {
	query := `UPDATE students 
	          SET nim = $1, name = $2, grade = $3, is_active = $4
	          WHERE id = $5
	          RETURNING ` + studentColumns

	updated, err := scanStudent(r.pool.QueryRow(ctx, query, s.NIM, s.Name, s.Grade, s.IsActive, s.ID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, err
	}
	return updated, nil
}

func (r *studentPostgresRepository) Delete(ctx context.Context, id int) error {
	cmd, err := r.pool.Exec(ctx, "DELETE FROM students WHERE id = $1", id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *studentPostgresRepository) FindAll(ctx context.Context, q model.ListQuery) ([]model.Student, int, error) {
	rows, err := r.pool.Query(ctx, "SELECT "+studentColumns+" FROM students ORDER BY id ASC")
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var students []model.Student
	for rows.Next() {
		s, err := scanStudent(rows)
		if err != nil {
			return nil, 0, err
		}
		students = append(students, s)
	}
	return students, len(students), nil
}