package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TokenRepository interface {
	Save(ctx context.Context, tokenHash string, userID int, expiresAt time.Time) error
	Revoke(ctx context.Context, tokenHash string) error
	IsValid(ctx context.Context, tokenHash string) (int, bool, error)
}

type tokenPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewTokenRepository(pool *pgxpool.Pool) TokenRepository {
	return &tokenPostgresRepository{pool: pool}
}

func (r *tokenPostgresRepository) Save(ctx context.Context, tokenHash string, userID int, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		"INSERT INTO refresh_tokens (token_hash, user_id, expires_at) VALUES ($1, $2, $3)",
		tokenHash, userID, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("gagal menyimpan refresh token: %w", err)
	}
	return nil
}

func (r *tokenPostgresRepository) Revoke(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, "UPDATE refresh_tokens SET is_revoked = TRUE WHERE token_hash = $1", tokenHash)
	if err != nil {
		return fmt.Errorf("gagal mencabut refresh token: %w", err)
	}
	return nil
}

func (r *tokenPostgresRepository) IsValid(ctx context.Context, tokenHash string) (int, bool, error) {
	var userID int
	var expiresAt time.Time
	var isRevoked bool

	row := r.pool.QueryRow(ctx,
		"SELECT user_id, expires_at, is_revoked FROM refresh_tokens WHERE token_hash = $1",
		tokenHash,
	)
	err := row.Scan(&userID, &expiresAt, &isRevoked)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("gagal memeriksa token: %w", err)
	}

	if isRevoked || time.Now().After(expiresAt) {
		return 0, false, nil
	}

	return userID, true, nil
}