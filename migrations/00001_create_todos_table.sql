-- +goose Up
CREATE TABLE IF NOT EXISTS todos (
    id         BIGSERIAL PRIMARY KEY,
    title      VARCHAR(500) NOT NULL,
    description TEXT,
    is_done    BOOLEAN NOT NULL DEFAULT FALSE,
    due_at     TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_todos_deleted_at ON todos (deleted_at);

-- +goose Down
DROP TABLE IF EXISTS todos;
