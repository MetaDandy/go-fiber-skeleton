package enum

type Event string

const (
	SignUpSuccess         Event = "signup_success"
	LoginSuccess          Event = "login_success"
	LoginFailed           Event = "login_failed"
	EmailNotVerified      Event = "email_not_verified"
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
	case LoginSuccess, LoginFailed, EmailNotVerified, PasswordReset, PasswordResetSuccess, PasswordChangeSuccess, PasswordChangeFailure, OAuthLogin, SessionRefresh, Logout:
		return true
	}
	return false
}

func (m Event) String() string {
	return string(m)
}
