package config

import (
	"fmt"
	"time"

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
	ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT" env-default:"10s"`
	WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"10s"`
	IdleTimeout  time.Duration `env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
}

type PGConfig struct {
	DSN string `env:"PG_DSN" env-required:"true"`
}

type RedisConfig struct {
	Addr     string `env:"REDIS_ADDR" env-required:"true"`
	Password string `env:"REDIS_PASSWORD" env-default:""`
	DB       int    `env:"REDIS_DB" env-default:"0"`

	// TTL для кеша (на будущее)
	DefaultTTL time.Duration `env:"REDIS_DEFAULT_TTL" env-default:"60s"`
}

func Load() (Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return Config{}, fmt.Errorf("read env: %w", err)
	}
	return cfg, nil
}
