package cache

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	dom "Worker/internal/domain"

	"github.com/redis/go-redis/v9"
)

const (
	keyListPrefix    = "todo:list:"
	keyOverduePrefix = "todo:overdue:"
	keySearchPrefix  = "todo:search:"
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

// GetList returns cached list for user or nil if miss.
func (c *TodoCache) GetList(ctx context.Context, userID int64) ([]dom.Todo, error) {
	key := keyListPrefix + userKey(userID)
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

// SetList stores the list in cache for user.
func (c *TodoCache) SetList(ctx context.Context, userID int64, list []dom.Todo) error {
	b, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, keyListPrefix+userKey(userID), b, c.ttl).Err()
}

// GetSearch returns cached search result for user and query q, or nil if miss.
func (c *TodoCache) GetSearch(ctx context.Context, userID int64, q string) ([]dom.Todo, error) {
	key := keySearchPrefix + userKey(userID) + ":" + normalizeQuery(q)
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

// SetSearch stores the search result in cache for user.
func (c *TodoCache) SetSearch(ctx context.Context, userID int64, q string, list []dom.Todo) error {
	b, err := json.Marshal(list)
	if err != nil {
		return err
	}
	key := keySearchPrefix + userKey(userID) + ":" + normalizeQuery(q)
	return c.rdb.Set(ctx, key, b, c.ttl).Err()
}

// GetOverdue returns cached overdue list for user or nil if miss.
func (c *TodoCache) GetOverdue(ctx context.Context, userID int64) ([]dom.Todo, error) {
	key := keyOverduePrefix + userKey(userID)
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

// SetOverdue stores the overdue list in cache for user.
func (c *TodoCache) SetOverdue(ctx context.Context, userID int64, list []dom.Todo) error {
	b, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, keyOverduePrefix+userKey(userID), b, c.ttl).Err()
}

// InvalidateAll removes list, overdue, and search keys for the user (cache invalidation on write).
func (c *TodoCache) InvalidateAll(ctx context.Context, userID int64) error {
	uk := userKey(userID)
	if err := c.rdb.Del(ctx, keyListPrefix+uk, keyOverduePrefix+uk).Err(); err != nil {
		return err
	}
	iter := c.rdb.Scan(ctx, 0, keySearchPrefix+uk+"*", 100).Iterator()
	for iter.Next(ctx) {
		if err := c.rdb.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

func userKey(userID int64) string {
	return strconv.FormatInt(userID, 10)
}

func normalizeQuery(q string) string {
	return strings.TrimSpace(strings.ToLower(q))
}
