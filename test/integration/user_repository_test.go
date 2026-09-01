package integration

import (
	"context"
	"errors"
	"testing"

	"github.com/humberto0/shopkeeper/internal/domain/user"
	"github.com/humberto0/shopkeeper/internal/infrastructure/postgres"
)

func newTestUser(t *testing.T, email string) *user.User {
	t.Helper()

	u, err := user.New("Humberto", email, "test1234", user.RoleOwner)
	if err != nil {
		t.Fatalf("failed to build user: %v", err)
	}
	return u
}

func TestUserRepository_SaveAndFindByID(t *testing.T) {
	truncateUsers(t)

	ctx := context.Background()
	repo := postgres.NewUserRepository(testPool)

	created := newTestUser(t, "humberto@shop.com")

	if err := repo.Save(ctx, created); err != nil {
		t.Fatalf("failed to save user: %v", err)
	}

	found, err := repo.FindByID(ctx, created.ID())
	if err != nil {
		t.Fatalf("failed to find user: %v", err)
	}
	if found.ID() != created.ID() {
		t.Errorf("expected id %q, got %q", created.ID(), found.ID())
	}

	if found.Name() != created.Name() {
		t.Errorf("expected name %q, got %q", created.Name(), found.Name())
	}

	if found.Email() != created.Email() {
		t.Errorf("expected email %q, got %q", created.Email(), found.Email())
	}

	if found.Role() != created.Role() {
		t.Errorf("expected role %q, got %q", created.Role(), found.Role())
	}

	if found.IsActive() != created.IsActive() {
		t.Errorf("expected IsActive %v, got %v", created.IsActive(), found.IsActive())
	}

	if found.PasswordHash() != created.PasswordHash() {
		t.Errorf("expected PasswordHash %v, got %v", created.PasswordHash(), found.PasswordHash())
	}

	if !found.CreatedAt().Truncate(1000).Equal(created.CreatedAt().Truncate(1000)) {
		t.Errorf("expected createdAt %v, got %v", created.CreatedAt(), found.CreatedAt())
	}
}

func TestUserRepository_FindByEmail(t *testing.T) {
	truncateUsers(t)

	ctx := context.Background()
	repo := postgres.NewUserRepository(testPool)

	created := newTestUser(t, "humberto@shop.com")
	if err := repo.Save(ctx, created); err != nil {
		t.Fatalf("expected no error saving user: %v", err)
	}
	found, err := repo.FindByEmail(ctx, "HUMBERTO@SHOP.COM")
	if err != nil {
		t.Fatalf("expected no error finding user by email, got %v", err)
	}
	if found.ID() != created.ID() {
		t.Errorf("expected id %q, got %q", created.ID(), found.ID())
	}

}

func TestUserRepository_DuplicateEmail(t *testing.T) {
	truncateUsers(t)

	ctx := context.Background()
	repo := postgres.NewUserRepository(testPool)

	first := newTestUser(t, "humberto@shop.com")
	if err := repo.Save(ctx, first); err != nil {
		t.Fatalf("expected no error saving user: %v", err)
	}

	second := newTestUser(t, "humberto@shop.com")
	err := repo.Save(ctx, second)

	if !errors.Is(err, user.ErrEmailAlreadyExists) {
		t.Fatalf("expected email already exists, got %v", err)
	}
}

func TestUserRepository_DuplicateEmailIgnoreCase(t *testing.T) {
	truncateUsers(t)

	ctx := context.Background()
	repo := postgres.NewUserRepository(testPool)

	first := newTestUser(t, "humberto@shop.com")
	if err := repo.Save(ctx, first); err != nil {
		t.Fatalf("expected no error saving user: %v", err)
	}

	second := newTestUser(t, "HUMBERTO@SHOP.COM")
	err := repo.Save(ctx, second)

	if !errors.Is(err, user.ErrEmailAlreadyExists) {
		t.Fatalf("expected email already exists, got %v", err)
	}
}

func TestUserRepository_Update(t *testing.T) {
	truncateUsers(t)

	ctx := context.Background()
	repo := postgres.NewUserRepository(testPool)

	created := newTestUser(t, "humberto@shop.com")
	if err := repo.Save(ctx, created); err != nil {
		t.Fatalf("expected no error saving user: %v", err)
	}

	if err := created.Rename("Humberto Junior"); err != nil {
		t.Fatalf("expected no error renaming user: %v", err)
	}

	created.Deactivate()

	if err := repo.Update(ctx, created); err != nil {
		t.Fatalf("expected no error saving user: %v", err)
	}

	found, err := repo.FindByID(ctx, created.ID())
	if err != nil {
		t.Fatalf("expected no error finding user: %v", err)
	}

	if found.Name() != "Humberto Junior" {
		t.Errorf("expected name %q, got %q", "Humberto Junior", found.Name())
	}

	if found.IsActive() {
		t.Errorf("expected IsActive %v, got %v", false, found.IsActive())
	}
}

func TestUserRepository_UpdateNotFound(t *testing.T) {
	truncateUsers(t)
	ctx := context.Background()
	repo := postgres.NewUserRepository(testPool)

	orphan := newTestUser(t, "orphan@shop.com")

	err := repo.Update(ctx, orphan)
	if !errors.Is(err, user.ErrNotFound) {
		t.Fatalf("expected %v, got %v", user.ErrNotFound, err)
	}
}

func TestUserRepository_ExistsByEmail(t *testing.T) {
	truncateUsers(t)
	ctx := context.Background()
	repo := postgres.NewUserRepository(testPool)

	if err := repo.Save(ctx, newTestUser(t, "humberto@shop.com")); err != nil {
		t.Fatalf("expected no error saving user: %v", err)
	}

	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"existing email", "humberto@shop.com", true},
		{"existing email in upper case", "HUMBERTO@SHOP.COM", true},
		{"unknow email", "nobody@shop.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.ExistsByEmail(ctx, tt.email)
			if err != nil {
				t.Fatalf("expected no error saving user: %v", err)
			}
			if got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
