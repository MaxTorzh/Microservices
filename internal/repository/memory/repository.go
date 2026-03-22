package memory

import (
	"errors"
	"simple_microservice/internal/domain"
	"sync"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrMailExists   = errors.New("email already exists")
)

// Реализация in-memory хранилища пользователей
type Repository struct {
	mu     sync.RWMutex
	users  map[string]domain.User
	emails map[string]string
}

// NewRepository конструктор factory function
func NewRepository() *Repository {
	return &Repository{
		users:  make(map[string]domain.User),
		emails: make(map[string]string),
	}
}

func (r *Repository) Create(user domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.emails[user.Email]; exists {
		return ErrMailExists
	}

	r.users[user.ID] = user
	r.emails[user.Email] = user.ID
	return nil
}

func (r *Repository) GetByID(id string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return domain.User{}, ErrUserNotFound
	}
	return user, nil
}

func (r *Repository) GetAll() []domain.User {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]domain.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}
	return users
}

func (r *Repository) Update(id string, updatedUser domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[id]
	if !exists {
		return ErrUserNotFound
	}

	if user.Email != updatedUser.Email {
		if existingID, exists := r.emails[updatedUser.Email]; exists && existingID != id {
			return ErrMailExists
		}

		delete(r.emails, user.Email)
		r.emails[updatedUser.Email] = id
	}

	updatedUser.CreatedAt = user.CreatedAt
	r.users[id] = updatedUser
	return nil
}

func (r *Repository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[id]
	if !exists {
		return ErrUserNotFound
	}

	delete(r.emails, user.Email)
	delete(r.users, id)
	return nil
}
