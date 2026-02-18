package repo

import (
	"context"
	"time"

	dom "Worker/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TodoRepo interface {
	Create(ctx context.Context, t dom.Todo) (dom.Todo, error)
	GetByID(ctx context.Context, userID, id int64) (dom.Todo, error)
	List(ctx context.Context, userID int64) ([]dom.Todo, error)
	Update(ctx context.Context, userID, id int64, patch dom.Todo) (dom.Todo, error)
	SoftDelete(ctx context.Context, userID, id int64) error
	MarkDone(ctx context.Context, userID, id int64, done bool) (dom.Todo, error)
	Search(ctx context.Context, userID int64, q string) ([]dom.Todo, error)
	Overdue(ctx context.Context, userID int64) ([]dom.Todo, error)
}

type PGTodoRepo struct {
	db *pgxpool.Pool
}

func NewPGTodoRepo(db *pgxpool.Pool) *PGTodoRepo {
	return &PGTodoRepo{db: db}
}

func (r *PGTodoRepo) Create(ctx context.Context, t dom.Todo) (dom.Todo, error) {
	query := `
		INSERT INTO todos (user_id, title, description, due_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, title, description, is_done, due_at, created_at, updated_at, deleted_at`
	var out dom.Todo
	err := r.db.QueryRow(ctx, query, t.UserID, t.Title, t.Description, t.DueAt).Scan(
		&out.ID, &out.UserID, &out.Title, &out.Description, &out.IsDone, &out.DueAt,
		&out.CreatedAt, &out.UpdatedAt, &out.DeletedAt,
	)
	return out, err
}

func (r *PGTodoRepo) GetByID(ctx context.Context, userID, id int64) (dom.Todo, error) {
	query := `
		SELECT id, user_id, title, description, is_done, due_at, created_at, updated_at, deleted_at
		FROM todos WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`
	var t dom.Todo
	err := r.db.QueryRow(ctx, query, id, userID).Scan(
		&t.ID, &t.UserID, &t.Title, &t.Description, &t.IsDone, &t.DueAt,
		&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
	)
	return t, err
}

func (r *PGTodoRepo) List(ctx context.Context, userID int64) ([]dom.Todo, error) {
	query := `
		SELECT id, user_id, title, description, is_done, due_at, created_at, updated_at, deleted_at
		FROM todos WHERE user_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []dom.Todo
	for rows.Next() {
		var t dom.Todo
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.IsDone, &t.DueAt,
			&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (r *PGTodoRepo) Update(ctx context.Context, userID, id int64, patch dom.Todo) (dom.Todo, error) {
	query := `
		UPDATE todos SET title = $3, description = $4, due_at = $5, is_done = $6, updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		RETURNING id, user_id, title, description, is_done, due_at, created_at, updated_at, deleted_at`
	var t dom.Todo
	err := r.db.QueryRow(ctx, query, id, userID, patch.Title, patch.Description, patch.DueAt, patch.IsDone).Scan(
		&t.ID, &t.UserID, &t.Title, &t.Description, &t.IsDone, &t.DueAt,
		&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
	)
	return t, err
}

func (r *PGTodoRepo) SoftDelete(ctx context.Context, userID, id int64) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(ctx, `UPDATE todos SET deleted_at = $3, updated_at = $3 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, id, userID, now)
	return err
}

func (r *PGTodoRepo) MarkDone(ctx context.Context, userID, id int64, done bool) (dom.Todo, error) {
	query := `
		UPDATE todos SET is_done = $3, updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		RETURNING id, user_id, title, description, is_done, due_at, created_at, updated_at, deleted_at`
	var t dom.Todo
	err := r.db.QueryRow(ctx, query, id, userID, done).Scan(
		&t.ID, &t.UserID, &t.Title, &t.Description, &t.IsDone, &t.DueAt,
		&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
	)
	return t, err
}

func (r *PGTodoRepo) Search(ctx context.Context, userID int64, q string) ([]dom.Todo, error) {
	pattern := "%" + q + "%"
	query := `
		SELECT id, user_id, title, description, is_done, due_at, created_at, updated_at, deleted_at
		FROM todos WHERE user_id = $1 AND deleted_at IS NULL AND (title ILIKE $2 OR description ILIKE $2)
		ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []dom.Todo
	for rows.Next() {
		var t dom.Todo
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.IsDone, &t.DueAt,
			&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (r *PGTodoRepo) Overdue(ctx context.Context, userID int64) ([]dom.Todo, error) {
	query := `
		SELECT id, user_id, title, description, is_done, due_at, created_at, updated_at, deleted_at
		FROM todos WHERE user_id = $1 AND deleted_at IS NULL AND is_done = FALSE AND due_at IS NOT NULL AND due_at < NOW()
		ORDER BY due_at ASC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []dom.Todo
	for rows.Next() {
		var t dom.Todo
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.IsDone, &t.DueAt,
			&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}
