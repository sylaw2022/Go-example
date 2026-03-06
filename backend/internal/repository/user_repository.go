package repository

import (
	"sync"

	"github.com/sylaw/fullstack-app/internal/domain"
)

// UserRepository interface defines data operations for a User
type UserRepository interface {
	GetAll() ([]domain.User, error)
	GetByID(id int) (domain.User, error)
}

// InMemoryUserRepository is a simple in-memory implementation of UserRepository
type InMemoryUserRepository struct {
	users map[int]domain.User
	mu    sync.RWMutex
}

// NewInMemoryUserRepository creates a new repository pre-populated with some data
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: map[int]domain.User{
			1: {ID: 1, Name: "Alice Smith", Email: "alice@example.com"},
			2: {ID: 2, Name: "Bob Jones", Email: "bob@example.com"},
			3: {ID: 3, Name: "Charlie Brown", Email: "charlie@example.com"},
		},
	}
}

func (r *InMemoryUserRepository) GetAll() ([]domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.User
	for _, user := range r.users {
		result = append(result, user)
	}
	return result, nil
}

func (r *InMemoryUserRepository) GetByID(id int) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return domain.User{}, domain.ErrUserNotFound
	}
	return user, nil
}
