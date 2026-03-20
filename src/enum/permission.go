package enum

type Permission string

var (
	PermissionRead   Permission = "permission.read"
	PermissionList   Permission = "permission.list"
	PermissionCreate Permission = "permission.create"

	UserCreate Permission = "user.create"
	UserUpdate Permission = "user.update"
	UserDelete Permission = "user.delete"
	UserList   Permission = "user.list"

	RoleCreate Permission = "role.create"
	RoleUpdate Permission = "role.update"
	RoleList   Permission = "role.list"

	SessionList      Permission = "session.list"
	SessionRevoke    Permission = "session.revoke"
	SessionRevokeAll Permission = "session.revokeAll"

	AuthLogRead Permission = "authLog.read"
	AuthLogList Permission = "authLog.list"
)

func (m Permission) IsValid() bool {
	switch m {
	case PermissionRead, PermissionList, PermissionCreate,
		UserCreate, UserUpdate, UserDelete, UserList,
		RoleCreate, RoleUpdate, RoleList,
		SessionList, SessionRevoke, SessionRevokeAll,
		AuthLogRead, AuthLogList:

		return true
	}
	return false
}

func (m Permission) String() string {
	return string(m)
}

func PermissionToArray() []string {
	return []string{
		string(PermissionRead),
		string(PermissionList),
		string(PermissionCreate),
		string(UserCreate),
		string(UserUpdate),
		string(UserDelete),
		string(UserList),
		string(RoleCreate),
		string(RoleUpdate),
		string(RoleList),
		string(SessionList),
		string(SessionRevoke),
		string(SessionRevokeAll),
		string(AuthLogRead),
		string(AuthLogList),
	}
}
