package tasks

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, userID int64, title, description string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx,
		`INSERT INTO tasks(user_id, title, description) VALUES($1,$2,$3) RETURNING id`,
		userID, title, description,
	).Scan(&id)
	return id, err
}

func (r *Repository) List(ctx context.Context, userID int64, limit, offset int) ([]Task, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, title, description, status, created_at, updated_at
		 FROM tasks
		 WHERE user_id = $1
		 ORDER BY id DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *Repository) Get(ctx context.Context, userID, taskID int64) (Task, error) {
	var t Task
	err := r.pool.QueryRow(ctx,
		`SELECT id, title, description, status, created_at, updated_at
		 FROM tasks
		 WHERE user_id=$1 AND id=$2`,
		userID, taskID,
	).Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	return t, err
}

func (r *Repository) Update(ctx context.Context, userID, taskID int64, title, description, status string) error {
	ct, err := r.pool.Exec(ctx,
		`UPDATE tasks
		 SET title=$1, description=$2, status=$3, updated_at=now()
		 WHERE user_id=$4 AND id=$5`,
		title, description, status, userID, taskID,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, userID, taskID int64) error {
	ct, err := r.pool.Exec(ctx,
		`DELETE FROM tasks WHERE user_id=$1 AND id=$2`,
		userID, taskID,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
