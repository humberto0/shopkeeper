package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/humberto0/shopkeeper/internal/domain/user"
)

var _ user.Repository = (*UserRepository)(nil)

const uniqueViolationCode = "23505"

const emailUniqueIndex = "users_email_unique_idx"

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

const insertUserQuery = `
	INSERT INTO users (id, name, email, password_hash, role, is_active, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

func (r *UserRepository) Save(ctx context.Context, u *user.User) error {
	_, err := r.pool.Exec(ctx, insertUserQuery,
		u.ID(),
		u.Name(),
		u.Email(),
		u.PasswordHash(),
		string(u.Role()),
		u.IsActive(),
		u.CreatedAt(),
		u.UpdatedAt(),
	)
	if err != nil {
		if isUniqueViolation(err, emailUniqueIndex) {
			return user.ErrEmailAlreadyExists
		}
		return fmt.Errorf("inserting user: %w", err)
	}
	return nil
}

const updateUserQuery = `
	UPDATE users
	SET name = $2, email = $3, password_hash = $4, role = $5, is_active = $6, updated_at = $7
	WHERE id = $1`

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	tag, err := r.pool.Exec(ctx, updateUserQuery,
		u.ID(),
		u.Name(),
		u.Email(),
		u.PasswordHash(),
		string(u.Role()),
		u.IsActive(),
		u.UpdatedAt(),
	)
	if err != nil {
		if isUniqueViolation(err, emailUniqueIndex) {
			return user.ErrEmailAlreadyExists
		}
		return fmt.Errorf("updating user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return user.ErrNotFound
	}
	return nil
}

const findByIDQuery = `
	SELECT id, name, email, password_hash, role, is_active, created_at, updated_at
	FROM users WHERE id = $1`

func (r *UserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	row := r.pool.QueryRow(ctx, findByIDQuery, id)
	u, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("finding user by id: %w", err)
	}
	return u, nil
}

const findByEmailQuery = `
	SELECT id, name, email, password_hash, role, is_active, created_at, updated_at
	FROM users WHERE lower(email) = lower($1)`

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	row := r.pool.QueryRow(ctx, findByEmailQuery, email)
	u, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("finding user by email: %w", err)
	}
	return u, nil
}

const existsByEmailQuery = `
	SELECT EXISTS (SELECT 1 FROM users WHERE lower(email) = lower($1))`

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	if err := r.pool.QueryRow(ctx, existsByEmailQuery, email).Scan(&exists); err != nil {
		return false, fmt.Errorf("checking if email exists: %w", err)
	}
	return exists, nil
}

func scanUser(row pgx.Row) (*user.User, error) {
	var (
		id, name, email, passwordHash, role string
		isActive                            bool
		createdAt, updatedAt                time.Time
	)
	if err := row.Scan(&id, &name, &email, &passwordHash, &role, &isActive, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	return user.Restore(id, name, email, passwordHash, user.Role(role), isActive, createdAt, updatedAt), nil
}

func isUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == uniqueViolationCode && pgErr.ConstraintName == constraint
}
