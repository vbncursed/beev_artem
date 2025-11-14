package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artem13815/hr/pkg/auth"
)

// UserRepository implements auth.UserRepository backed by PostgreSQL (pgx).
type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) (*UserRepository, error) {
	repo := &UserRepository{pool: pool}
	if err := repo.ensureSchema(context.Background()); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *UserRepository) ensureSchema(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			is_admin BOOLEAN NOT NULL DEFAULT FALSE
		);
		-- backfill for older schemas
		ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN NOT NULL DEFAULT FALSE;
	`)
	return err
}

func (r *UserRepository) Create(ctx context.Context, user auth.User) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, created_at, is_admin)
		VALUES ($1, $2, $3, $4, $5)
	`, user.ID, strings.ToLower(user.Email), user.PasswordHash, user.CreatedAt, user.IsAdmin)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return auth.ErrUserAlreadyExists
		}
		return err
	}
	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (auth.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, created_at, is_admin
		FROM users WHERE email = $1
	`, strings.ToLower(email))
	var user auth.User
	var createdAt time.Time
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &createdAt, &user.IsAdmin); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.User{}, auth.ErrNotFound
		}
		return auth.User{}, err
	}
	user.CreatedAt = createdAt.UTC()
	return user, nil
}
