package user

import (
	"context"
	"fmt"

	"github.com/humberto0/shopkeeper/internal/domain/user"
)

type registerUserRepository interface {
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	Save(ctx context.Context, u *user.User) error
}

type RegisterUserInput struct {
	Name     string
	Email    string
	Password string
	Role     user.Role
}

type RegisterUserOutput struct {
	ID    string
	Name  string
	Email string
	Role  user.Role
}

type RegisterUser struct {
	repo registerUserRepository
}

func NewRegisterUser(repo registerUserRepository) *RegisterUser {
	return &RegisterUser{repo: repo}
}

func (uc *RegisterUser) Execute(ctx context.Context, in RegisterUserInput) (*RegisterUserOutput, error) {
	exists, err := uc.repo.ExistsByEmail(ctx, in.Email)
	if err != nil {
		return nil, fmt.Errorf("checking if email exists: %w", err)
	}
	if exists {
		return nil, user.ErrEmailAlreadyExists
	}

	u, err := user.New(in.Name, in.Email, in.Password, in.Role)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Save(ctx, u); err != nil {
		return nil, fmt.Errorf("saving user: %w", err)
	}

	return &RegisterUserOutput{
		ID:    u.ID(),
		Name:  u.Name(),
		Email: u.Email(),
		Role:  u.Role(),
	}, nil
}
