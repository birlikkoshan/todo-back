package repo

import (
	"context"
	"time"

	dom "Worker/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TodoRepo interface {
	Create(ctx context.Context, t dom.Todo) (dom.Todo, error)
	GetByID(ctx context.Context, id int64) (dom.Todo, error)
	List(ctx context.Context) ([]dom.Todo, error)
	Update(ctx context.Context, id int64, patch dom.Todo) (dom.Todo, error)
	SoftDelete(ctx context.Context, id int64) error
	MarkDone(ctx context.Context, id int64, done bool) (dom.Todo, error)
	Search(ctx context.Context, q string) ([]dom.Todo, error)
	Overdue(ctx context.Context) ([]dom.Todo, error)
}

type PGTodoRepo struct {
	db *pgxpool.Pool
}

func NewPGTodoRepo(db *pgxpool.Pool) *PGTodoRepo {
	return &PGTodoRepo{db: db}
}

func (r *PGTodoRepo) Create(ctx context.Context, t dom.Todo) (dom.Todo, error) {
	query := `
		INSERT INTO todos (title, description, due_at)
		VALUES ($1, $2, $3)
		RETURNING id, title, description, is_done, due_at, created_at, updated_at, deleted_at`
	var out dom.Todo
	err := r.db.QueryRow(ctx, query, t.Title, t.Description, t.DueAt).Scan(
		&out.ID, &out.Title, &out.Description, &out.IsDone, &out.DueAt,
		&out.CreatedAt, &out.UpdatedAt, &out.DeletedAt,
	)
	return out, err
}

func (r *PGTodoRepo) GetByID(ctx context.Context, id int64) (dom.Todo, error) {
	query := `
		SELECT id, title, description, is_done, due_at, created_at, updated_at, deleted_at
		FROM todos WHERE id = $1 AND deleted_at IS NULL`
	var t dom.Todo
	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.Title, &t.Description, &t.IsDone, &t.DueAt,
		&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
	)
	return t, err
}

func (r *PGTodoRepo) List(ctx context.Context) ([]dom.Todo, error) {
	query := `
		SELECT id, title, description, is_done, due_at, created_at, updated_at, deleted_at
		FROM todos WHERE deleted_at IS NULL ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []dom.Todo
	for rows.Next() {
		var t dom.Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.IsDone, &t.DueAt,
			&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (r *PGTodoRepo) Update(ctx context.Context, id int64, patch dom.Todo) (dom.Todo, error) {
	query := `
		UPDATE todos SET title = $2, description = $3, due_at = $4, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, title, description, is_done, due_at, created_at, updated_at, deleted_at`
	var t dom.Todo
	err := r.db.QueryRow(ctx, query, id, patch.Title, patch.Description, patch.DueAt).Scan(
		&t.ID, &t.Title, &t.Description, &t.IsDone, &t.DueAt,
		&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
	)
	return t, err
}

func (r *PGTodoRepo) SoftDelete(ctx context.Context, id int64) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(ctx, `UPDATE todos SET deleted_at = $2, updated_at = $2 WHERE id = $1 AND deleted_at IS NULL`, id, now)
	return err
}

func (r *PGTodoRepo) MarkDone(ctx context.Context, id int64, done bool) (dom.Todo, error) {
	query := `
		UPDATE todos SET is_done = $2, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, title, description, is_done, due_at, created_at, updated_at, deleted_at`
	var t dom.Todo
	err := r.db.QueryRow(ctx, query, id, done).Scan(
		&t.ID, &t.Title, &t.Description, &t.IsDone, &t.DueAt,
		&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
	)
	return t, err
}

func (r *PGTodoRepo) Search(ctx context.Context, q string) ([]dom.Todo, error) {
	pattern := "%" + q + "%"
	query := `
		SELECT id, title, description, is_done, due_at, created_at, updated_at, deleted_at
		FROM todos WHERE deleted_at IS NULL AND (title ILIKE $1 OR description ILIKE $1)
		ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []dom.Todo
	for rows.Next() {
		var t dom.Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.IsDone, &t.DueAt,
			&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (r *PGTodoRepo) Overdue(ctx context.Context) ([]dom.Todo, error) {
	query := `
		SELECT id, title, description, is_done, due_at, created_at, updated_at, deleted_at
		FROM todos WHERE deleted_at IS NULL AND is_done = FALSE AND due_at IS NOT NULL AND due_at < NOW()
		ORDER BY due_at ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []dom.Todo
	for rows.Next() {
		var t dom.Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.IsDone, &t.DueAt,
			&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}
