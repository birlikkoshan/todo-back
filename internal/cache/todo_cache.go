package cache

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	dom "Worker/internal/domain"

	"github.com/redis/go-redis/v9"
)

const (
	keyList    = "todo:list"
	keyOverdue = "todo:overdue"
	keySearch  = "todo:search:"
)

// TodoCache caches todo list, search, and overdue results in Redis.
type TodoCache struct {
	rdb *redis.Client
	ttl time.Duration
}

// NewTodoCache returns a new TodoCache.
func NewTodoCache(rdb *redis.Client, ttl time.Duration) *TodoCache {
	return &TodoCache{rdb: rdb, ttl: ttl}
}

// GetList returns cached list or nil if miss.
func (c *TodoCache) GetList(ctx context.Context) ([]dom.Todo, error) {
	b, err := c.rdb.Get(ctx, keyList).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var list []dom.Todo
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// SetList stores the list in cache.
func (c *TodoCache) SetList(ctx context.Context, list []dom.Todo) error {
	b, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, keyList, b, c.ttl).Err()
}

// GetSearch returns cached search result for query q, or nil if miss.
func (c *TodoCache) GetSearch(ctx context.Context, q string) ([]dom.Todo, error) {
	key := keySearch + normalizeQuery(q)
	b, err := c.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var list []dom.Todo
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// SetSearch stores the search result in cache.
func (c *TodoCache) SetSearch(ctx context.Context, q string, list []dom.Todo) error {
	b, err := json.Marshal(list)
	if err != nil {
		return err
	}
	key := keySearch + normalizeQuery(q)
	return c.rdb.Set(ctx, key, b, c.ttl).Err()
}

// GetOverdue returns cached overdue list or nil if miss.
func (c *TodoCache) GetOverdue(ctx context.Context) ([]dom.Todo, error) {
	b, err := c.rdb.Get(ctx, keyOverdue).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var list []dom.Todo
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// SetOverdue stores the overdue list in cache.
func (c *TodoCache) SetOverdue(ctx context.Context, list []dom.Todo) error {
	b, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, keyOverdue, b, c.ttl).Err()
}

// InvalidateAll removes list, overdue, and all search keys (cache invalidation on write).
func (c *TodoCache) InvalidateAll(ctx context.Context) error {
	if err := c.rdb.Del(ctx, keyList, keyOverdue).Err(); err != nil {
		return err
	}
	iter := c.rdb.Scan(ctx, 0, keySearch+"*", 100).Iterator()
	for iter.Next(ctx) {
		if err := c.rdb.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

func normalizeQuery(q string) string {
	return strings.TrimSpace(strings.ToLower(q))
}
