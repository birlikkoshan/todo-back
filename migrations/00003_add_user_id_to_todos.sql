-- +goose Up
ALTER TABLE todos ADD COLUMN user_id BIGINT REFERENCES users (id);
-- Backfill: assign existing rows to first user (e.g. admin)
UPDATE todos SET user_id = (SELECT id FROM users ORDER BY id ASC LIMIT 1) WHERE user_id IS NULL;
ALTER TABLE todos ALTER COLUMN user_id SET NOT NULL;
CREATE INDEX IF NOT EXISTS idx_todos_user_id ON todos (user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_todos_user_id;
ALTER TABLE todos DROP COLUMN user_id;
