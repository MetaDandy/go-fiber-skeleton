package authtentication

type Service interface {
	UserAuthMethods(email string) []string
}
