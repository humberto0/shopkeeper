package user

import (
	"context"
	"errors"
	"testing"

	domainuser "github.com/humberto0/shopkeeper/internal/domain/user"
	"github.com/humberto0/shopkeeper/internal/infrastructure/memory"
)

func validEditUserInput(id string, email string, version int) EditUserInput {
	return EditUserInput{
		ID:      id,
		Name:    "humbertinho",
		Email:   email,
		Role:    domainuser.RoleClerk,
		Version: version,
	}
}

func TestEditUser_Success(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewUserRepository()
	ed := NewEditUser(repo)

	created, err := domainuser.New("humberto", "humberto@shop.com", "test1234", domainuser.RoleOwner)
	if err != nil {
		t.Fatalf("creating user: %v", err)
	}
	if err := repo.Save(ctx, created); err != nil {
		t.Fatalf("saving user: %v", err)
	}

	originalVersion := created.Version()
	in := validEditUserInput(created.ID(), created.Email(), originalVersion)
	edited, err := ed.Execute(ctx, in)
	if err != nil {
		t.Fatalf("editing user: %v", err)
	}

	if edited.Name != in.Name {
		t.Errorf("expected name %s, got %s", in.Name, edited.Name)
	}
	if edited.Email != in.Email {
		t.Errorf("expected email %s, got %s", in.Email, edited.Email)
	}
	if edited.Role != in.Role {
		t.Errorf("expected role %s, got %s", in.Role, edited.Role)
	}
	if edited.Version != originalVersion {
		t.Errorf("expected version %d, got %d", originalVersion+1, edited.Version)
	}
}

func TestEditUser_StaleVersionConflict(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewUserRepository()
	ed := NewEditUser(repo)

	created, err := domainuser.New("humberto", "humberto@shop.com", "test1234", domainuser.RoleOwner)
	if err != nil {
		t.Fatalf("creating user: %v", err)
	}
	if err := repo.Save(ctx, created); err != nil {
		t.Fatalf("saving user: %v", err)
	}

	staleVersion := created.Version() - 1
	edition, err := ed.Execute(ctx, validEditUserInput(created.ID(), created.Email(), staleVersion))
	if !errors.Is(err, domainuser.ErrConflict) {
		t.Errorf("expected error %v, got %v", domainuser.ErrConflict, err)
	}
	assertNilOutput(t, edition)
}

func TestEditUser_MalformedId(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewUserRepository()
	ed := NewEditUser(repo)

	edition, err := ed.Execute(ctx, validEditUserInput("test", "humberto@shop.com", 0))
	if !errors.Is(err, domainuser.ErrInvalidID) {
		t.Errorf("expected error %v, got %v", domainuser.ErrInvalidID, err)
	}
	assertNilOutput(t, edition)
}

func TestEditUser_UserNotFound(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewUserRepository()
	ed := NewEditUser(repo)

	edition, err := ed.Execute(ctx, validEditUserInput("3d0ca315-aff9-4fc2-be61-3b76b9a2d798", "humberto@shop.com", 0))
	if !errors.Is(err, domainuser.ErrNotFound) {
		t.Errorf("expected error %v, got %v", domainuser.ErrNotFound, err)
	}
	assertNilOutput(t, edition)
}

func TestEditUser_MalformedEmail(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewUserRepository()
	ed := NewEditUser(repo)

	created, err := domainuser.New("humberto", "humberto@shop.com", "test1234", domainuser.RoleOwner)
	if err != nil {
		t.Fatalf("creating user: %v", err)
	}
	if err := repo.Save(ctx, created); err != nil {
		t.Fatalf("saving user: %v", err)
	}

	edition, err := ed.Execute(ctx, validEditUserInput(created.ID(), "humberto", created.Version()))
	if !errors.Is(err, domainuser.ErrInvalidEmail) {
		t.Errorf("expected error %v, got %v", domainuser.ErrInvalidEmail, err)
	}
	assertNilOutput(t, edition)
}

func assertNilOutput(t *testing.T, out *EditUserOutput) {
	t.Helper()
	if out != nil {
		t.Errorf("expected a nil output on failure, got %v", out)
	}
}
