package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	sessionKeyPrefix = "session:"
	sessionTTL       = 24 * time.Hour
)

// Store manages sessions in Redis. Value is user_id (int64 as string).
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

// Create stores a new session for the given user and returns session ID.
func (s *Store) Create(ctx context.Context, userID int64) (string, error) {
	id, err := newSessionID()
	if err != nil {
		return "", err
	}
	key := sessionKeyPrefix + id
	val := strconv.FormatInt(userID, 10)
	if err := s.rdb.Set(ctx, key, val, s.ttl).Err(); err != nil {
		return "", err
	}
	return id, nil
}

// GetUserID returns user ID for the session, or 0 and false if not found/invalid.
func (s *Store) GetUserID(ctx context.Context, sessionID string) (int64, bool) {
	if sessionID == "" {
		return 0, false
	}
	val, err := s.rdb.Get(ctx, sessionKeyPrefix+sessionID).Result()
	if err == redis.Nil || err != nil {
		return 0, false
	}
	userID, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, false
	}
	return userID, true
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
