package user

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"uuid"

	"github.com/humberto0/shopkeeper/internal/domain/shared"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidName        = errors.New("user: name must not be empty")
	ErrInvalidEmail       = errors.New("user: invalid email address")
	ErrWeakPassword       = errors.New("user: password must be at least 8 characters")
	ErrInvalidRole        = errors.New("user: invalid role")
	ErrInvalidID          = errors.New("user: invalid id")
	ErrInvalidCredentials = errors.New("user: invalid credentials")
	ErrUserInactive       = errors.New("user: account is inactive")
	ErrEmailAlreadyExists = errors.New("user: email already exists")
	ErrConflict           = errors.New("user: version conflict")
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
	version      int
	createdAt    time.Time
	updatedAt    time.Time
}

func ValidateUserID(id string) error {
	if err := shared.ValidateID(id); err != nil {
		return ErrInvalidID
	}
	return nil
}

func New(name, email, plainPassword string, role Role) (*User, error) {
	u := &User{
		id:        uuid.NewV7().String(),
		isActive:  true,
		version:   1,
		createdAt: time.Now().UTC(),
	}

	if err := u.Rename(name); err != nil {
		return nil, err
	}
	if err := u.ChangeEmail(email); err != nil {
		return nil, err
	}
	if err := u.ChangePassword(plainPassword); err != nil {
		return nil, err
	}
	if err := u.ChangeRole(role); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	u.createdAt = now
	u.updatedAt = now

	return u, nil
}

func Restore(
	id, name, email, passwordHash string,
	role Role,
	isActive bool,
	version int,
	createdAt, updatedAt time.Time,
) *User {
	return &User{
		id:           id,
		name:         name,
		email:        email,
		passwordHash: passwordHash,
		role:         role,
		isActive:     isActive,
		version:      version,
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

func (u *User) ChangeEmail(email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if !emailPattern.MatchString(email) {
		return ErrInvalidEmail
	}
	u.email = email
	u.touch()
	return nil
}

func (u *User) ChangeRole(role Role) error {
	if !role.IsValid() {
		return ErrInvalidRole
	}
	u.role = role
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

func (u *User) SetVersion(version int) {
	u.version = version
}

func (u *User) ID() string           { return u.id }
func (u *User) PasswordHash() string { return u.passwordHash }
func (u *User) Name() string         { return u.name }
func (u *User) Email() string        { return u.email }
func (u *User) Role() Role           { return u.role }
func (u *User) IsActive() bool       { return u.isActive }
func (u *User) Version() int         { return u.version }
func (u *User) CreatedAt() time.Time { return u.createdAt }
func (u *User) UpdatedAt() time.Time { return u.updatedAt }
