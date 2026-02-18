package service

import (
	"context"
	"errors"
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

func (s *TodoService) Create(ctx context.Context, title, desc string, dueAt *time.Time) (dom.Todo, error) {
	title = strings.TrimSpace(title)
	desc = strings.TrimSpace(desc)

	if dueAt != nil && dueAt.Before(time.Now().UTC()) {
		return dom.Todo{}, ErrInvalidDueDate
	}

	t, err := s.repo.Create(ctx, dom.Todo{
		Title:       title,
		Description: desc,
		DueAt:       dueAt,
	})
	if err != nil {
		return dom.Todo{}, err
	}
	s.invalidateCache(ctx)
	return t, nil
}

func (s *TodoService) List(ctx context.Context) ([]dom.Todo, error) {
	if s.cache != nil {
		v, err, _ := s.sf.Do("list", func() (interface{}, error) {
			if list, err := s.cache.GetList(ctx); err == nil && list != nil {
				return list, nil
			}
			list, err := s.repo.List(ctx)
			if err != nil {
				return nil, err
			}
			_ = s.cache.SetList(ctx, list)
			return list, nil
		})
		if err != nil {
			return nil, err
		}
		return v.([]dom.Todo), nil
	}
	return s.repo.List(ctx)
}

func (s *TodoService) GetByID(ctx context.Context, id int64) (dom.Todo, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dom.Todo{}, ErrNotFound
		}
		return dom.Todo{}, err
	}
	return t, nil
}

func (s *TodoService) Update(ctx context.Context, id int64, title *string, desc *string, dueAt *time.Time) (dom.Todo, error) {
	existing, err := s.repo.GetByID(ctx, id)
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
	t, err := s.repo.Update(ctx, id, patch)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dom.Todo{}, ErrNotFound
		}
		return dom.Todo{}, err
	}
	s.invalidateCache(ctx)
	return t, nil
}

func (s *TodoService) Complete(ctx context.Context, id int64) (dom.Todo, error) {
	t, err := s.repo.MarkDone(ctx, id, true)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dom.Todo{}, ErrNotFound
		}
		return dom.Todo{}, err
	}
	s.invalidateCache(ctx)
	return t, nil
}

func (s *TodoService) Delete(ctx context.Context, id int64) error {
	err := s.repo.SoftDelete(ctx, id)
	if err != nil {
		return err
	}
	s.invalidateCache(ctx)
	return nil
}

func (s *TodoService) Search(ctx context.Context, q string) ([]dom.Todo, error) {
	q = strings.TrimSpace(q)
	if s.cache != nil {
		key := "search:" + strings.ToLower(q)
		v, err, _ := s.sf.Do(key, func() (interface{}, error) {
			if list, err := s.cache.GetSearch(ctx, q); err == nil && list != nil {
				return list, nil
			}
			list, err := s.repo.Search(ctx, q)
			if err != nil {
				return nil, err
			}
			_ = s.cache.SetSearch(ctx, q, list)
			return list, nil
		})
		if err != nil {
			return nil, err
		}
		return v.([]dom.Todo), nil
	}
	return s.repo.Search(ctx, q)
}

func (s *TodoService) Overdue(ctx context.Context) ([]dom.Todo, error) {
	if s.cache != nil {
		v, err, _ := s.sf.Do("overdue", func() (interface{}, error) {
			if list, err := s.cache.GetOverdue(ctx); err == nil && list != nil {
				return list, nil
			}
			list, err := s.repo.Overdue(ctx)
			if err != nil {
				return nil, err
			}
			_ = s.cache.SetOverdue(ctx, list)
			return list, nil
		})
		if err != nil {
			return nil, err
		}
		return v.([]dom.Todo), nil
	}
	return s.repo.Overdue(ctx)
}

func (s *TodoService) invalidateCache(ctx context.Context) {
	if s.cache != nil {
		_ = s.cache.InvalidateAll(ctx)
	}
}
