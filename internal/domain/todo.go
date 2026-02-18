package domain

import "time"

// Domain entity: бизнес-объект (истина).
// Не зависит от Gin, Postgres, Redis.
type Todo struct {
	ID          int64
	Title       string
	Description string
	IsDone      bool
	DueAt       *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
