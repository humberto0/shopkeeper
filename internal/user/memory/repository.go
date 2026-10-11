package usermem

import (
	"cmp"
	"context"
	"slices"
	"strings"
	"sync"

	user "github.com/humberto0/shopkeeper/internal/user/domain"
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

	r.byID[u.ID()] = clone(u)
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

	if existing.Version() != u.Version() {
		return user.ErrConflict
	}

	if existing.Email() != u.Email() {
		if _, taken := r.byEmail[u.Email()]; taken {
			return user.ErrEmailAlreadyExists
		}
		delete(r.byEmail, existing.Email())
		r.byEmail[u.Email()] = u.ID()
	}

	u.SetVersion(existing.Version() + 1)
	r.byID[u.ID()] = clone(u)
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.byID[id]
	if !ok {
		return nil, user.ErrNotFound
	}
	return clone(u), nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.byEmail[normalizeEmail(email)]
	if !ok {
		return nil, user.ErrNotFound
	}
	return clone(r.byID[id]), nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.byEmail[normalizeEmail(email)]
	return ok, nil
}

// List mirrors the Postgres ordering (created_at DESC, id DESC) so use case
// tests see the same pages they would against the real database.
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*user.User, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	all := make([]*user.User, 0, len(r.byID))
	for _, u := range r.byID {
		all = append(all, u)
	}
	slices.SortFunc(all, func(a, b *user.User) int {
		if c := b.CreatedAt().Compare(a.CreatedAt()); c != 0 {
			return c
		}
		return cmp.Compare(b.ID(), a.ID())
	})

	total := len(all)
	if offset >= total {
		return []*user.User{}, total, nil
	}
	end := min(offset+limit, total)

	page := make([]*user.User, 0, end-offset)
	for _, u := range all[offset:end] {
		page = append(page, clone(u))
	}
	return page, total, nil
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

// clone copies a user so callers never share the stored pointer: changes
// only reach the repository through Save or Update, as with Postgres.
func clone(u *user.User) *user.User {
	return user.Restore(
		u.ID(), u.Name(), u.Email(), u.PasswordHash(),
		u.Role(), u.IsActive(), u.Version(),
		u.CreatedAt(), u.UpdatedAt(),
	)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
