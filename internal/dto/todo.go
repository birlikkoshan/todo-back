package dto

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// DueAt parses due_at from JSON as either date-only ("2006-01-02") or RFC3339.
// Date-only is stored as start of that day in UTC.
type DueAt struct{ t *time.Time }

func (d *DueAt) UnmarshalJSON(data []byte) error {
	var raw *string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw == nil || strings.TrimSpace(*raw) == "" {
		d.t = nil
		return nil
	}
	s := strings.TrimSpace(*raw)
	layouts := []string{
		"2006-01-02",        // date only
		time.RFC3339,        // 2006-01-02T15:04:05Z07:00
		time.RFC3339Nano,    // with nanoseconds
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, s)
		if err == nil {
			// If it was date-only (no time component), use start of day UTC
			if layout == "2006-01-02" {
				parsed = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
			}
			d.t = &parsed
			return nil
		}
	}
	return fmt.Errorf("due_at: use date (YYYY-MM-DD) or RFC3339 datetime")
}

// Ptr returns *time.Time for use in service/domain.
func (d DueAt) Ptr() *time.Time { return d.t }

type CreateTodoRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=120"`
	Description string `json:"description" binding:"max=1000"`
	DueAt       DueAt  `json:"due_at"` // optional: "2026-02-19" or RFC3339
}

type UpdateTodoRequest struct {
	Title       *string `json:"title" binding:"omitempty,min=1,max=120"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
	DueAt       *DueAt  `json:"due_at"` // nil = не менять, значение = поставить
}

type TodoResponse struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	IsDone      bool       `json:"is_done"`
	DueAt       *time.Time `json:"due_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type ListTodosResponse struct {
	Items []TodoResponse `json:"items"`
}
