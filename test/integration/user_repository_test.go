package integration

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"uuid"

	identity "github.com/humberto0/shopkeeper/internal/identity/domain"
	identitypg "github.com/humberto0/shopkeeper/internal/identity/postgres"
)

func newTestUser(t *testing.T, email string) *identity.User {
	t.Helper()

	u, err := identity.New("Humberto", email, "test1234", identity.RoleOwner)
	if err != nil {
		t.Fatalf("failed to build user: %v", err)
	}
	return u
}

func TestUserRepository_SaveAndFindByID(t *testing.T) {
	truncateUsers(t)

	ctx := context.Background()
	repo := identitypg.NewUserRepository(testPool)

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

	if found.Version() != 1 {
		t.Errorf("expected version 1, got %d", found.Version())
	}
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	truncateUsers(t)

	ctx := context.Background()
	repo := identitypg.NewUserRepository(testPool)

	_, err := repo.FindByID(ctx, "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, identity.ErrNotFound) {
		t.Fatalf("expected %v, got %v", identity.ErrNotFound, err)
	}
}

func TestUserRepository_FindByEmail(t *testing.T) {
	truncateUsers(t)

	ctx := context.Background()
	repo := identitypg.NewUserRepository(testPool)

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

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	truncateUsers(t)

	ctx := context.Background()
	repo := identitypg.NewUserRepository(testPool)

	_, err := repo.FindByEmail(ctx, "nobody@shop.com")
	if !errors.Is(err, identity.ErrNotFound) {
		t.Fatalf("expected %v, got %v", identity.ErrNotFound, err)
	}
}

func TestUserRepository_DuplicateEmail(t *testing.T) {
	truncateUsers(t)

	ctx := context.Background()
	repo := identitypg.NewUserRepository(testPool)

	first := newTestUser(t, "humberto@shop.com")
	if err := repo.Save(ctx, first); err != nil {
		t.Fatalf("expected no error saving user: %v", err)
	}

	second := newTestUser(t, "humberto@shop.com")
	err := repo.Save(ctx, second)

	if !errors.Is(err, identity.ErrEmailAlreadyExists) {
		t.Fatalf("expected email already exists, got %v", err)
	}
}

func TestUserRepository_DuplicateEmailIgnoreCase(t *testing.T) {
	truncateUsers(t)

	ctx := context.Background()
	repo := identitypg.NewUserRepository(testPool)

	first := newTestUser(t, "humberto@shop.com")
	if err := repo.Save(ctx, first); err != nil {
		t.Fatalf("expected no error saving user: %v", err)
	}

	second := newTestUser(t, "HUMBERTO@SHOP.COM")
	err := repo.Save(ctx, second)

	if !errors.Is(err, identity.ErrEmailAlreadyExists) {
		t.Fatalf("expected email already exists, got %v", err)
	}
}

func TestUserRepository_Update(t *testing.T) {
	truncateUsers(t)

	ctx := context.Background()
	repo := identitypg.NewUserRepository(testPool)

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

	if created.Version() != 2 {
		t.Errorf("expected in-memory version to be updated to 2, got %d", created.Version())
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

	if found.Version() != 2 {
		t.Errorf("expected version 2, got %d", found.Version())
	}
}

func TestUserRepository_UpdateAllFields(t *testing.T) {
	truncateUsers(t)

	ctx := context.Background()
	repo := identitypg.NewUserRepository(testPool)

	created := newTestUser(t, "humberto@shop.com")
	if err := repo.Save(ctx, created); err != nil {
		t.Fatalf("expected no error saving user: %v", err)
	}

	if err := created.ChangeEmail("humberto.jr@shop.com"); err != nil {
		t.Fatalf("expected no error changing email: %v", err)
	}
	if err := created.ChangePassword("newpassword1"); err != nil {
		t.Fatalf("expected no error changing password: %v", err)
	}
	if err := created.ChangeRole(identity.RoleClerk); err != nil {
		t.Fatalf("expected no error changing role: %v", err)
	}

	if err := repo.Update(ctx, created); err != nil {
		t.Fatalf("expected no error updating user: %v", err)
	}

	found, err := repo.FindByID(ctx, created.ID())
	if err != nil {
		t.Fatalf("expected no error finding user: %v", err)
	}

	if found.Email() != "humberto.jr@shop.com" {
		t.Errorf("expected email %q, got %q", "humberto.jr@shop.com", found.Email())
	}
	if found.PasswordHash() != created.PasswordHash() {
		t.Errorf("expected password hash %q, got %q", created.PasswordHash(), found.PasswordHash())
	}
	if err := found.Authenticate("newpassword1"); err != nil {
		t.Errorf("expected authentication with the new password to succeed, got %v", err)
	}
	if found.Role() != identity.RoleClerk {
		t.Errorf("expected role %q, got %q", identity.RoleClerk, found.Role())
	}
}

func TestUserRepository_UpdateDuplicateEmail(t *testing.T) {
	truncateUsers(t)

	ctx := context.Background()
	repo := identitypg.NewUserRepository(testPool)

	taken := newTestUser(t, "taken@shop.com")
	if err := repo.Save(ctx, taken); err != nil {
		t.Fatalf("expected no error saving user: %v", err)
	}

	other := newTestUser(t, "free@shop.com")
	if err := repo.Save(ctx, other); err != nil {
		t.Fatalf("expected no error saving user: %v", err)
	}

	if err := other.ChangeEmail("taken@shop.com"); err != nil {
		t.Fatalf("expected no error changing email: %v", err)
	}

	err := repo.Update(ctx, other)
	if !errors.Is(err, identity.ErrEmailAlreadyExists) {
		t.Fatalf("expected %v, got %v", identity.ErrEmailAlreadyExists, err)
	}
}

func TestUserRepository_UpdateVersionConflict(t *testing.T) {
	truncateUsers(t)

	ctx := context.Background()
	repo := identitypg.NewUserRepository(testPool)

	created := newTestUser(t, "humberto@shop.com")
	if err := repo.Save(ctx, created); err != nil {
		t.Fatalf("expected no error saving user: %v", err)
	}

	stale, err := repo.FindByID(ctx, created.ID())
	if err != nil {
		t.Fatalf("expected no error finding user: %v", err)
	}

	if err := created.Rename("Humberto Junior"); err != nil {
		t.Fatalf("expected no error renaming user: %v", err)
	}
	if err := repo.Update(ctx, created); err != nil {
		t.Fatalf("expected no error updating user: %v", err)
	}

	if err := stale.Rename("Someone Else"); err != nil {
		t.Fatalf("expected no error renaming user: %v", err)
	}
	err = repo.Update(ctx, stale)
	if !errors.Is(err, identity.ErrConflict) {
		t.Fatalf("expected %v, got %v", identity.ErrConflict, err)
	}
}

func TestUserRepository_UpdateNotFound(t *testing.T) {
	truncateUsers(t)
	ctx := context.Background()
	repo := identitypg.NewUserRepository(testPool)

	orphan := newTestUser(t, "orphan@shop.com")

	err := repo.Update(ctx, orphan)
	if !errors.Is(err, identity.ErrNotFound) {
		t.Fatalf("expected %v, got %v", identity.ErrNotFound, err)
	}
}

func TestUserRepository_ExistsByEmail(t *testing.T) {
	truncateUsers(t)
	ctx := context.Background()
	repo := identitypg.NewUserRepository(testPool)

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

// saveUsersCreatedAt stores n users created one minute apart (index 0 is the
// oldest) and returns their ids in creation order.
func saveUsersCreatedAt(t *testing.T, repo *identitypg.UserRepository, n int) []string {
	t.Helper()

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ids := make([]string, 0, n)
	for i := range n {
		createdAt := base.Add(time.Duration(i) * time.Minute)
		u := identity.Restore(
			uuid.NewV7().String(),
			fmt.Sprintf("User %d", i),
			fmt.Sprintf("user%d@shop.com", i),
			"hash",
			identity.RoleClerk,
			true,
			1,
			createdAt,
			createdAt,
		)
		if err := repo.Save(context.Background(), u); err != nil {
			t.Fatalf("failed to save user %d: %v", i, err)
		}
		ids = append(ids, u.ID())
	}
	return ids
}

func TestUserRepository_List(t *testing.T) {
	truncateUsers(t)

	ctx := context.Background()
	repo := identitypg.NewUserRepository(testPool)
	ids := saveUsersCreatedAt(t, repo, 5)

	tests := []struct {
		name          string
		limit, offset int
		wantIDs       []string
	}{
		{"first page newest first", 2, 0, []string{ids[4], ids[3]}},
		{"second page", 2, 2, []string{ids[2], ids[1]}},
		{"partial last page", 2, 4, []string{ids[0]}},
		{"offset past the end", 2, 10, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, total, err := repo.List(ctx, tt.limit, tt.offset)
			if err != nil {
				t.Fatalf("failed to list users: %v", err)
			}
			if total != 5 {
				t.Errorf("expected total 5, got %d", total)
			}
			if got == nil {
				t.Fatal("expected a non-nil slice")
			}
			if len(got) != len(tt.wantIDs) {
				t.Fatalf("expected %d users, got %d", len(tt.wantIDs), len(got))
			}
			for i, u := range got {
				if u.ID() != tt.wantIDs[i] {
					t.Errorf("position %d: expected id %s, got %s", i, tt.wantIDs[i], u.ID())
				}
			}
		})
	}
}

func TestUserRepository_List_RestoresAllFields(t *testing.T) {
	truncateUsers(t)

	ctx := context.Background()
	repo := identitypg.NewUserRepository(testPool)

	created := newTestUser(t, "humberto@shop.com")
	if err := repo.Save(ctx, created); err != nil {
		t.Fatalf("failed to save user: %v", err)
	}

	got, total, err := repo.List(ctx, 10, 0)
	if err != nil {
		t.Fatalf("failed to list users: %v", err)
	}
	if total != 1 || len(got) != 1 {
		t.Fatalf("expected exactly one user, got total %d len %d", total, len(got))
	}

	u := got[0]
	if u.ID() != created.ID() || u.Name() != created.Name() || u.Email() != created.Email() ||
		u.Role() != created.Role() || u.IsActive() != created.IsActive() || u.Version() != 1 {
		t.Errorf("listed user does not match saved user: got %+v", u)
	}
	if !u.CreatedAt().Equal(created.CreatedAt().Truncate(time.Microsecond)) {
		t.Errorf("expected createdAt %v, got %v", created.CreatedAt(), u.CreatedAt())
	}
}

func TestUserRepository_List_Empty(t *testing.T) {
	truncateUsers(t)

	got, total, err := identitypg.NewUserRepository(testPool).List(context.Background(), 20, 0)
	if err != nil {
		t.Fatalf("failed to list users: %v", err)
	}
	if total != 0 || len(got) != 0 {
		t.Errorf("expected no users, got total %d len %d", total, len(got))
	}
}
