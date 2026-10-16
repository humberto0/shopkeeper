package identityapp

import (
	"context"
	"errors"
	"testing"

	identity "github.com/humberto0/shopkeeper/internal/identity/domain"
	identitymem "github.com/humberto0/shopkeeper/internal/identity/memory"
)

func validEditUserInput(id string, email string, version int) EditUserInput {
	return EditUserInput{
		ID:      id,
		Name:    "humbertinho",
		Email:   email,
		Role:    identity.RoleClerk,
		Version: version,
	}
}

func TestEditUser_Success(t *testing.T) {
	ctx := context.Background()
	repo := identitymem.NewUserRepository()
	ed := NewEditUser(repo)

	created, err := identity.New("humberto", "humberto@shop.com", "test1234", identity.RoleOwner)
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
	if edited.Version != originalVersion+1 {
		t.Errorf("expected version %d, got %d", originalVersion+1, edited.Version)
	}
}

func TestEditUser_ConsecutiveEditsUseReturnedVersion(t *testing.T) {
	ctx := context.Background()
	repo := identitymem.NewUserRepository()
	ed := NewEditUser(repo)

	created, err := identity.New("humberto", "humberto@shop.com", "test1234", identity.RoleOwner)
	if err != nil {
		t.Fatalf("creating user: %v", err)
	}
	if err := repo.Save(ctx, created); err != nil {
		t.Fatalf("saving user: %v", err)
	}

	originalVersion := created.Version()
	first, err := ed.Execute(ctx, validEditUserInput(created.ID(), created.Email(), originalVersion))
	if err != nil {
		t.Fatalf("first edit: %v", err)
	}

	second := validEditUserInput(created.ID(), "humbertinho@shop.com", first.Version)
	edited, err := ed.Execute(ctx, second)
	if err != nil {
		t.Fatalf("second edit with version %d: %v", first.Version, err)
	}
	if edited.Version != originalVersion+2 {
		t.Errorf("expected version %d, got %d", originalVersion+2, edited.Version)
	}

	stale, err := ed.Execute(ctx, validEditUserInput(created.ID(), second.Email, originalVersion))
	if !errors.Is(err, identity.ErrConflict) {
		t.Errorf("expected error %v reusing version %d, got %v", identity.ErrConflict, originalVersion, err)
	}
	assertNilOutput(t, stale)

	stored, err := repo.FindByEmail(ctx, second.Email)
	if err != nil {
		t.Fatalf("finding user by new email: %v", err)
	}
	if stored.Version() != originalVersion+2 {
		t.Errorf("expected stored version %d, got %d", originalVersion+2, stored.Version())
	}
	if exists, _ := repo.ExistsByEmail(ctx, created.Email()); exists {
		t.Errorf("expected old email %s to be released", created.Email())
	}
}

func TestEditUser_StaleVersionConflict(t *testing.T) {
	ctx := context.Background()
	repo := identitymem.NewUserRepository()
	ed := NewEditUser(repo)

	created, err := identity.New("humberto", "humberto@shop.com", "test1234", identity.RoleOwner)
	if err != nil {
		t.Fatalf("creating user: %v", err)
	}
	if err := repo.Save(ctx, created); err != nil {
		t.Fatalf("saving user: %v", err)
	}

	staleVersion := created.Version() - 1
	edition, err := ed.Execute(ctx, validEditUserInput(created.ID(), created.Email(), staleVersion))
	if !errors.Is(err, identity.ErrConflict) {
		t.Errorf("expected error %v, got %v", identity.ErrConflict, err)
	}
	assertNilOutput(t, edition)
}

func TestEditUser_MalformedId(t *testing.T) {
	ctx := context.Background()
	repo := identitymem.NewUserRepository()
	ed := NewEditUser(repo)

	edition, err := ed.Execute(ctx, validEditUserInput("test", "humberto@shop.com", 0))
	if !errors.Is(err, identity.ErrInvalidID) {
		t.Errorf("expected error %v, got %v", identity.ErrInvalidID, err)
	}
	assertNilOutput(t, edition)
}

func TestEditUser_UserNotFound(t *testing.T) {
	ctx := context.Background()
	repo := identitymem.NewUserRepository()
	ed := NewEditUser(repo)

	edition, err := ed.Execute(ctx, validEditUserInput("3d0ca315-aff9-4fc2-be61-3b76b9a2d798", "humberto@shop.com", 0))
	if !errors.Is(err, identity.ErrNotFound) {
		t.Errorf("expected error %v, got %v", identity.ErrNotFound, err)
	}
	assertNilOutput(t, edition)
}

func TestEditUser_MalformedEmail(t *testing.T) {
	ctx := context.Background()
	repo := identitymem.NewUserRepository()
	ed := NewEditUser(repo)

	created, err := identity.New("humberto", "humberto@shop.com", "test1234", identity.RoleOwner)
	if err != nil {
		t.Fatalf("creating user: %v", err)
	}
	if err := repo.Save(ctx, created); err != nil {
		t.Fatalf("saving user: %v", err)
	}

	edition, err := ed.Execute(ctx, validEditUserInput(created.ID(), "humberto", created.Version()))
	if !errors.Is(err, identity.ErrInvalidEmail) {
		t.Errorf("expected error %v, got %v", identity.ErrInvalidEmail, err)
	}
	assertNilOutput(t, edition)
}

func assertNilOutput(t *testing.T, out *EditUserOutput) {
	t.Helper()
	if out != nil {
		t.Errorf("expected a nil output on failure, got %v", out)
	}
}
