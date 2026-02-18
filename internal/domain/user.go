package domain

import "time"

// User is the domain entity for a user account.
type User struct {
	ID           int64
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}
