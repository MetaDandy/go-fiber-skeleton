package user

import (
	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/MetaDandy/go-fiber-skeleton/src/response"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
)

type Service interface {
	Create(input Create) *api_error.Error
	FindByID(id string) (*response.User, *api_error.Error)
	FindAll(opts *helper.FindAllOptions) (*response.Paginated[response.User], *api_error.Error)
	Update(id string, input Update) *api_error.Error
	Delete(id string) *api_error.Error
}

type service struct {
	repo Repo
}

func NewService(repo Repo) Service {
	return &service{repo: repo}
}

func (s *service) Create(input Create) *api_error.Error {
	user := model.User{}
	copier.Copy(&user, &input)
	user.ID = uuid.New()

	if err := s.repo.Create(user); err != nil {
		return api_error.InternalServerError("Could not create user").WithErr(err)
	}

	return nil
}

func (s *service) FindByID(id string) (*response.User, *api_error.Error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, api_error.NotFound("User not found").WithErr(err)
	}
	dto := response.UserToDto(&user)

	return &dto, nil
}

func (s *service) FindAll(opts *helper.FindAllOptions) (*response.Paginated[response.User], *api_error.Error) {
	finded, total, err := s.repo.FindAll(opts)
	if err != nil {
		return nil, api_error.InternalServerError("Could not retrieve users").WithErr(err)
	}
	dtos := response.UserToListDto(finded)
	pages := uint((total + int64(opts.Limit) - 1) / int64(opts.Limit))

	return &response.Paginated[response.User]{
		Data:   dtos,
		Total:  total,
		Limit:  opts.Limit,
		Offset: opts.Offset,
		Pages:  pages,
	}, nil
}

func (s *service) Update(id string, input Update) *api_error.Error {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return api_error.NotFound("User not found").WithErr(err)
	}

	opt := copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}

	if err := copier.CopyWithOption(&user, &input, opt); err != nil {
		return api_error.InternalServerError("Error processing update").WithErr(err)
	}

	if err := s.repo.Update(user); err != nil {
		return api_error.InternalServerError("Could not update user").WithErr(err)
	}

	return nil
}

func (s *service) Delete(id string) *api_error.Error {
	if err := s.repo.Delete(id); err != nil {
		return api_error.InternalServerError("Could not delete user").WithErr(err)
	}
	return nil
}
