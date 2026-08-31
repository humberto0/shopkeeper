package postgres

import (
	"time"

	"github.com/humberto0/shopkeeper/internal/domain/user"
)

type userMapper struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
	Role         string
	IsActive     bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m userMapper) ToDomainUser() *user.User {
	return user.Restore(
		m.ID, m.Name, m.Email, m.PasswordHash, user.Role(m.Role), m.IsActive, m.CreatedAt, m.UpdatedAt,
	)
}
