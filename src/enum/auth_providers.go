package enum

type AuthProvider string

const (
	Google  AuthProvider = "google"
	Github  AuthProvider = "github"
	Discord AuthProvider = "discord"
)

func (a AuthProvider) String() string {
	return string(a)
}

func IsValidAuthProvider(provider string) bool {
	switch provider {
	case Google.String(), Github.String(), Discord.String():
		return true
	default:
		return false
	}
}
