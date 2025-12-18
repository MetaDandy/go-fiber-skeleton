-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE
);

CREATE TYPE status_enum AS ENUM ('pending', 'in_progress', 'done');

CREATE TABLE task (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    status status_enum NOT NULL DEFAULT 'pending',
    user_id UUID,
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE SET NULL
);

-- +goose Down
DROP TABLE IF EXISTS task;
DROP TYPE IF EXISTS status_enum;
DROP TABLE IF EXISTS users;