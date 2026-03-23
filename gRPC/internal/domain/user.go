package domain

import (
	"errors"
	"time"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrEmailExists    = errors.New("email already exists")
	ErrEmptyID        = errors.New("empty user ID")
	ErrInvalidRequest = errors.New("invalid request")
)

type User struct {
	ID        string
	Email     string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateUserRequest struct {
	Email string
	Name  string
}

type UpdateUserRequest struct {
	Email string
	Name  string
}

type UserResponse struct {
	ID        string
	Email     string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
