package service

import (
	"JWT/internal/domain"
	"JWT/internal/repository/memory"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

type UserService struct {
	repo *memory.Repository
}

func NewUserService(repo *memory.Repository) UserService {
	return UserService{repo: repo}
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func checkPassword(password, hash string) bool {
	return hashPassword(password) == hash
}

func (s UserService) CreateUser(req domain.CreateUserRequest) (domain.UserResponse, error) {
	if req.Email == "" || req.Name == "" || req.Password == "" {
		return domain.UserResponse{}, domain.ErrInvalidRequest
	}

	now := time.Now()
	user := domain.User{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Password:  hashPassword(req.Password), // Хешируем пароль
		Name:      req.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(user); err != nil {
		return domain.UserResponse{}, err
	}

	return user.ToResponse(), nil
}

func (s UserService) Login(req domain.LoginRequest) (domain.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return domain.LoginResponse{}, domain.ErrInvalidRequest
	}

	user, err := s.repo.GetByEmail(req.Email)
	if err != nil {
		return domain.LoginResponse{}, domain.ErrInvalidPassword
	}

	if !checkPassword(req.Password, user.Password) {
		return domain.LoginResponse{}, domain.ErrInvalidPassword
	}

	// Возвращение данных пользователя (токен будет добавлен в handler)
	return domain.LoginResponse{
		Token: "", // Токен будет сгенерирован в handler
	}, nil
}

func (s UserService) GetUserByID(id string) (domain.UserResponse, error) {
	if id == "" {
		return domain.UserResponse{}, domain.ErrEmptyID
	}

	user, err := s.repo.GetByID(id)
	if err != nil {
		return domain.UserResponse{}, err
	}
	return user.ToResponse(), nil
}

func (s UserService) GetUserByEmail(email string) (domain.UserResponse, error) {
	if email == "" {
		return domain.UserResponse{}, domain.ErrInvalidRequest
	}

	user, err := s.repo.GetByEmail(email)
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
		Password:  existingUser.Password, // Сохраняем старый пароль
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