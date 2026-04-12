package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/serhatkilbas/lms-poc/internal/domain"
)

type memoryUserRepository struct {
	users map[string]domain.User
	mutex sync.RWMutex
}

func NewMemoryUserRepository() domain.UserRepository {
	return &memoryUserRepository{
		users: make(map[string]domain.User),
	}
}

func (r *memoryUserRepository) Create(ctx context.Context, user domain.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.users[user.Email]; exists {
		return errors.New("user already exists")
	}

	r.users[user.Email] = user
	return nil
}

func (r *memoryUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	user, exists := r.users[email]
	if !exists {
		return nil, errors.New("user not found")
	}

	return &user, nil
}
