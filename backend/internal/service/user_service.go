package service

import (
	"github.com/sylaw/fullstack-app/internal/domain"
	"github.com/sylaw/fullstack-app/internal/repository"
)

// UserService defines the business logic operations for users
type UserService interface {
	GetUsers() ([]domain.User, error)
	GetUser(id int) (domain.User, error)
}

// userServiceImpl implements UserService
type userServiceImpl struct {
	repo repository.UserRepository
}

// NewUserService creates a new user service with the injected repository
func NewUserService(repo repository.UserRepository) UserService {
	return &userServiceImpl{
		repo: repo,
	}
}

func (s *userServiceImpl) GetUsers() ([]domain.User, error) {
	return s.repo.GetAll()
}

func (s *userServiceImpl) GetUser(id int) (domain.User, error) {
	return s.repo.GetByID(id)
}
