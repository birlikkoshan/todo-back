package app

import (
	"context"
	"fmt"
	"time"

	"Worker/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
)

type App struct {
	cfg    config.Config
	db     *pgxpool.Pool
	redis  *redis.Client
	router *gin.Engine
}

func New(cfg config.Config) (*App, error) {
	a := &App{cfg: cfg}

	// 1) Connect Postgres
	db, err := newPostgres(cfg.PG.DSN)
	if err != nil {
		return nil, err
	}
	a.db = db

	// 2) Connect Redis
	rdb, err := newRedis(cfg.Redis)
	if err != nil {
		db.Close()
		return nil, err
	}
	a.redis = rdb

	// 3) Run migrations (goose)
	//    Важно: migrations path должен существовать внутри контейнера.
	//    Мы копировали ./migrations в Dockerfile → значит ok.
	if err := runMigrations(cfg.PG.DSN, "./migrations"); err != nil {
		a.redis.Close()
		a.db.Close()
		return nil, err
	}

	a.router = newRouter(cfg, a.db, a.redis)
	return a, nil
}

func (a *App) Router() *gin.Engine {
	return a.router
}

// Close закрывает ресурсы приложения.
// Контекст здесь на будущее (если добавишь закрытие с таймаутом, drain очередей и т.д.)
func (a *App) Close(ctx context.Context) error {
	_ = ctx // сейчас не нужен, но оставляем для "взрослого" API

	// Закрывать лучше в обратном порядке создания
	if a.redis != nil {
		_ = a.redis.Close()
	}
	if a.db != nil {
		a.db.Close()
	}
	return nil
}

// -------------------- helpers --------------------

func newPostgres(dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("pg parse config: %w", err)
	}

	// Настройки пула — минимум для адекватной работы
	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("pg connect: %w", err)
	}

	// Проверка соединения (fail fast)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pg ping: %w", err)
	}

	return pool, nil
}

func newRedis(cfg config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return rdb, nil
}

func runMigrations(dsn string, migrationsDir string) error {
	// goose требует database/sql, поэтому открываем sql.DB через pgx stdlib
	// Это нормально: миграции — отдельная прослойка, не runtime-пул.

	db, err := goose.OpenDBWithDriver("pgx", dsn)
	if err != nil {
		return fmt.Errorf("goose open db: %w", err)
	}
	defer db.Close()

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}

func newRouter(cfg config.Config, db *pgxpool.Pool, rdb *redis.Client) *gin.Engine {
	r := gin.Default()
	Setup(r, cfg, db, rdb)
	return r
}
