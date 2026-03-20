package authentication

import (
	"fmt"

	"github.com/MetaDandy/go-fiber-skeleton/constant"
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/enum"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/google/uuid"
)

type Service interface {
	UserAuthProviders(email string) ([]string, error)
	SignUpPassword(input SignUpPassword) error
}

type uRepo interface {
	FindByEmail(email string) (model.User, error)
	ExistsByEmail(email string) error
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

func (s *service) SignUpPassword(input SignUpPassword) error {
	if err := s.uRepo.ExistsByEmail(input.Email); err == nil {
		return fmt.Errorf("%s already exist", input.Email)
	}

	if err := input.Validate(); err != nil {
		return err
	}

	hash, err := helper.HashPassword(input.Password)
	if err != nil {
		return err
	}

	u := model.User{
		ID:            uuid.New(),
		Email:         input.Email,
		Password:      hash,
		EmailVerified: false,
		RoleID:        constant.GenericID,
	}

	al := model.AuthLog{
		ID:        uuid.New(),
		Event:     enum.SignUpSuccess,
		UserID:    u.ID,
		Ip:        input.Ip,
		UserAgent: input.UserAgent,
	}

	if err := s.repo.Create(u, al, nil); err != nil {
		return err
	}

	return nil
}
