package repository

import (
	"context"
	"errors"
	"fmt"

	"api-students/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrStudentNotFound = errors.New("student tidak ditemukan")

type StudentRepository interface {
	FindAfterCursor(ctx context.Context, q model.CursorQuery) ([]model.Student, error)
	Create(ctx context.Context, s model.Student) (model.Student, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	Update(ctx context.Context, id int, s model.Student) (model.Student, error)
	Delete(ctx context.Context, id int) error
}

type studentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentPostgresRepository{pool: pool}
}

func (r *studentPostgresRepository) FindAfterCursor(
	ctx context.Context, q model.CursorQuery,
) ([]model.Student, error) {
	args := []any{}
	where := " WHERE 1 = 1"

	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		where += fmt.Sprintf(" AND name ILIKE $%d", len(args))
	}
	if q.After != nil {
		args = append(args, q.After.CreatedAt, q.After.ID)
		where += fmt.Sprintf(" AND (created_at, id) < ($%d, $%d)", len(args)-1, len(args))
	}

	args = append(args, q.Limit+1)
	query := fmt.Sprintf(
		"SELECT id, nim, name, grade, owner_id, created_at FROM students%s ORDER BY created_at DESC, id DESC LIMIT $%d",
		where, len(args),
	)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar student: %w", err)
	}
	defer rows.Close()

	result := []model.Student{}
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.Nim, &s.Name, &s.Grade, &s.OwnerID, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("membaca row student: %w", err)
		}
		result = append(result, s)
	}
	return result, nil
}

func (r *studentPostgresRepository) Create(ctx context.Context, s model.Student) (model.Student, error) {
	query := `INSERT INTO students (nim, name, grade, owner_id) 
	          VALUES ($1, $2, $3, $4) 
	          RETURNING id, nim, name, grade, owner_id, created_at`

	row := r.pool.QueryRow(ctx, query, s.Nim, s.Name, s.Grade, s.OwnerID)
	var created model.Student
	err := row.Scan(&created.ID, &created.Nim, &created.Name, &created.Grade, &created.OwnerID, &created.CreatedAt)
	if err != nil {
		return model.Student{}, fmt.Errorf("gagal membuat student: %w", err)
	}
	return created, nil
}

func (r *studentPostgresRepository) FindByID(ctx context.Context, id int) (model.Student, error) {
	query := `SELECT id, nim, name, grade, owner_id, created_at FROM students WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)

	var s model.Student
	err := row.Scan(&s.ID, &s.Nim, &s.Name, &s.Grade, &s.OwnerID, &s.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrStudentNotFound
		}
		return model.Student{}, fmt.Errorf("gagal mencari student by id: %w", err)
	}
	return s, nil
}

func (r *studentPostgresRepository) Update(ctx context.Context, id int, s model.Student) (model.Student, error) {
	query := `UPDATE students SET name = $1, grade = $2 WHERE id = $3 
	          RETURNING id, nim, name, grade, owner_id, created_at`

	row := r.pool.QueryRow(ctx, query, s.Name, s.Grade, id)
	var updated model.Student
	err := row.Scan(&updated.ID, &updated.Nim, &updated.Name, &updated.Grade, &updated.OwnerID, &updated.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrStudentNotFound
		}
		return model.Student{}, fmt.Errorf("gagal update student: %w", err)
	}
	return updated, nil
}

func (r *studentPostgresRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM students WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("gagal delete student: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrStudentNotFound
	}
	return nil
}
