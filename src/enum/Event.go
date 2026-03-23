package enum

type Event string

const (
	SignUpSuccess         Event = "signup_success"
	LoginSuccess          Event = "login_success"
	LoginFailed           Event = "login_failed"
	PasswordReset         Event = "password_reset"
	PasswordResetSuccess  Event = "password_reset_success"
	PasswordChangeSuccess Event = "password_change_success"
	PasswordChangeFailure Event = "password_change_failure"
	OAuthLogin            Event = "oauth_login"
	SessionRefresh        Event = "session_refresh"
	Logout                Event = "logout"
)

func (m Event) IsValid() bool {
	switch m {
	case LoginSuccess, LoginFailed, PasswordReset, PasswordResetSuccess, PasswordChangeSuccess, PasswordChangeFailure, OAuthLogin, SessionRefresh, Logout:
		return true
	}
	return false
}

func (m Event) String() string {
	return string(m)
}

func EventToArray() []string {
	return []string{
		string(LoginSuccess),
		string(LoginFailed),
		string(PasswordReset),
		string(PasswordResetSuccess),
		string(PasswordChangeSuccess),
		string(PasswordChangeFailure),
		string(OAuthLogin),
		string(SessionRefresh),
		string(Logout),
	}
}
