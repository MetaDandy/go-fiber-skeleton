package seed

import (
	"testing"

	"github.com/MetaDandy/go-fiber-skeleton/src/enum"
)

func TestPermissionModules_HasCorrectCount(t *testing.T) {
	if len(permissionModules) != 19 {
		t.Errorf("expected 19 permissions, got %d", len(permissionModules))
	}
}

func TestPermissionModules_ContainsAllExpectedPermissions(t *testing.T) {
	expectedPermissions := map[enum.Permission]bool{
		// Existing (5)
		enum.PermissionRead:   false,
		enum.PermissionList:    false,
		enum.RoleCreate:        false,
		enum.UserPermissionCreate: false,
		enum.UserPermissionRead:    false,

		// New (14)
		enum.PermissionCreate:     false,
		enum.UserCreate:           false,
		enum.UserUpdate:           false,
		enum.UserDelete:           false,
		enum.UserList:             false,
		enum.RoleUpdate:           false,
		enum.RoleList:             false,
		enum.SessionList:          false,
		enum.SessionRevoke:        false,
		enum.SessionRevokeAll:     false,
		enum.AuthLogRead:          false,
		enum.AuthLogList:          false,
		enum.UserPermissionDelete: false,
		enum.UserPermissionList:   false,
	}

	for _, p := range permissionModules {
		if _, exists := expectedPermissions[enum.Permission(p.ID)]; !exists {
			t.Errorf("unexpected permission found: %s", p.ID)
		} else {
			expectedPermissions[enum.Permission(p.ID)] = true
		}
	}

	for perm, found := range expectedPermissions {
		if !found {
			t.Errorf("missing expected permission: %s", perm)
		}
	}
}

func TestPermissionModules_NoDuplicates(t *testing.T) {
	seen := make(map[string]bool)
	for _, p := range permissionModules {
		if seen[p.ID] {
			t.Errorf("duplicate permission found: %s", p.ID)
		}
		seen[p.ID] = true
	}
}
