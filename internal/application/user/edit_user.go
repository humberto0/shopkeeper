package user

import (
	"context"
	"fmt"
	"strings"

	"uuid"

	"github.com/humberto0/shopkeeper/internal/domain/user"
)

type editUserRepository interface {
	FindByID(ctx context.Context, id string) (*user.User, error)
	Update(ctx context.Context, u *user.User) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type EditUserInput struct {
	ID    string
	Name  string
	Email string
	Role  user.Role
}

type EditUserOutput struct {
	ID    string
	Name  string
	Email string
	Role  user.Role
}

type EditUser struct {
	repo editUserRepository
}

func NewEditUser(repo editUserRepository) *EditUser {
	return &EditUser{repo: repo}
}

func (e *EditUser) Execute(ctx context.Context, in EditUserInput) (*EditUserOutput, error) {
	if _, err := uuid.Parse(in.ID); err != nil {
		return nil, user.ErrInvalidID
	}

	u, err := e.repo.FindByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	if in.Email != "" && !strings.EqualFold(u.Email(), in.Email) {
		exists, err := e.repo.ExistsByEmail(ctx, in.Email)
		if err != nil {
			return nil, fmt.Errorf("checking if email exists: %w", err)
		}
		if exists {
			return nil, user.ErrEmailAlreadyExists
		}
		if err := u.ChangeEmail(in.Email); err != nil {
			return nil, err
		}
	}

	if err := u.Rename(in.Name); err != nil {
		return nil, err
	}

	if err := u.ChangeRole(in.Role); err != nil {
		return nil, err
	}

	if err := e.repo.Update(ctx, u); err != nil {
		return nil, fmt.Errorf("updating user: %w", err)
	}

	return &EditUserOutput{
		ID:    u.ID(),
		Name:  u.Name(),
		Email: u.Email(),
		Role:  u.Role(),
	}, nil
}
