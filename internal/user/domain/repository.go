package user

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("user: not found")

type Reader interface {
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	// List returns one page of users, newest first, plus the total number of
	// users so callers can compute how many pages exist.
	List(ctx context.Context, limit, offset int) (users []*User, total int, err error)
}

type Writer interface {
	Save(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
}

type Repository interface {
	Reader
	Writer
}
