package user

import (
	"errors"
	"testing"
)

func TestNewUser(t *testing.T) {
	t.Run("Creates a valid user", func(t *testing.T) {
		u, err := New("Humberto", "humberto@shop.com", "test1234", RoleOwner)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if u.Name() != "Humberto" {
			t.Errorf("expected name %q, got %q", "Humberto", u.Name())
		}
		if !u.IsActive() {
			t.Errorf("expected IsActive to be true")
		}
		if u.ID() == "" {
			t.Errorf("expected a generated ID")
		}
	})

	t.Run("never stores the password in plain text", func(t *testing.T) {
		const plain = "test1234"
		u, err := New("Humberto", "humberto@shop.com", plain, RoleClerk)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if u.PasswordHash() == plain {
			t.Fatalf("password was stored in plain text")
		}
		if u.PasswordHash() == "" {
			t.Fatalf("expected a password hash")
		}

	})

	t.Run("rejects invalid input", func(t *testing.T) {
		tests := []struct {
			name     string
			userName string
			email    string
			password string
			role     Role
			wantErr  error
		}{
			{"empty name", "", "humberto@shop.com", "test1234", RoleOwner, ErrInvalidName},
			{"email without at sign", "Humberto", "humbertoshop.com", "test1234", RoleOwner, ErrInvalidEmail},
			{"empty email", "Humberto", "", "test1234", RoleOwner, ErrInvalidEmail},
			{"password too short", "Humberto", "humberto@shop.com", "123", RoleOwner, ErrWeakPassword},
			{"unknown role", "Humberto", "humberto@shop.com", "test1234", Role("manager"), ErrInvalidRole},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := New(tt.userName, tt.email, tt.password, tt.role)
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected %v, got %v", tt.wantErr, err)
				}
			})
		}
	})
}

func TestAuthenticate(t *testing.T) {
	t.Run("Sucessed with the correct password", func(t *testing.T) {
		u := mustNewUser(t)

		if err := u.Authenticate("test1234"); err != nil {
			t.Errorf("expected no error,, got %v", err)
		}
	})
	t.Run("fails with a wrong password", func(t *testing.T) {
		u := mustNewUser(t)
		if err := u.Authenticate("wrongpassword"); !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("expected %v, got %v", ErrInvalidCredentials, err)
		}
	})
	t.Run("fails when the user is inactive", func(t *testing.T) {
		u := mustNewUser(t)
		u.Deactivate()

		if err := u.Authenticate("test1234"); !errors.Is(err, ErrUserInactive) {
			t.Errorf("expect %v, got %v", ErrUserInactive, err)
		}
	})
}

func mustNewUser(t *testing.T) *User {
	t.Helper()

	u, err := New("Humberto", "humberto@shop.com", "test1234", RoleOwner)
	if err != nil {
		t.Fatalf("failed to build user: %v", err)
	}
	return u
}
