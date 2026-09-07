package user

import (
	"errors"
	"testing"
	"time"
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
			{"blank name", "   ", "humberto@shop.com", "test1234", RoleOwner, ErrInvalidName},
			{"email without at sign", "Humberto", "humbertoshop.com", "test1234", RoleOwner, ErrInvalidEmail},
			{"empty email", "Humberto", "", "test1234", RoleOwner, ErrInvalidEmail},
			{"password too short", "Humberto", "humberto@shop.com", "123", RoleOwner, ErrWeakPassword},
			{"password one char below minimum", "Humberto", "humberto@shop.com", "1234567", RoleOwner, ErrWeakPassword},
			{"unknown role", "Humberto", "humberto@shop.com", "test1234", Role("manager"), ErrInvalidRole},
			{"empty role", "Humberto", "humberto@shop.com", "test1234", Role(""), ErrInvalidRole},
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

	t.Run("accepts a password exactly at the minimum length", func(t *testing.T) {
		_, err := New("Humberto", "humberto@shop.com", "12345678", RoleOwner)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("normalizes name and email", func(t *testing.T) {
		u, err := New("  Humberto  ", "  HUMBERTO@Shop.com  ", "test1234", RoleOwner)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if u.Name() != "Humberto" {
			t.Errorf("expected trimmed name %q, got %q", "Humberto", u.Name())
		}
		if u.Email() != "humberto@shop.com" {
			t.Errorf("expected normalized email %q, got %q", "humberto@shop.com", u.Email())
		}
	})

	t.Run("starts at version 1 with createdAt equal to updatedAt", func(t *testing.T) {
		u := mustNewUser(t)
		if u.Version() != 1 {
			t.Errorf("expected version 1, got %d", u.Version())
		}
		if !u.CreatedAt().Equal(u.UpdatedAt()) {
			t.Errorf("expected createdAt %v to equal updatedAt %v", u.CreatedAt(), u.UpdatedAt())
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

func TestRename(t *testing.T) {
	t.Run("renames and touches updatedAt", func(t *testing.T) {
		u := mustNewUser(t)
		before := u.UpdatedAt()
		time.Sleep(time.Millisecond)

		if err := u.Rename("New Name"); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if u.Name() != "New Name" {
			t.Errorf("expected name %q, got %q", "New Name", u.Name())
		}
		if !u.UpdatedAt().After(before) {
			t.Errorf("expected updatedAt to advance past %v, got %v", before, u.UpdatedAt())
		}
	})

	t.Run("trims the name", func(t *testing.T) {
		u := mustNewUser(t)
		if err := u.Rename("  Trimmed  "); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if u.Name() != "Trimmed" {
			t.Errorf("expected trimmed name %q, got %q", "Trimmed", u.Name())
		}
	})

	t.Run("rejects a blank name and leaves the user unchanged", func(t *testing.T) {
		u := mustNewUser(t)
		before := u.UpdatedAt()

		if err := u.Rename("   "); !errors.Is(err, ErrInvalidName) {
			t.Errorf("expected %v, got %v", ErrInvalidName, err)
		}
		if u.Name() != "Humberto" {
			t.Errorf("expected name to remain %q, got %q", "Humberto", u.Name())
		}
		if !u.UpdatedAt().Equal(before) {
			t.Errorf("expected updatedAt to remain %v, got %v", before, u.UpdatedAt())
		}
	})
}

func TestChangeEmail(t *testing.T) {
	t.Run("changes and normalizes the email, touching updatedAt", func(t *testing.T) {
		u := mustNewUser(t)
		before := u.UpdatedAt()
		time.Sleep(time.Millisecond)

		if err := u.ChangeEmail("  NEW@Shop.com  "); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if u.Email() != "new@shop.com" {
			t.Errorf("expected normalized email %q, got %q", "new@shop.com", u.Email())
		}
		if !u.UpdatedAt().After(before) {
			t.Errorf("expected updatedAt to advance past %v, got %v", before, u.UpdatedAt())
		}
	})

	t.Run("rejects an invalid email and leaves the user unchanged", func(t *testing.T) {
		u := mustNewUser(t)
		before := u.UpdatedAt()

		if err := u.ChangeEmail("not-an-email"); !errors.Is(err, ErrInvalidEmail) {
			t.Errorf("expected %v, got %v", ErrInvalidEmail, err)
		}
		if u.Email() != "humberto@shop.com" {
			t.Errorf("expected email to remain %q, got %q", "humberto@shop.com", u.Email())
		}
		if !u.UpdatedAt().Equal(before) {
			t.Errorf("expected updatedAt to remain %v, got %v", before, u.UpdatedAt())
		}
	})
}

func TestChangePassword(t *testing.T) {
	t.Run("changes the password hash and touches updatedAt", func(t *testing.T) {
		u := mustNewUser(t)
		oldHash := u.PasswordHash()
		before := u.UpdatedAt()
		time.Sleep(time.Millisecond)

		if err := u.ChangePassword("newpassword1"); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if u.PasswordHash() == oldHash {
			t.Errorf("expected password hash to change")
		}
		if err := u.Authenticate("newpassword1"); err != nil {
			t.Errorf("expected authentication with the new password to succeed, got %v", err)
		}
		if !u.UpdatedAt().After(before) {
			t.Errorf("expected updatedAt to advance past %v, got %v", before, u.UpdatedAt())
		}
	})

	t.Run("rejects a weak password and leaves the user unchanged", func(t *testing.T) {
		u := mustNewUser(t)
		oldHash := u.PasswordHash()

		if err := u.ChangePassword("weak"); !errors.Is(err, ErrWeakPassword) {
			t.Errorf("expected %v, got %v", ErrWeakPassword, err)
		}
		if u.PasswordHash() != oldHash {
			t.Errorf("expected password hash to remain unchanged")
		}
	})
}

func TestChangeRole(t *testing.T) {
	t.Run("changes the role and touches updatedAt", func(t *testing.T) {
		u := mustNewUser(t)
		before := u.UpdatedAt()
		time.Sleep(time.Millisecond)

		if err := u.ChangeRole(RoleClerk); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if u.Role() != RoleClerk {
			t.Errorf("expected role %v, got %v", RoleClerk, u.Role())
		}
		if !u.UpdatedAt().After(before) {
			t.Errorf("expected updatedAt to advance past %v, got %v", before, u.UpdatedAt())
		}
	})

	t.Run("rejects an unknown role and leaves the user unchanged", func(t *testing.T) {
		u := mustNewUser(t)
		before := u.UpdatedAt()

		if err := u.ChangeRole("manager"); !errors.Is(err, ErrInvalidRole) {
			t.Errorf("expected %v, got %v", ErrInvalidRole, err)
		}
		if u.Role() != RoleOwner {
			t.Errorf("expected role to remain %v, got %v", RoleOwner, u.Role())
		}
		if !u.UpdatedAt().Equal(before) {
			t.Errorf("expected updatedAt to remain %v, got %v", before, u.UpdatedAt())
		}
	})
}

func TestActivateDeactivate(t *testing.T) {
	t.Run("deactivate flips isActive and touches updatedAt", func(t *testing.T) {
		u := mustNewUser(t)
		before := u.UpdatedAt()
		time.Sleep(time.Millisecond)

		u.Deactivate()

		if u.IsActive() {
			t.Errorf("expected IsActive to be false")
		}
		if !u.UpdatedAt().After(before) {
			t.Errorf("expected updatedAt to advance past %v, got %v", before, u.UpdatedAt())
		}
	})

	t.Run("deactivate is idempotent and does not touch again", func(t *testing.T) {
		u := mustNewUser(t)
		u.Deactivate()
		before := u.UpdatedAt()
		time.Sleep(time.Millisecond)

		u.Deactivate()

		if !u.UpdatedAt().Equal(before) {
			t.Errorf("expected updatedAt to remain %v, got %v", before, u.UpdatedAt())
		}
	})

	t.Run("activate flips isActive and touches updatedAt", func(t *testing.T) {
		u := mustNewUser(t)
		u.Deactivate()
		before := u.UpdatedAt()
		time.Sleep(time.Millisecond)

		u.Activate()

		if !u.IsActive() {
			t.Errorf("expected IsActive to be true")
		}
		if !u.UpdatedAt().After(before) {
			t.Errorf("expected updatedAt to advance past %v, got %v", before, u.UpdatedAt())
		}
	})

	t.Run("activate is idempotent and does not touch again", func(t *testing.T) {
		u := mustNewUser(t)
		before := u.UpdatedAt()
		time.Sleep(time.Millisecond)

		u.Activate()

		if !u.UpdatedAt().Equal(before) {
			t.Errorf("expected updatedAt to remain %v, got %v", before, u.UpdatedAt())
		}
	})
}

func TestSetVersion(t *testing.T) {
	u := mustNewUser(t)
	u.SetVersion(5)
	if u.Version() != 5 {
		t.Errorf("expected version 5, got %d", u.Version())
	}
}

func TestRestore(t *testing.T) {
	createdAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)

	u := Restore(
		"11111111-1111-1111-1111-111111111111",
		"Restored Name",
		"restored@shop.com",
		"hashed-password",
		RoleClerk,
		false,
		3,
		createdAt,
		updatedAt,
	)

	if u.ID() != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("expected id %q, got %q", "11111111-1111-1111-1111-111111111111", u.ID())
	}
	if u.Name() != "Restored Name" {
		t.Errorf("expected name %q, got %q", "Restored Name", u.Name())
	}
	if u.Email() != "restored@shop.com" {
		t.Errorf("expected email %q, got %q", "restored@shop.com", u.Email())
	}
	if u.PasswordHash() != "hashed-password" {
		t.Errorf("expected password hash %q, got %q", "hashed-password", u.PasswordHash())
	}
	if u.Role() != RoleClerk {
		t.Errorf("expected role %v, got %v", RoleClerk, u.Role())
	}
	if u.IsActive() {
		t.Errorf("expected IsActive to be false")
	}
	if u.Version() != 3 {
		t.Errorf("expected version 3, got %d", u.Version())
	}
	if !u.CreatedAt().Equal(createdAt) {
		t.Errorf("expected createdAt %v, got %v", createdAt, u.CreatedAt())
	}
	if !u.UpdatedAt().Equal(updatedAt) {
		t.Errorf("expected updatedAt %v, got %v", updatedAt, u.UpdatedAt())
	}
}

func TestValidateUserID(t *testing.T) {
	t.Run("accepts a valid uuid", func(t *testing.T) {
		if err := ValidateUserID("11111111-1111-1111-1111-111111111111"); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("rejects a malformed id", func(t *testing.T) {
		if err := ValidateUserID("not-a-uuid"); !errors.Is(err, ErrInvalidID) {
			t.Errorf("expected %v, got %v", ErrInvalidID, err)
		}
	})
}

func TestRoleIsValid(t *testing.T) {
	tests := []struct {
		role Role
		want bool
	}{
		{RoleOwner, true},
		{RoleClerk, true},
		{Role("manager"), false},
		{Role(""), false},
	}

	for _, tt := range tests {
		if got := tt.role.IsValid(); got != tt.want {
			t.Errorf("Role(%q).IsValid() = %v, want %v", tt.role, got, tt.want)
		}
	}
}

func mustNewUser(t *testing.T) *User {
	t.Helper()

	u, err := New("Humberto", "humberto@shop.com", "test1234", RoleOwner)
	if err != nil {
		t.Fatalf("failed to build user: %v", err)
	}
	return u
}
