package user

import (
	"context"
	"time"

	"uuid"

	"github.com/humberto0/shopkeeper/internal/domain/user"
)

type findUserRepository interface {
	FindByID(ctx context.Context, id string) (*user.User, error)
}

type FindUserByIDResult struct {
	ID        string
	Name      string
	Email     string
	Role      user.Role
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FindUserByID struct {
	repo findUserRepository
}

func NewFindUserByID(repo findUserRepository) *FindUserByID {
	return &FindUserByID{repo: repo}
}

func (uc *FindUserByID) Execute(ctx context.Context, id string) (*FindUserByIDResult, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, user.ErrInvalidID
	}

	find, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &FindUserByIDResult{
		ID:        find.ID(),
		Name:      find.Name(),
		Email:     find.Email(),
		Role:      find.Role(),
		IsActive:  find.IsActive(),
		CreatedAt: find.CreatedAt(),
		UpdatedAt: find.UpdatedAt(),
	}, nil
}
