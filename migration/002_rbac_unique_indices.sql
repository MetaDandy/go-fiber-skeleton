-- +goose Up
ALTER TABLE RolePermissions ADD CONSTRAINT uq_role_permission UNIQUE (role_id, permission_id);
ALTER TABLE RoleEffectivePermissions ADD CONSTRAINT uq_role_effective_permission UNIQUE (role_id, permission_id);

-- +goose Down
ALTER TABLE RoleEffectivePermissions DROP CONSTRAINT uq_role_effective_permission;
ALTER TABLE RolePermissions DROP CONSTRAINT uq_role_permission;