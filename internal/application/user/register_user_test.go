package user

import (
	"context"
	"errors"
	"testing"

	domainuser "github.com/humberto0/shopkeeper/internal/domain/user"
	"github.com/humberto0/shopkeeper/internal/infrastructure/memory"
)

type stubRepository struct {
	existsByEmail func(ctx context.Context, email string) (bool, error)
	save          func(ctx context.Context, u *domainuser.User) error
}

func (s stubRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if s.existsByEmail == nil {
		return false, nil
	}
	return s.existsByEmail(ctx, email)
}

func (s stubRepository) Save(ctx context.Context, u *domainuser.User) error {
	if s.save == nil {
		return nil
	}
	return s.save(ctx, u)
}

func validInput() RegisterUserInput {
	return RegisterUserInput{
		Name:     "Humberto",
		Email:    "humberto@shop.com",
		Password: "test1234",
		Role:     domainuser.RoleOwner,
	}
}

func TestRegisterUser_Success(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewUserRepository()
	uc := NewRegisterUser(repo)

	out, err := uc.Execute(ctx, validInput())
	if err != nil {
		t.Fatalf("expected err to be nil, got %v", err)
	}

	if out.ID == "" {
		t.Error("expected a generated ID")
	}
	if out.Name != "Humberto" {
		t.Errorf("expected name %q, got %q", "Humberto", out.Name)
	}
	if out.Email != "humberto@shop.com" {
		t.Errorf("expected email %q, got %q", "humberto@shop.com", out.Email)
	}
	if out.Role != domainuser.RoleOwner {
		t.Errorf("expected role %q, got %q", domainuser.RoleOwner, out.Role)
	}
}

func TestRegisterUser_PersistsTheUser(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewUserRepository()
	uc := NewRegisterUser(repo)

	out, err := uc.Execute(ctx, validInput())
	if err != nil {
		t.Fatalf("expected err to be nil, got %v", err)
	}

	saved, err := repo.FindByEmail(ctx, validInput().Email)

	if err != nil {
		t.Fatalf("expected err to be nil, got %v", err)
	}
	if saved.ID() != out.ID {
		t.Errorf("expected ID %q, got %q", saved.ID(), out.ID)
	}
	if saved.Name() != out.Name {
		t.Errorf("expected name %q, got %q", saved.Name(), out.Name)
	}
	if !saved.IsActive() {
		t.Error("expected IsActive to be true")
	}
	if saved.PasswordHash() == "" {
		t.Error("expected tge stored user to carry a password hash")
	}
	if saved.PasswordHash() == validInput().Password {
		t.Error("password was stored in plain text")
	}
}

func TestRegisterUser_EmailAlreadyExists(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewUserRepository()
	uc := NewRegisterUser(repo)

	if _, err := uc.Execute(ctx, validInput()); err != nil {
		t.Fatalf("expected err to be nil, got %v", err)
	}

	_, err := uc.Execute(ctx, validInput())
	if !errors.Is(err, domainuser.ErrEmailAlreadyExists) {
		t.Errorf("expected err to be %v, got %v", domainuser.ErrEmailAlreadyExists, err)
	}

	if repo.Count() != 1 {
		t.Errorf("expected 1 stored user, got %d", repo.Count())
	}
}

func TestRegisterUser_InvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*RegisterUserInput)
		wantErr error
	}{
		{
			name:    "empty name",
			mutate:  func(in *RegisterUserInput) { in.Name = "" },
			wantErr: domainuser.ErrInvalidName,
		},
		{
			name:    "malformed email",
			mutate:  func(in *RegisterUserInput) { in.Email = "humberto.shop.com" },
			wantErr: domainuser.ErrInvalidEmail,
		},
		{
			name:    "password too short",
			mutate:  func(in *RegisterUserInput) { in.Password = "234" },
			wantErr: domainuser.ErrWeakPassword,
		},
		{
			name:    "unknown role",
			mutate:  func(in *RegisterUserInput) { in.Role = domainuser.Role("maneger") },
			wantErr: domainuser.ErrInvalidRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repo := memory.NewUserRepository()
			uc := NewRegisterUser(repo)

			in := validInput()
			tt.mutate(&in)
			out, err := uc.Execute(ctx, in)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected %v , got %v", tt.wantErr, err)
			}
			if out != nil {
				t.Error("expected a nil output on failure")
			}
			if repo.Count() != 0 {
				t.Errorf("expected 0 stored user, got %d", repo.Count())
			}

		})
	}
}

func TestRegisterUser_RepositoryFailure(t *testing.T) {
	reporErr := errors.New("database is down")

	t.Run("propagates a failure while checking the email", func(t *testing.T) {
		uc := NewRegisterUser(stubRepository{
			existsByEmail: func(ctx context.Context, email string) (bool, error) {
				return false, reporErr
			},
		})
		_, err := uc.Execute(context.Background(), validInput())
		if !errors.Is(err, reporErr) {
			t.Fatalf("expected err to be %v, got %v", reporErr, err)
		}
	})
}
