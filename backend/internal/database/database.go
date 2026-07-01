package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"tasks-app/internal/models"
)

// ErrNotFound возвращается, когда задача не найдена.
var ErrNotFound = errors.New("task not found")

// DB оборачивает пул соединений с PostgreSQL.
type DB struct {
	pool *pgxpool.Pool
}

// New создаёт пул соединений и проверяет доступность БД.
func New(ctx context.Context, dsn string) (*DB, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("создание пула: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping БД: %w", err)
	}

	return &DB{pool: pool}, nil
}

// Close закрывает пул соединений.
func (db *DB) Close() {
	db.pool.Close()
}

// ListTasks возвращает все задачи, отсортированные по дате создания (новые сверху).
func (db *DB) ListTasks(ctx context.Context) ([]models.Task, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, title, description, done, created_at, updated_at
		FROM tasks
		ORDER BY created_at DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]models.Task, 0)
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Done, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// GetTask возвращает одну задачу по id.
func (db *DB) GetTask(ctx context.Context, id int64) (models.Task, error) {
	var t models.Task
	err := db.pool.QueryRow(ctx, `
		SELECT id, title, description, done, created_at, updated_at
		FROM tasks WHERE id = $1
	`, id).Scan(&t.ID, &t.Title, &t.Description, &t.Done, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return t, ErrNotFound
	}
	return t, err
}

// CreateTask создаёт новую задачу и возвращает её.
func (db *DB) CreateTask(ctx context.Context, in models.TaskInput) (models.Task, error) {
	var t models.Task
	err := db.pool.QueryRow(ctx, `
		INSERT INTO tasks (title, description, done)
		VALUES ($1, $2, $3)
		RETURNING id, title, description, done, created_at, updated_at
	`, in.Title, in.Description, in.Done).
		Scan(&t.ID, &t.Title, &t.Description, &t.Done, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

// UpdateTask обновляет задачу по id.
func (db *DB) UpdateTask(ctx context.Context, id int64, in models.TaskInput) (models.Task, error) {
	var t models.Task
	err := db.pool.QueryRow(ctx, `
		UPDATE tasks
		SET title = $1, description = $2, done = $3, updated_at = now()
		WHERE id = $4
		RETURNING id, title, description, done, created_at, updated_at
	`, in.Title, in.Description, in.Done, id).
		Scan(&t.ID, &t.Title, &t.Description, &t.Done, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return t, ErrNotFound
	}
	return t, err
}

// DeleteTask удаляет задачу по id.
func (db *DB) DeleteTask(ctx context.Context, id int64) error {
	tag, err := db.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
