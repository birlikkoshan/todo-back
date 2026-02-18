package repo

import (
	"context"

	dom "Worker/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepo provides user persistence.
type UserRepo interface {
	GetByUsername(ctx context.Context, username string) (dom.User, error)
	Create(ctx context.Context, username, passwordHash string) (dom.User, error)
}

// PGUserRepo implements UserRepo with Postgres.
type PGUserRepo struct {
	db *pgxpool.Pool
}

// NewPGUserRepo returns a new PGUserRepo.
func NewPGUserRepo(db *pgxpool.Pool) *PGUserRepo {
	return &PGUserRepo{db: db}
}

// GetByUsername returns the user by username.
func (r *PGUserRepo) GetByUsername(ctx context.Context, username string) (dom.User, error) {
	var u dom.User
	err := r.db.QueryRow(ctx,
		`SELECT id, username, password_hash, created_at FROM users WHERE username = $1`,
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	return u, err
}

// Create inserts a new user and returns it.
func (r *PGUserRepo) Create(ctx context.Context, username, passwordHash string) (dom.User, error) {
	query := `
		INSERT INTO users (username, password_hash)
		VALUES ($1, $2)
		RETURNING id, username, password_hash, created_at`
	var u dom.User
	err := r.db.QueryRow(ctx, query, username, passwordHash).Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt,
	)
	return u, err
}
