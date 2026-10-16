package identityapp

import (
	"context"
	"fmt"
	"time"

	identity "github.com/humberto0/shopkeeper/internal/identity/domain"
	"github.com/humberto0/shopkeeper/internal/platform/pagination"
)

type listUsersRepository interface {
	List(ctx context.Context, limit, offset int) ([]*identity.User, int, error)
}

type ListUsersInput struct {
	Page  int
	Limit int
}

// ListUsersItem is the read model of one user in a listing; it never carries
// the password hash.
type ListUsersItem struct {
	ID        string
	Name      string
	Email     string
	Role      identity.Role
	IsActive  bool
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ListUsersOutput = pagination.OffsetPage[ListUsersItem]

type ListUsers struct {
	repo listUsersRepository
}

func NewListUsers(repo listUsersRepository) *ListUsers {
	return &ListUsers{repo: repo}
}

// Execute normalizes the requested page (page < 1 → 1, limit clamped to
// 1..100, default 20) instead of rejecting it, so clients can always send
// whatever the UI has and get a valid page back.
func (uc *ListUsers) Execute(ctx context.Context, in ListUsersInput) (*ListUsersOutput, error) {
	p := pagination.OffsetPagination{Page: in.Page, Limit: in.Limit}.Normalize()

	users, total, err := uc.repo.List(ctx, p.Limit, p.Offset())
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}

	items := make([]ListUsersItem, 0, len(users))
	for _, u := range users {
		items = append(items, ListUsersItem{
			ID:        u.ID(),
			Name:      u.Name(),
			Email:     u.Email(),
			Role:      u.Role(),
			IsActive:  u.IsActive(),
			Version:   u.Version(),
			CreatedAt: u.CreatedAt(),
			UpdatedAt: u.UpdatedAt(),
		})
	}

	page := pagination.NewOffsetPage(items, p, total)
	return &page, nil
}
