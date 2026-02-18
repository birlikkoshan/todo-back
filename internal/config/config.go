package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// durationSeconds parses env as time.Duration: "10s", "5m" or bare number = seconds (e.g. "10" -> 10s).
type durationSeconds time.Duration

func (d *durationSeconds) UnmarshalEnvironment(data string) error {
	v, err := parseDuration(data)
	if err != nil {
		return err
	}
	*d = durationSeconds(v)
	return nil
}

func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	// Strip optional surrounding quotes: "10s" or '10s'
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		s = s[1 : len(s)-1]
	}

	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	// Bare number first (e.g. Railway HTTP_READ_TIMEOUT=10) — so "10s" never goes to ParseInt
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Duration(n) * time.Second, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("duration must be like 10s, 5m or a number of seconds: %w", err)
	}
	return d, nil
}

func (d durationSeconds) Duration() time.Duration { return time.Duration(d) }

type Config struct {
	App   AppConfig
	HTTP  HTTPConfig
	PG    PGConfig
	Redis RedisConfig
}

type AppConfig struct {
	Env     string `env:"APP_ENV" env-default:"dev"`
	Version string `env:"VERSION" env-default:"dev"`
}

type HTTPConfig struct {
	Port string `env:"HTTP_PORT" env-default:"8080"`

	// Эти поля пригодятся позже, если захочешь перекинуть таймауты в main через cfg
	// Значение: "10s", "5m" или число секунд без суффикса (например 10).
	ReadTimeout  durationSeconds `env:"HTTP_READ_TIMEOUT" env-default:"10s"`
	WriteTimeout durationSeconds `env:"HTTP_WRITE_TIMEOUT" env-default:"10s"`
	IdleTimeout  durationSeconds `env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
}

type PGConfig struct {
	DSN string `env:"PG_DSN" env-required:"true"`
}

type RedisConfig struct {
	Addr     string `env:"REDIS_ADDR" env-required:"true"`
	Password string `env:"REDIS_PASSWORD" env-default:""`
	DB       int    `env:"REDIS_DB" env-default:"0"`

	// TTL для кеша (на будущее). Значение: "60s", "5m" или число секунд.
	DefaultTTL durationSeconds `env:"REDIS_DEFAULT_TTL" env-default:"60s"`
}

func Load() (Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return Config{}, fmt.Errorf("read env: %w", err)
	}
	return cfg, nil
}
