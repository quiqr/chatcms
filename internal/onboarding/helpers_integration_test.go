//go:build integration

package onboarding_test

import (
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms/internal/gitsync"
)

func quiet() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func gitBinary(t *testing.T) string {
	t.Helper()

	if bin := os.Getenv("CHATCMS_GIT_BIN"); bin != "" {
		return bin
	}
	found, err := exec.LookPath("git")
	if err != nil {
		t.Skipf("git is not available: %v", err)
	}
	return found
}

func gitRun(t *testing.T, cfg gitsync.Config, dir string, args ...string) {
	t.Helper()

	if err := gitsync.RunForTest(t.Context(), cfg, dir, t.TempDir(), args...); err != nil {
		t.Fatalf("git %v failed: %v", args, err)
	}
}

func write(t *testing.T, path, contents string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func copyTree(t *testing.T, from, to string) {
	t.Helper()

	err := filepath.WalkDir(from, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(from, path)
		if err != nil || relative == "." {
			return err
		}
		if entry.IsDir() {
			return os.MkdirAll(filepath.Join(to, relative), 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(to, relative), data, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}

// contentOf reads every file a project's working copy holds, which is how two
// imports are compared.
func contentOf(t *testing.T, h *harness, projectID uuid.UUID) map[string]string {
	t.Helper()

	repository, err := h.sites.Open(t.Context(), projectID.String())
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()

	out := map[string]string{}
	err = fs.WalkDir(repository.FS(), ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		data, err := fs.ReadFile(repository.FS(), name)
		if err != nil {
			return err
		}
		out[name] = string(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}
