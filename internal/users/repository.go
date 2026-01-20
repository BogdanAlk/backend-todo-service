package users

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEmailAlreadyExists = errors.New("email already exists")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateUser(ctx context.Context, email, passwordHash string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users(email, password_hash) VALUES($1, $2) RETURNING id`,
		email, passwordHash,
	).Scan(&id)

	if err == nil {
		return id, nil
	}

	// Unique violation (email)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return 0, ErrEmailAlreadyExists
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

	return 0, err
}
