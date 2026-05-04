-- +goose Up
-- Enable pg_trgm extension for trigram-based text search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Users: name and email are used in ILIKE '%...%' searches
CREATE INDEX CONCURRENTLY idx_users_name_trgm ON users USING gin (name gin_trgm_ops);
CREATE INDEX CONCURRENTLY idx_users_email_trgm ON users USING gin (email gin_trgm_ops);

-- Roles: name and description are used in ILIKE '%...%' searches
CREATE INDEX CONCURRENTLY idx_roles_name_trgm ON roles USING gin (name gin_trgm_ops);
CREATE INDEX CONCURRENTLY idx_roles_description_trgm ON roles USING gin (description gin_trgm_ops);

-- Permissions: name and description are used in ILIKE '%...%' searches
CREATE INDEX CONCURRENTLY idx_permissions_name_trgm ON permissions USING gin (name gin_trgm_ops);
CREATE INDEX CONCURRENTLY idx_permissions_description_trgm ON permissions USING gin (description gin_trgm_ops);

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS idx_permissions_description_trgm;
DROP INDEX CONCURRENTLY IF EXISTS idx_permissions_name_trgm;
DROP INDEX CONCURRENTLY IF EXISTS idx_roles_description_trgm;
DROP INDEX CONCURRENTLY IF EXISTS idx_roles_name_trgm;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_email_trgm;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_name_trgm;
