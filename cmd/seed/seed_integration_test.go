//go:build integration

package main

import (
	"io"
	"log/slog"
	"testing"

	"github.com/quiqr/chatcms/internal/auth"
	"github.com/quiqr/chatcms/internal/policy"
	"github.com/quiqr/chatcms/internal/store"
	"github.com/quiqr/chatcms/internal/tenancy"
	"github.com/quiqr/chatcms/internal/testsupport"
)

func TestEveryRoleCanSignInAfterSeeding(t *testing.T) {
	t.Setenv("CHATCMS_STATE_DIR", t.TempDir())
	t.Setenv("CHATCMS_SEED_REMOTE", "")

	db, err := store.Open(t.Context(), testsupport.Postgres(t))
	if err != nil {
		t.Fatalf("store.Open() error = %v, want nil", err)
	}
	t.Cleanup(db.Close)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := seed(t.Context(), db, logger); err != nil {
		t.Fatalf("seed() error = %v, want nil", err)
	}

	identity := auth.NewService(db, logger, "http://127.0.0.1:18090")
	service := tenancy.NewService(tenancy.NewStore(db), identity)

	for _, tc := range []struct {
		email string
		role  policy.Role
	}{
		{"owner@example.org", policy.RoleImplementer},
		{"marieke@korenaar.example", policy.RoleEditor},
		{"sam@korenaar.example", policy.RoleViewer},
	} {
		t.Run(tc.email, func(t *testing.T) {
			who, _, _, err := identity.SignIn(t.Context(), tc.email, passphrase, "test", false)
			if err != nil {
				t.Fatalf("SignIn(%q) error = %v, want nil", tc.email, err)
			}

			projects, err := service.ListProjects(t.Context(), who, "technative")
			if err != nil {
				t.Fatalf("ListProjects() error = %v, want nil", err)
			}
			var role string
			for _, project := range projects {
				if project.Slug == "korenaar" {
					role = project.YourRole
				}
			}
			if role != string(tc.role) {
				t.Errorf("role on korenaar = %q, want %q", role, tc.role)
			}
		})
	}

	if _, _, _, err := identity.SignIn(t.Context(), "admin@example.org", passphrase, "test", false); err != nil {
		t.Errorf("SignIn(admin) error = %v, want nil", err)
	}

	if _, _, _, err := identity.SignIn(t.Context(), "noor@korenaar.example", passphrase, "test", false); err == nil {
		t.Errorf("SignIn(noor) error = nil, want a refusal, because her invitation is still open")
	}
}
