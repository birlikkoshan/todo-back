package config

import (
	"fmt"
	"strings"
	"time"

	"Worker/internal/utils"

	"github.com/ilyakaznacheev/cleanenv"
)

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
	ReadTimeoutRaw  string        `env:"HTTP_READ_TIMEOUT" env-default:"10s"`
	WriteTimeoutRaw string        `env:"HTTP_WRITE_TIMEOUT" env-default:"10s"`
	IdleTimeoutRaw  string        `env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
	ReadTimeout     time.Duration `env:"-"`
	WriteTimeout    time.Duration `env:"-"`
	IdleTimeout     time.Duration `env:"-"`
}

type PGConfig struct {
	DSN string `env:"PG_DSN" env-required:"true"`
}

type RedisConfig struct {
	// Addr is "host:port". Optional if URL is set (e.g. Railway REDIS_URL).
	Addr     string `env:"REDIS_ADDR" env-default:""`
	Password string `env:"REDIS_PASSWORD" env-default:""`
	DB       int    `env:"REDIS_DB" env-default:"0"`
	// URL overrides Addr/Password/DB if set. Example: redis://default:password@host:35459
	URL string `env:"REDIS_URL" env-default:""`

	// TTL для кеша (на будущее). Значение: "60s", "5m" или число секунд.
	DefaultTTLRaw string        `env:"REDIS_DEFAULT_TTL" env-default:"60"`
	DefaultTTL    time.Duration `env:"-"`
}

func Load() (Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return Config{}, fmt.Errorf("read env: %w", err)
	}
	// Parse Redis URL / Addr overrides
	if cfg.Redis.URL != "" {
		addr, password, db, err := utils.ParseRedisURL(cfg.Redis.URL)
		if err != nil {
			return Config{}, fmt.Errorf("REDIS_URL: %w", err)
		}
		cfg.Redis.Addr = addr
		cfg.Redis.Password = password
		cfg.Redis.DB = db
	} else if s := strings.TrimSpace(cfg.Redis.Addr); strings.HasPrefix(s, "redis://") || strings.HasPrefix(s, "rediss://") {
		// Railway and others sometimes set REDIS_ADDR to the full URL
		addr, password, db, err := utils.ParseRedisURL(s)
		if err != nil {
			return Config{}, fmt.Errorf("REDIS_ADDR (URL): %w", err)
		}
		cfg.Redis.Addr = addr
		cfg.Redis.Password = password
		cfg.Redis.DB = db
	}
	if cfg.Redis.Addr == "" {
		return Config{}, fmt.Errorf("REDIS_ADDR or REDIS_URL is required")
	}

	// Parse HTTP durations
	var err error
	if cfg.HTTP.ReadTimeout, err = utils.ParseDurationEnv(cfg.HTTP.ReadTimeoutRaw); err != nil {
		return Config{}, fmt.Errorf("HTTP_READ_TIMEOUT: %w", err)
	}
	if cfg.HTTP.WriteTimeout, err = utils.ParseDurationEnv(cfg.HTTP.WriteTimeoutRaw); err != nil {
		return Config{}, fmt.Errorf("HTTP_WRITE_TIMEOUT: %w", err)
	}
	if cfg.HTTP.IdleTimeout, err = utils.ParseDurationEnv(cfg.HTTP.IdleTimeoutRaw); err != nil {
		return Config{}, fmt.Errorf("HTTP_IDLE_TIMEOUT: %w", err)
	}

	// Parse Redis TTL
	if cfg.Redis.DefaultTTL, err = utils.ParseDurationEnv(cfg.Redis.DefaultTTLRaw); err != nil {
		return Config{}, fmt.Errorf("REDIS_DEFAULT_TTL: %w", err)
	}

	return cfg, nil
}
