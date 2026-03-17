-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE Roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    role_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (role_id) REFERENCES Roles(id) ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE TABLE Permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    description TEXT,
    name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    email_verified BOOLEAN NOT NULL DEFAULT false,
    email_verified_at TIMESTAMP WITH TIME ZONE,
    phone TEXT,
    password TEXT,
    picture TEXT,
    role_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (role_id) REFERENCES Roles(id) ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE TABLE RolePermissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID,
    permission_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (role_id) REFERENCES Roles(id) ON UPDATE CASCADE ON DELETE SET NULL,
    FOREIGN KEY (permission_id) REFERENCES Permissions(id) ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE TABLE RoleEffectivePermissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID,
    permission_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (role_id) REFERENCES Roles(id) ON UPDATE CASCADE ON DELETE SET NULL,
    FOREIGN KEY (permission_id) REFERENCES Permissions(id) ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE TABLE UserPermissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    permission_id UUID,
    user_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (permission_id) REFERENCES Permissions(id) ON UPDATE CASCADE ON DELETE SET NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE TABLE AuthProviders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider TEXT,
    provider_user_id TEXT,
    user_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE TABLE AuthLogs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event TEXT,
    user_agent TEXT,
    ip TEXT,
    user_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE TABLE EmailVerificationTokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash TEXT,
    used_at TIMESTAMP WITH TIME ZONE,
    user_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE TABLE PasswordResetTokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash TEXT,
    used_at TIMESTAMP WITH TIME ZONE,
    user_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider TEXT,
    refresh_token_hash TEXT,
    expires_at TEXT,
    ip TEXT,
    user_agent TEXT,
    revoked_at TEXT,
    user_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE TABLE UserRoles (
    user_id UUID NOT NULL,
    role_id UUID NOT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES Roles(id) ON UPDATE CASCADE ON DELETE CASCADE
);


-- Indexes
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_roles_deleted_at ON Roles(deleted_at);
CREATE INDEX idx_permissions_deleted_at ON Permissions(deleted_at);
CREATE INDEX idx_sessions_deleted_at ON sessions(deleted_at);

-- +goose Down
DROP TABLE IF EXISTS UserRoles;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS PasswordResetTokens;
DROP TABLE IF EXISTS EmailVerificationTokens;
DROP TABLE IF EXISTS AuthLogs;
DROP TABLE IF EXISTS AuthProviders;
DROP TABLE IF EXISTS UserPermissions;
DROP TABLE IF EXISTS RoleEffectivePermissions;
DROP TABLE IF EXISTS RolePermissions;
DROP TABLE IF EXISTS Permissions;
DROP TABLE IF EXISTS Roles;
DROP TABLE IF EXISTS users;