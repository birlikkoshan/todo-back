# Worker — Todo API

REST API для задач (todos) с аутентификацией, поиском и кешированием. Стек: Go, Gin, PostgreSQL (pgx), Redis, миграции через Goose.

---

## Стек и назначение

| Компонент | Назначение |
|-----------|------------|
| **Go 1.26** | Язык и рантайм |
| **Gin** | HTTP-фреймворк: роутинг, middleware, binding, recovery |
| **pgx/v5** | Драйвер и пул соединений к PostgreSQL |
| **Redis (go-redis)** | Сессии аутентификации + кеш для todos |
| **Goose** | SQL-миграции при старте приложения |
| **cleanenv** | Загрузка конфигурации из переменных окружения |
| **validator (go-playground)** | Валидация DTO (теги `validate`) |
| **Swaggo (swag + gin-swagger)** | Генерация OpenAPI и UI Swagger |
| **golang.org/x/sync** | Синхронизация (errgroup и т.п. при необходимости) |

---

## Зависимости (go.mod)

### Прямые (direct)

| Пакет | Для чего |
|-------|----------|
| `github.com/gin-gonic/gin` | Роутер, контекст, binding, middleware |
| `github.com/ilyakaznacheev/cleanenv` | Чтение конфига из `env` (и при желании из `.env`) |
| `github.com/jackc/pgx/v5` | Подключение к PostgreSQL, пул, запросы |
| `github.com/pressly/goose/v3` | Запуск миграций из `./migrations` при старте |
| `github.com/redis/go-redis/v9` | Клиент Redis: сессии, кеш todos |
| `github.com/swaggo/swag` | Генерация Swagger-документа из комментариев |
| `golang.org/x/sync` | Утилиты синхронизации |

### Косвенные (часть, что реально используется)

| Пакет | Для чего |
|-------|----------|
| `github.com/go-playground/validator/v10` | Валидация структур (используется через Gin binding) |
| `github.com/swaggo/files` | Раздача статики для Swagger UI |
| `github.com/swaggo/gin-swagger` | Middleware для отдачи Swagger UI в Gin |
| `github.com/jackc/pgx/v5/stdlib` | Режим `database/sql` для Goose (миграции) |
| `github.com/joho/godotenv` | Опционально: загрузка `.env` (через cleanenv) |
| `github.com/BurntSushi/toml` | Парсинг TOML, если конфиг будет из файла |

UUID в коде может быть через стандартную библиотеку или отдельный пакет — в `go.mod` явно не указан как прямая зависимость.

---

## API эндпоинты

Базовый путь API: **`/api/v1`**.

### Публичные (без авторизации)

| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/health` | Проверка живости: `{"ok": true, "env": "..."}` |
| `GET` | `/version` | Версия приложения из конфига |
| `GET` | `/swagger-doc.json` | OpenAPI JSON для Swagger |
| `GET` | `/swagger/*any` | Swagger UI (интерфейс документации) |

### Auth (`/api/v1`)

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/v1/auth/register` | Регистрация пользователя |
| `POST` | `/api/v1/auth/login` | Вход (создание сессии в Redis) |
| `POST` | `/api/v1/auth/logout` | Выход (удаление сессии) |

### Todos (`/api/v1`) — требуют сессию

Все запросы к todos проходят через middleware `RequireSession`. Каждый пользователь видит и изменяет только свои задачи (фильтрация по `user_id` в БД и кеше).

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/v1/todos` | Создать задачу |
| `GET` | `/api/v1/todos` | Список задач |
| `GET` | `/api/v1/todos/search` | Поиск задач |
| `GET` | `/api/v1/todos/overdue` | Просроченные задачи |
| `GET` | `/api/v1/todos/:id` | Одна задача по ID |
| `PATCH` | `/api/v1/todos/:id` | Обновить задачу |
| `DELETE` | `/api/v1/todos/:id` | Удалить задачу |
| `POST` | `/api/v1/todos/:id/complete` | Отметить выполненной |

---

## Конфигурация (переменные окружения)

| Переменная | Обязательность | По умолчанию | Описание |
|------------|----------------|--------------|----------|
| `APP_ENV` | нет | `dev` | Окружение (dev/prod) |
| `VERSION` | нет | `dev` | Версия (для `/version`) |
| `HTTP_PORT` | нет | `8080` | Порт HTTP-сервера |
| `HTTP_READ_TIMEOUT` | нет | `10s` | Таймаут чтения |
| `HTTP_WRITE_TIMEOUT` | нет | `10s` | Таймаут записи |
| `HTTP_IDLE_TIMEOUT` | нет | `60s` | Idle таймаут |
| `PG_DSN` | **да** | — | DSN PostgreSQL |
| `REDIS_ADDR` | **да** | — | Адрес Redis (host:port) |
| `REDIS_PASSWORD` | нет | пусто | Пароль Redis |
| `REDIS_DB` | нет | `0` | Номер БД Redis |
| `REDIS_DEFAULT_TTL` | нет | `60s` | TTL кеша в Redis |

Конфиг загружается через **cleanenv** из переменных окружения (файл `.env` в коде не подгружается автоматически — при необходимости вызвать `godotenv.Load()` до `config.Load()`).

---

## Аутентификация

- **Регистрация / логин**: пароль хешируется через bcrypt; после успешного входа создаётся сессия в Redis (ключ `session:<id>`, значение — `user_id`).
- **Cookie**: в ответ клиенту выставляется `session_id` (HttpOnly, 24 часа). Все запросы к `/api/v1/todos*` требуют эту куку.
- **Middleware** `RequireSession`: читает куку, по session_id получает user_id из Redis, кладёт user_id в контекст Gin. Без валидной сессии — 401.

---

## Форматы запросов и ответов (примеры)

**Создание задачи** `POST /api/v1/todos`:

```json
{
  "title": "Задача",
  "description": "Описание",
  "due_at": "2026-02-19"
}
```

Поле `due_at` опционально; принимается дата **только** (`YYYY-MM-DD`) или RFC3339 (с временем). В БД хранится как TIMESTAMPTZ.

**Частичное обновление** `PATCH /api/v1/todos/:id` (все поля опциональны):

```json
{
  "title": "Новый заголовок",
  "is_done": true,
  "due_at": "2026-03-01"
}
```

**Ответ задачи** (в списке и по ID): `id`, `title`, `description`, `is_done`, `due_at` (строка RFC3339 или null), `created_at`, `updated_at`.

---

## Миграции

| Файл | Назначение |
|------|------------|
| `00001_create_todos_table.sql` | Таблица `todos` (id, title, description, is_done, due_at, created_at, updated_at, deleted_at). |
| `00002_create_users_table.sql` | Таблица `users` (id, username, password_hash, created_at); дефолтный пользователь admin. |
| `00003_add_user_id_to_todos.sql` | Колонка `user_id` в `todos` (FK на users), индекс, backfill существующих строк. |

Миграции применяются при старте приложения (Goose Up). Откат — вручную или через `goose down`.

---

## Кеш (Redis)

- Кешируются: список задач пользователя, результаты поиска по запросу, список просроченных — с разделением по **user_id** (ключи вида `todo:list:<userID>`, `todo:search:<userID>:<query>`, `todo:overdue:<userID>`).
- TTL задаётся конфигом `REDIS_DEFAULT_TTL` (по умолчанию 60s).
- При любой записи (create/update/delete/complete) для данного пользователя вызывается инвалидация его ключей (list, overdue, все search). Используется **singleflight**, чтобы не дублировать запросы к БД при одновременных одинаковых вызовах.

---

## Запуск

### Локально (без Docker)

```bash
go build -o api ./cmd/api
```

Перед запуском задать переменные окружения (или использовать `.env` через оболочку):

- `PG_DSN` — например `postgres://app:app@localhost:5432/app?sslmode=disable`
- `REDIS_ADDR` — например `localhost:6379`

Затем: `./api` (или `.\api.exe` на Windows). Миграции выполняются автоматически при старте (Goose, каталог `./migrations`).

### Docker Compose

```bash
docker compose up -d --build
```

- **postgres** — порт 5432, БД `app`, пользователь/пароль `app`/`app`.
- **redis** — порт 6379, без пароля.
- **api** — порт 8080, подключается к сервисам `postgres` и `redis` по именам.

Переменные для api переопределяются в `docker-compose.yml` (PG_DSN, REDIS_ADDR). Локальный `.env` подхватывается через `env_file`.

---

## Структура приложения (кратко)

- **cmd/api** — точка входа, загрузка конфига, создание `App`, HTTP-сервер, graceful shutdown.
- **internal/app** — инициализация роутера, регистрация маршрутов, подключение БД/Redis, запуск миграций.
- **internal/config** — структуры конфига и загрузка через cleanenv.
- **internal/handlers** — HTTP-обработчики (auth, todo).
- **internal/service** — бизнес-логика (user, todo).
- **internal/repo** — доступ к PostgreSQL (users, todos).
- **internal/cache** — кеш todos в Redis.
- **internal/auth** — сессии в Redis, middleware проверки сессии.
- **internal/domain**, **internal/dto** — доменные модели и DTO.
- **migrations** — SQL-миграции Goose: `00001_create_todos_table.sql`, `00002_create_users_table.sql`, `00003_add_user_id_to_todos.sql`.
- **docs** — сгенерированный Swagger (команда `swag init`).
- **scripts** — вспомогательные скрипты (например, генерация хеша пароля).

---

## Возможности Gin (используемые в проекте)

1. **Context** — основной объект запроса/ответа, передача данных между middleware и хендлерами.
2. **Binding** — парсинг JSON/form и привязка к структурам с валидацией (validator).
3. **Middleware pipeline** — цепочка: логирование, recovery, проверка сессии для `/api/v1/todos*`.
4. **Recovery** — перехват panic и ответ 500 (Gin по умолчанию).
5. **Группы роутов** — `/api/v1`, отдельная группа для защищённых маршрутов с `RequireSession`.
6. **Content negotiation** — отдача JSON, Swagger JSON.
7. **Пулы объектов** — снижение аллокаций при сериализации.
8. **Логирование** — встроенный логер запросов (Gin default).
9. **Streaming и отмена** — контекст запроса для таймаутов и отмены.

---

## Swagger

- Документ генерируется из аннотаций в коде (пакет `docs`).
- Обновление после изменения аннотаций: `swag init -g cmd/api/main.go`.
- В браузере: `http://localhost:8080/swagger/index.html`.

---

## Дополнительно

- **Каверзные вопросы по коду** (для собеседований): см. [TRICKY_INTERVIEW_QUESTIONS.md](TRICKY_INTERVIEW_QUESTIONS.md).
