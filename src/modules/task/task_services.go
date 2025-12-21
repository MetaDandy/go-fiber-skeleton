package task

import (
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/enum"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/MetaDandy/go-fiber-skeleton/src/response"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
)

type TaskService interface {
	Create(input CreateTaskDto) (*response.TaskResponse, error)
	FindByID(id string) (*response.TaskResponse, error)
	FindAll(opts *helper.FindAllOptions) (*response.Paginated[response.TaskResponse], error)
	Update(id string, input UpdateTaskDto) (*response.TaskResponse, error)
	Delete(id string) error
}

type Service struct {
	repo TaskRepo
}

func NewService(repo TaskRepo) TaskService {
	return &Service{repo: repo}
}

func (s *Service) Create(input CreateTaskDto) (*response.TaskResponse, error) {
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		return nil, err
	}

	task := model.Task{
		ID:          uuid.New(),
		Title:       input.Title,
		Description: input.Description,
		Status:      enum.StatusEnum(input.Status),
		UserID:      userID,
	}

	if err := s.repo.Create(task); err != nil {
		return nil, err
	}

	dto := response.TaskToDto(&task)
	return &dto, nil
}

func (s *Service) FindByID(id string) (*response.TaskResponse, error) {
	task, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	dto := response.TaskToDto(&task)

	return &dto, nil
}

func (s *Service) FindAll(opts *helper.FindAllOptions) (*response.Paginated[response.TaskResponse], error) {
	finded, total, err := s.repo.FindAll(opts)
	if err != nil {
		return nil, err
	}
	dtos := response.TaskToListDto(finded)
	pages := uint((total + int64(opts.Limit) - 1) / int64(opts.Limit))

	return &response.Paginated[response.TaskResponse]{
		Data:   dtos,
		Total:  total,
		Limit:  opts.Limit,
		Offset: opts.Offset,
		Pages:  pages,
	}, nil
}

func (s *Service) Update(id string, input UpdateTaskDto) (*response.TaskResponse, error) {
	task, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	opt := copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}

	if err := copier.CopyWithOption(&task, &input, opt); err != nil {
		return nil, err
	}

	if err := s.repo.Update(task); err != nil {
		return nil, err
	}

	dto := response.TaskToDto(&task)
	return &dto, nil
}

func (s *Service) Delete(id string) error {
	return s.repo.Delete(id)
}
