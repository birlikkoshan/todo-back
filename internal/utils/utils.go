package utils

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

// ParseDurationEnv parses an env value as time.Duration:
// - "10s", "5m" etc. (time.ParseDuration)
// - bare number "10" = seconds (10s)
func ParseDurationEnv(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	// Strip optional surrounding quotes: "10s" or '10s'
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		s = s[1 : len(s)-1]
	}

	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	// Bare number first (e.g. HTTP_READ_TIMEOUT=10) — treat as seconds.
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Duration(n) * time.Second, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("duration must be like 10s, 5m or a number of seconds: %w", err)
	}
	return d, nil
}

// ParseRedisURL extracts host:port, password and DB from redis:// or rediss:// URL.
func ParseRedisURL(s string) (addr, password string, db int, err error) {
	u, err := url.Parse(strings.TrimSpace(s))
	if err != nil {
		return "", "", 0, err
	}
	if u.Scheme != "redis" && u.Scheme != "rediss" {
		return "", "", 0, fmt.Errorf("scheme must be redis or rediss, got %q", u.Scheme)
	}
	addr = u.Host
	if addr == "" {
		return "", "", 0, fmt.Errorf("missing host in Redis URL")
	}
	if u.User != nil {
		password, _ = u.User.Password()
	}
	if u.Path != "" && len(u.Path) > 1 {
		db, _ = strconv.Atoi(strings.TrimPrefix(u.Path, "/"))
	}
	return addr, password, db, nil
}

// IsPGUniqueViolation reports whether error is PostgreSQL unique constraint violation (code 23505).
func IsPGUniqueViolation(err error) bool {
	var pge *pgconn.PgError
	if errors.As(err, &pge) {
		return pge.Code == "23505"
	}
	return false
}

