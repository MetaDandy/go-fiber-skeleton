package authentication

import "github.com/MetaDandy/go-fiber-skeleton/src/model"

type Service interface {
	UserAuthProviders(email string) ([]string, error)
}

type uRepo interface {
	FindByEmail(email string) (model.User, error)
}

type service struct {
	repo  Repo
	uRepo uRepo
}

func NewService(repo Repo, uRepo uRepo) Service {
	return &service{repo: repo, uRepo: uRepo}
}

func (s *service) UserAuthProviders(email string) ([]string, error) {
	user, err := s.uRepo.FindByEmail(email)
	if err != nil {
		return []string{}, err
	}

	providers := s.repo.UserAuthProviders(user.ID.String())
	if user.Password != "" {
		providers = append(providers, "password")
	}

	return providers, nil
}
