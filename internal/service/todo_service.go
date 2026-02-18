package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	dom "Worker/internal/domain"
	"Worker/internal/repo"

	"github.com/jackc/pgx/v5"
	"golang.org/x/sync/singleflight"

	"Worker/internal/cache"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrInvalidDueDate = errors.New("due_at is in the past")
)

type TodoService struct {
	repo  repo.TodoRepo
	cache *cache.TodoCache
	sf    singleflight.Group
}

// NewTodoService creates a TodoService. If c is nil, caching is disabled.
func NewTodoService(r repo.TodoRepo, c *cache.TodoCache) *TodoService {
	return &TodoService{repo: r, cache: c}
}

func (s *TodoService) Create(ctx context.Context, userID int64, title, desc string, dueAt *time.Time) (dom.Todo, error) {
	title = strings.TrimSpace(title)
	desc = strings.TrimSpace(desc)

	if dueAt != nil && dueAt.Before(time.Now().UTC()) {
		return dom.Todo{}, ErrInvalidDueDate
	}

	t, err := s.repo.Create(ctx, dom.Todo{
		UserID:      userID,
		Title:       title,
		Description: desc,
		DueAt:       dueAt,
	})
	if err != nil {
		return dom.Todo{}, err
	}
	s.invalidateCache(ctx, userID)
	return t, nil
}

func (s *TodoService) List(ctx context.Context, userID int64) ([]dom.Todo, error) {
	if s.cache != nil {
		key := "list:" + strconv.FormatInt(userID, 10)
		v, err, _ := s.sf.Do(key, func() (interface{}, error) {
			if list, err := s.cache.GetList(ctx, userID); err == nil && list != nil {
				return list, nil
			}
			list, err := s.repo.List(ctx, userID)
			if err != nil {
				return nil, err
			}
			_ = s.cache.SetList(ctx, userID, list)
			return list, nil
		})
		if err != nil {
			return nil, err
		}
		return v.([]dom.Todo), nil
	}
	return s.repo.List(ctx, userID)
}

func (s *TodoService) GetByID(ctx context.Context, userID, id int64) (dom.Todo, error) {
	t, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dom.Todo{}, ErrNotFound
		}
		return dom.Todo{}, err
	}
	return t, nil
}

func (s *TodoService) Update(ctx context.Context, userID, id int64, title *string, desc *string, dueAt *time.Time, isDone *bool) (dom.Todo, error) {
	existing, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dom.Todo{}, ErrNotFound
		}
		return dom.Todo{}, err
	}
	patch := existing
	if title != nil {
		patch.Title = strings.TrimSpace(*title)
	}
	if desc != nil {
		patch.Description = strings.TrimSpace(*desc)
	}
	if dueAt != nil {
		if dueAt.Before(time.Now().UTC()) {
			return dom.Todo{}, ErrInvalidDueDate
		}
		patch.DueAt = dueAt
	}
	if isDone != nil {
		patch.IsDone = *isDone
	}
	t, err := s.repo.Update(ctx, userID, id, patch)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dom.Todo{}, ErrNotFound
		}
		return dom.Todo{}, err
	}
	s.invalidateCache(ctx, userID)
	return t, nil
}

func (s *TodoService) Complete(ctx context.Context, userID, id int64) (dom.Todo, error) {
	t, err := s.repo.MarkDone(ctx, userID, id, true)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dom.Todo{}, ErrNotFound
		}
		return dom.Todo{}, err
	}
	s.invalidateCache(ctx, userID)
	return t, nil
}

func (s *TodoService) Delete(ctx context.Context, userID, id int64) error {
	err := s.repo.SoftDelete(ctx, userID, id)
	if err != nil {
		return err
	}
	s.invalidateCache(ctx, userID)
	return nil
}

func (s *TodoService) Search(ctx context.Context, userID int64, q string) ([]dom.Todo, error) {
	q = strings.TrimSpace(q)
	if s.cache != nil {
		key := "search:" + strconv.FormatInt(userID, 10) + ":" + strings.ToLower(q)
		v, err, _ := s.sf.Do(key, func() (interface{}, error) {
			if list, err := s.cache.GetSearch(ctx, userID, q); err == nil && list != nil {
				return list, nil
			}
			list, err := s.repo.Search(ctx, userID, q)
			if err != nil {
				return nil, err
			}
			_ = s.cache.SetSearch(ctx, userID, q, list)
			return list, nil
		})
		if err != nil {
			return nil, err
		}
		return v.([]dom.Todo), nil
	}
	return s.repo.Search(ctx, userID, q)
}

func (s *TodoService) Overdue(ctx context.Context, userID int64) ([]dom.Todo, error) {
	if s.cache != nil {
		key := "overdue:" + strconv.FormatInt(userID, 10)
		v, err, _ := s.sf.Do(key, func() (interface{}, error) {
			if list, err := s.cache.GetOverdue(ctx, userID); err == nil && list != nil {
				return list, nil
			}
			list, err := s.repo.Overdue(ctx, userID)
			if err != nil {
				return nil, err
			}
			_ = s.cache.SetOverdue(ctx, userID, list)
			return list, nil
		})
		if err != nil {
			return nil, err
		}
		return v.([]dom.Todo), nil
	}
	return s.repo.Overdue(ctx, userID)
}

func (s *TodoService) invalidateCache(ctx context.Context, userID int64) {
	if s.cache != nil {
		_ = s.cache.InvalidateAll(ctx, userID)
	}
}
