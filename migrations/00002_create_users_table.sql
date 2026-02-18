-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    username      VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);

-- One default user: admin / admin
INSERT INTO users (username, password_hash)
VALUES ('admin', '$2a$10$6m7OKhL8jFLFqqU0MVIB1ONXUtSGYUkjGVe//M1SGfQ6A2/OPXZXu')
ON CONFLICT (username) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS users;
