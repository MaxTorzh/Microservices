package service

import (
	"gRPC/internal/domain"
	"gRPC/internal/repository/memory"
	"time"

	"github.com/google/uuid"
)

type UserService struct {
	repo *memory.Repository
}

func NewUserService(repo *memory.Repository) UserService {
	return UserService{repo: repo}
}

func (s UserService) CreateUser(req domain.CreateUserRequest) (domain.UserResponse, error) {
	if req.Email == "" || req.Name == "" {
		return domain.UserResponse{}, domain.ErrInvalidRequest
	}

	now := time.Now()
	user := domain.User{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Name:      req.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(user); err != nil {
		return domain.UserResponse{}, err
	}

	return user.ToResponse(), nil
}

func (s UserService) GetUser(id string) (domain.UserResponse, error) {
	if id == "" {
		return domain.UserResponse{}, domain.ErrEmptyID
	}

	user, err := s.repo.GetByID(id)
	if err != nil {
		return domain.UserResponse{}, err
	}
	return user.ToResponse(), nil
}

func (s UserService) GetAllUsers() []domain.UserResponse {
	users := s.repo.GetAll()
	response := make([]domain.UserResponse, len(users))
	for i, user := range users {
		response[i] = user.ToResponse()
	}
	return response
}

func (s UserService) UpdateUser(id string, req domain.UpdateUserRequest) (domain.UserResponse, error) {
	if id == "" {
		return domain.UserResponse{}, domain.ErrEmptyID
	}
	if req.Email == "" || req.Name == "" {
		return domain.UserResponse{}, domain.ErrInvalidRequest
	}

	existingUser, err := s.repo.GetByID(id)
	if err != nil {
		return domain.UserResponse{}, err
	}

	updatedUser := domain.User{
		ID:        existingUser.ID,
		Email:     req.Email,
		Name:      req.Name,
		CreatedAt: existingUser.CreatedAt,
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Update(id, updatedUser); err != nil {
		return domain.UserResponse{}, err
	}

	return updatedUser.ToResponse(), nil
}

func (s UserService) DeleteUser(id string) error {
	if id == "" {
		return domain.ErrEmptyID
	}
	return s.repo.Delete(id)
}