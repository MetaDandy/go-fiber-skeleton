package enum

type Event string

const (
	SignUpSuccess  Event = "signup_success"
	LoginSuccess   Event = "login_success"
	LoginFailed    Event = "login_failed"
	PasswordReset  Event = "password_reset"
	OAuthLogin     Event = "oauth_login"
	SessionRefresh Event = "session_refresh"
	Logout         Event = "logout"
)

func (m Event) IsValid() bool {
	switch m {
	case LoginSuccess, LoginFailed, PasswordReset, OAuthLogin, SessionRefresh, Logout:
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
		string(OAuthLogin),
		string(SessionRefresh),
		string(Logout),
	}
}
