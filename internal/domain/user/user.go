package user

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"uuid"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidName        = errors.New("user: name must not be empty")
	ErrInvalidEmail       = errors.New("user: invalid email address")
	ErrWeakPassword       = errors.New("user: password must be at least 8 characters")
	ErrInvalidRole        = errors.New("user: invalid role")
	ErrInvalidCredentials = errors.New("user: invalid credentials")
	ErrUserInactive       = errors.New("user: account is inactive")
	ErrEmailAlreadyExists = errors.New("user: email already exists")
)

const minPasswordLength = 8

type Role string

const (
	RoleOwner Role = "owner"
	RoleClerk Role = "clerk"
)

func (r Role) IsValid() bool {
	return r == RoleOwner || r == RoleClerk
}

var emailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

type User struct {
	id           string
	name         string
	email        string
	passwordHash string
	role         Role
	isActive     bool
	createdAt    time.Time
	updatedAt    time.Time
}

func New(name, email, plainPassword string, role Role) (*User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidName
	}

	email = strings.ToLower(strings.TrimSpace(email))
	if !emailPattern.MatchString(email) {
		return nil, ErrInvalidEmail
	}

	if len(plainPassword) < minPasswordLength {
		return nil, ErrWeakPassword
	}

	if !role.IsValid() {
		return nil, ErrInvalidRole
	}

	id := uuid.NewV7()

	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()

	return &User{
		id:           id.String(),
		name:         name,
		email:        email,
		passwordHash: string(hash),
		role:         role,
		isActive:     true,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

func Restore(
	id, name, email, passwordHash string,
	role Role,
	isActive bool,
	createdAt, updatedAt time.Time,
) *User {
	return &User{
		id:           id,
		name:         name,
		email:        email,
		passwordHash: passwordHash,
		role:         role,
		isActive:     isActive,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

func (u *User) Authenticate(plainPassword string) error {
	if !u.isActive {
		return ErrUserInactive
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.passwordHash), []byte(plainPassword)); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}

func (u *User) ChangePassword(plainPassword string) error {
	if len(plainPassword) < minPasswordLength {
		return ErrWeakPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.passwordHash = string(hash)
	u.touch()
	return nil
}

func (u *User) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrInvalidName
	}
	u.name = name
	u.touch()
	return nil
}

func (u *User) Deactivate() {
	if !u.isActive {
		return
	}
	u.isActive = false
	u.touch()
}

func (u *User) Activate() {
	if u.isActive {
		return
	}
	u.isActive = true
	u.touch()
}

func (u *User) touch() {
	u.updatedAt = time.Now().UTC()
}

func (u *User) ID() string           { return u.id }
func (u *User) PasswordHash() string { return u.passwordHash }
func (u *User) Name() string         { return u.name }
func (u *User) Email() string        { return u.email }
func (u *User) Role() Role           { return u.role }
func (u *User) IsActive() bool       { return u.isActive }
func (u *User) CreatedAt() time.Time { return u.createdAt }
func (u *User) UpdatedAt() time.Time { return u.updatedAt }
