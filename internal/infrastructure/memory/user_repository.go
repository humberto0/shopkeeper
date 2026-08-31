package memory

import (
	"context"
	"strings"
	"sync"

	"github.com/humberto0/shopkeeper/internal/domain/user"
)

var _ user.Repository = (*UserRepository)(nil)

type UserRepository struct {
	mu      sync.RWMutex
	byID    map[string]*user.User
	byEmail map[string]string // normalized email → id
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		byID:    make(map[string]*user.User),
		byEmail: make(map[string]string),
	}
}

func (r *UserRepository) Save(ctx context.Context, u *user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byEmail[u.Email()]; exists {
		return user.ErrEmailAlreadyExists
	}

	r.byID[u.ID()] = u
	r.byEmail[u.Email()] = u.ID()
	return nil
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.byID[u.ID()]
	if !ok {
		return user.ErrNotFound
	}

	if existing.Email() != u.Email() {
		if _, taken := r.byEmail[u.Email()]; taken {
			return user.ErrEmailAlreadyExists
		}
		delete(r.byEmail, existing.Email())
		r.byEmail[u.Email()] = u.ID()
	}

	r.byID[u.ID()] = u
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.byID[id]
	if !ok {
		return nil, user.ErrNotFound
	}
	return u, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.byEmail[normalizeEmail(email)]
	if !ok {
		return nil, user.ErrNotFound
	}
	return r.byID[id], nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.byEmail[normalizeEmail(email)]
	return ok, nil
}

func (r *UserRepository) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.byID = make(map[string]*user.User)
	r.byEmail = make(map[string]string)
}

func (r *UserRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.byID)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
