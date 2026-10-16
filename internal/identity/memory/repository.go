package identitymem

import (
	"cmp"
	"context"
	"slices"
	"strings"
	"sync"

	identity "github.com/humberto0/shopkeeper/internal/identity/domain"
)

var _ identity.Repository = (*UserRepository)(nil)

type UserRepository struct {
	mu      sync.RWMutex
	byID    map[string]*identity.User
	byEmail map[string]string // normalized email → id
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		byID:    make(map[string]*identity.User),
		byEmail: make(map[string]string),
	}
}

func (r *UserRepository) Save(ctx context.Context, u *identity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byEmail[u.Email()]; exists {
		return identity.ErrEmailAlreadyExists
	}

	r.byID[u.ID()] = clone(u)
	r.byEmail[u.Email()] = u.ID()
	return nil
}

func (r *UserRepository) Update(ctx context.Context, u *identity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.byID[u.ID()]
	if !ok {
		return identity.ErrNotFound
	}

	if existing.Version() != u.Version() {
		return identity.ErrConflict
	}

	if existing.Email() != u.Email() {
		if _, taken := r.byEmail[u.Email()]; taken {
			return identity.ErrEmailAlreadyExists
		}
		delete(r.byEmail, existing.Email())
		r.byEmail[u.Email()] = u.ID()
	}

	u.SetVersion(existing.Version() + 1)
	r.byID[u.ID()] = clone(u)
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*identity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.byID[id]
	if !ok {
		return nil, identity.ErrNotFound
	}
	return clone(u), nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*identity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.byEmail[normalizeEmail(email)]
	if !ok {
		return nil, identity.ErrNotFound
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
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*identity.User, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	all := make([]*identity.User, 0, len(r.byID))
	for _, u := range r.byID {
		all = append(all, u)
	}
	slices.SortFunc(all, func(a, b *identity.User) int {
		if c := b.CreatedAt().Compare(a.CreatedAt()); c != 0 {
			return c
		}
		return cmp.Compare(b.ID(), a.ID())
	})

	total := len(all)
	if offset >= total {
		return []*identity.User{}, total, nil
	}
	end := min(offset+limit, total)

	page := make([]*identity.User, 0, end-offset)
	for _, u := range all[offset:end] {
		page = append(page, clone(u))
	}
	return page, total, nil
}

func (r *UserRepository) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.byID = make(map[string]*identity.User)
	r.byEmail = make(map[string]string)
}

func (r *UserRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.byID)
}

// clone copies a user so callers never share the stored pointer: changes
// only reach the repository through Save or Update, as with Postgres.
func clone(u *identity.User) *identity.User {
	return identity.Restore(
		u.ID(), u.Name(), u.Email(), u.PasswordHash(),
		u.Role(), u.IsActive(), u.Version(),
		u.CreatedAt(), u.UpdatedAt(),
	)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
