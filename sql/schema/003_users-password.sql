-- +goose Up
ALTER TABLE users add COLUMN hashed_password TEXT NOT NULL DEFAULT 'unset';

-- +goose Down
ALTER TABLE users DROP COLUMN hashed_password;