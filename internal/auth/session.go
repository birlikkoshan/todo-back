package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	sessionKeyPrefix = "session:"
	sessionTTL       = 24 * time.Hour
)

// Store manages sessions in Redis.
type Store struct {
	rdb *redis.Client
	ttl time.Duration
}

// NewStore returns a new session store.
func NewStore(rdb *redis.Client, ttl time.Duration) *Store {
	if ttl <= 0 {
		ttl = sessionTTL
	}
	return &Store{rdb: rdb, ttl: ttl}
}

// Create stores a new session and returns its ID.
func (s *Store) Create(ctx context.Context) (string, error) {
	id, err := newSessionID()
	if err != nil {
		return "", err
	}
	key := sessionKeyPrefix + id
	if err := s.rdb.Set(ctx, key, "1", s.ttl).Err(); err != nil {
		return "", err
	}
	return id, nil
}

// Delete removes a session by ID.
func (s *Store) Delete(ctx context.Context, id string) error {
	return s.rdb.Del(ctx, sessionKeyPrefix+id).Err()
}

// Exists returns true if the session exists.
func (s *Store) Exists(ctx context.Context, id string) (bool, error) {
	n, err := s.rdb.Exists(ctx, sessionKeyPrefix+id).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func newSessionID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("rand: %w", err)
	}
	return hex.EncodeToString(b), nil
}
