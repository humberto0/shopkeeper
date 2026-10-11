package user

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"uuid"

	domainuser "github.com/humberto0/shopkeeper/internal/domain/user"
	"github.com/humberto0/shopkeeper/internal/infrastructure/memory"
)

// seedUsers stores n users created one minute apart (user 0 is the oldest)
// and returns their ids in creation order.
func seedUsers(t *testing.T, repo *memory.UserRepository, n int) []string {
	t.Helper()

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ids := make([]string, 0, n)
	for i := range n {
		createdAt := base.Add(time.Duration(i) * time.Minute)
		u := domainuser.Restore(
			uuid.NewV7().String(),
			fmt.Sprintf("User %d", i),
			fmt.Sprintf("user%d@shop.com", i),
			"hash",
			domainuser.RoleClerk,
			true,
			1,
			createdAt,
			createdAt,
		)
		if err := repo.Save(context.Background(), u); err != nil {
			t.Fatalf("saving user %d: %v", i, err)
		}
		ids = append(ids, u.ID())
	}
	return ids
}

func TestListUsers_NewestFirstWithTotals(t *testing.T) {
	repo := memory.NewUserRepository()
	ids := seedUsers(t, repo, 5)
	uc := NewListUsers(repo)

	out, err := uc.Execute(context.Background(), ListUsersInput{Page: 1, Limit: 2})
	if err != nil {
		t.Fatalf("listing users: %v", err)
	}

	if out.Page != 1 || out.Limit != 2 {
		t.Errorf("expected page 1 limit 2, got page %d limit %d", out.Page, out.Limit)
	}
	if out.TotalItems != 5 {
		t.Errorf("expected 5 total items, got %d", out.TotalItems)
	}
	if out.TotalPages != 3 {
		t.Errorf("expected 3 total pages, got %d", out.TotalPages)
	}
	if len(out.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(out.Items))
	}
	if out.Items[0].ID != ids[4] || out.Items[1].ID != ids[3] {
		t.Errorf("expected newest users first (%s, %s), got (%s, %s)",
			ids[4], ids[3], out.Items[0].ID, out.Items[1].ID)
	}
	if out.Items[0].Email != "user4@shop.com" || out.Items[0].Role != domainuser.RoleClerk || !out.Items[0].IsActive {
		t.Errorf("unexpected item fields: %+v", out.Items[0])
	}
}

func TestListUsers_LastPartialPage(t *testing.T) {
	repo := memory.NewUserRepository()
	ids := seedUsers(t, repo, 5)
	uc := NewListUsers(repo)

	out, err := uc.Execute(context.Background(), ListUsersInput{Page: 3, Limit: 2})
	if err != nil {
		t.Fatalf("listing users: %v", err)
	}
	if len(out.Items) != 1 || out.Items[0].ID != ids[0] {
		t.Fatalf("expected only the oldest user %s, got %+v", ids[0], out.Items)
	}
	if out.HasNext() {
		t.Error("expected no next page")
	}
}

func TestListUsers_PagePastTheEndIsEmpty(t *testing.T) {
	repo := memory.NewUserRepository()
	seedUsers(t, repo, 3)
	uc := NewListUsers(repo)

	out, err := uc.Execute(context.Background(), ListUsersInput{Page: 10, Limit: 2})
	if err != nil {
		t.Fatalf("listing users: %v", err)
	}
	if out.Items == nil || len(out.Items) != 0 {
		t.Errorf("expected an empty, non-nil items slice, got %#v", out.Items)
	}
	if out.TotalItems != 3 || out.Page != 10 {
		t.Errorf("expected total 3 on page 10, got total %d on page %d", out.TotalItems, out.Page)
	}
}

func TestListUsers_EmptyRepository(t *testing.T) {
	uc := NewListUsers(memory.NewUserRepository())

	out, err := uc.Execute(context.Background(), ListUsersInput{})
	if err != nil {
		t.Fatalf("listing users: %v", err)
	}
	if len(out.Items) != 0 || out.TotalItems != 0 || out.TotalPages != 0 {
		t.Errorf("expected an empty page, got %+v", out)
	}
}

func TestListUsers_NormalizesInput(t *testing.T) {
	tests := []struct {
		name      string
		in        ListUsersInput
		wantPage  int
		wantLimit int
	}{
		{"defaults", ListUsersInput{}, 1, 20},
		{"negative values", ListUsersInput{Page: -1, Limit: -5}, 1, 20},
		{"limit capped", ListUsersInput{Page: 1, Limit: 1000}, 1, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &listUsersRepositoryStub{}
			out, err := NewListUsers(repo).Execute(context.Background(), tt.in)
			if err != nil {
				t.Fatalf("listing users: %v", err)
			}
			if out.Page != tt.wantPage || out.Limit != tt.wantLimit {
				t.Errorf("got page %d limit %d, want page %d limit %d", out.Page, out.Limit, tt.wantPage, tt.wantLimit)
			}
			wantOffset := (tt.wantPage - 1) * tt.wantLimit
			if repo.gotLimit != tt.wantLimit || repo.gotOffset != wantOffset {
				t.Errorf("repository called with limit %d offset %d, want limit %d offset %d",
					repo.gotLimit, repo.gotOffset, tt.wantLimit, wantOffset)
			}
		})
	}
}

func TestListUsers_RepositoryFailure(t *testing.T) {
	errDB := errors.New("connection refused")
	uc := NewListUsers(&listUsersRepositoryStub{err: errDB})

	out, err := uc.Execute(context.Background(), ListUsersInput{})
	if !errors.Is(err, errDB) {
		t.Fatalf("expected wrapped %v, got %v", errDB, err)
	}
	if out != nil {
		t.Errorf("expected nil output, got %+v", out)
	}
}

type listUsersRepositoryStub struct {
	gotLimit, gotOffset int
	err                 error
}

func (s *listUsersRepositoryStub) List(_ context.Context, limit, offset int) ([]*domainuser.User, int, error) {
	s.gotLimit, s.gotOffset = limit, offset
	if s.err != nil {
		return nil, 0, s.err
	}
	return []*domainuser.User{}, 0, nil
}

func TestListUsers_ItemsAreDetachedFromRepository(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewUserRepository()
	ids := seedUsers(t, repo, 1)

	users, _, err := repo.List(ctx, 10, 0)
	if err != nil {
		t.Fatalf("listing users: %v", err)
	}
	if err := users[0].Rename("Changed outside the repository"); err != nil {
		t.Fatalf("renaming user: %v", err)
	}

	stored, err := repo.FindByID(ctx, ids[0])
	if err != nil {
		t.Fatalf("finding user: %v", err)
	}
	if stored.Name() != "User 0" {
		t.Errorf("expected stored name %q to be unchanged, got %q", "User 0", stored.Name())
	}
}
