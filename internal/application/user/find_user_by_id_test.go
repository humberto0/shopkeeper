package user

import (
	"context"
	"errors"
	"testing"

	domainuser "github.com/humberto0/shopkeeper/internal/domain/user"
	"github.com/humberto0/shopkeeper/internal/infrastructure/memory"
)

func TestFindUserByID_Success(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewUserRepository()
	uc := NewFindUserByID(repo)

	u, err := domainuser.New("Humberto", "humberto@shop.com", "test1234", domainuser.RoleOwner)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if err := repo.Save(ctx, u); err != nil {
		t.Fatalf("error saving user: %v", err)
	}

	find, err := uc.Execute(ctx, u.ID())
	if err != nil {
		t.Fatalf("error finding user: %v", err)
	}
	if find.Name != u.Name() {
		t.Fatalf("expected user name %s, got %s", u.Name(), find.Name)
	}
	if find.Email != u.Email() {
		t.Fatalf("expected user email %s, got %s", u.Email(), find.Email)
	}
	if find.Role != u.Role() {
		t.Fatalf("expected user role %s, got %s", u.Role(), find.Role)
	}
	if find.Role != domainuser.RoleOwner {
		t.Fatalf("expected user role %s, got %s", domainuser.RoleOwner, find.Role)
	}
	if find.ID != u.ID() {
		t.Fatalf("expected user id %s, got %s", u.ID(), find.ID)
	}
	if find.IsActive != u.IsActive() {
		t.Fatalf("expected user isActive %v, got %v", u.IsActive(), find.IsActive)
	}
}

func TestFindUserByID_MalformedID(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewUserRepository()
	uc := NewFindUserByID(repo)

	find, err := uc.Execute(ctx, "not-a-uuid")
	if !errors.Is(err, domainuser.ErrInvalidID) {
		t.Fatalf("expected %v, got %v", domainuser.ErrInvalidID, err)
	}
	if find != nil {
		t.Fatalf("expected nil result, got %v", find)
	}
}

func TestFindUserByID_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewUserRepository()
	uc := NewFindUserByID(repo)

	find, err := uc.Execute(ctx, "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, domainuser.ErrNotFound) {
		t.Fatalf("expected %v, got %v", domainuser.ErrNotFound, err)
	}
	if find != nil {
		t.Fatalf("expected nil result, got %v", find)
	}
}
