package gitsync

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// testConfig uses the pinned git binary when one is configured, which is how
// the gate runs, and otherwise the one on PATH, which is how a developer runs.
func testConfig(t *testing.T) Config {
	t.Helper()

	bin := os.Getenv("CHATCMS_GIT_BIN")
	if bin == "" {
		found, err := exec.LookPath("git")
		if err != nil {
			t.Skipf("git is not available: %v", err)
		}
		bin = found
	}

	return Config{
		GitBin: bin,
		System: Identity{Name: "Quiqr ChatCMS", Email: "chatcms@example.org"},
	}
}

func editor(name string) Identity {
	return Identity{Name: name, Email: name + "@example.org"}
}

// bareRemote builds an empty remote with one version in it, so a clone has
// something to clone and a branch to be on.
func bareRemote(t *testing.T, cfg Config) string {
	t.Helper()

	remote := filepath.Join(t.TempDir(), "remote.git")
	run(t, cfg, "", "init", "--bare", "--initial-branch=main", remote)

	seed := filepath.Join(t.TempDir(), "seed")
	run(t, cfg, "", "clone", "--", remote, seed)
	writeFile(t, filepath.Join(seed, "content", "_index.md"), "---\ntitle: Home\n---\n")
	run(t, cfg, seed, "add", "--all")
	run(t, cfg, seed, "-c", "user.name=Seed", "-c", "user.email=seed@example.org",
		"commit", "--message=First version")
	run(t, cfg, seed, "push", "origin", "HEAD:refs/heads/main")

	return remote
}

// cloned gives a working copy of a fresh remote, and the remote's path.
func cloned(t *testing.T) (Repo, string, Config) {
	t.Helper()

	cfg := testConfig(t)
	remote := bareRemote(t, cfg)
	dir := filepath.Join(t.TempDir(), "site")

	repository, err := Clone(t.Context(), cfg, remote, dir)
	if err != nil {
		t.Fatalf("Clone() error = %v, want nil", err)
	}
	t.Cleanup(func() { repository.Close() })

	return repository, remote, cfg
}

// elsewhere is a second working copy of the same remote, standing in for
// someone editing the site outside ChatCMS.
func elsewhere(t *testing.T, cfg Config, remote string) string {
	t.Helper()

	dir := filepath.Join(t.TempDir(), "elsewhere")
	run(t, cfg, "", "clone", "--", remote, dir)
	return dir
}

// pushFrom makes a version in another working copy and sends it to the remote.
func pushFrom(t *testing.T, cfg Config, dir, path, contents, message string) {
	t.Helper()

	writeFile(t, filepath.Join(dir, filepath.FromSlash(path)), contents)
	run(t, cfg, dir, "add", "--all")
	run(t, cfg, dir, "-c", "user.name=Someone", "-c", "user.email=someone@example.org",
		"commit", "--message="+message)
	run(t, cfg, dir, "push", "origin", "HEAD:refs/heads/main")
}

func run(t *testing.T, cfg Config, dir string, args ...string) string {
	t.Helper()

	home := t.TempDir()
	out, err := newRunner(cfg, dir, home).run(context.Background(), args...)
	if err != nil {
		t.Fatalf("git %v failed: %v", args, err)
	}
	return out
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func save(t *testing.T, repository Repo, who, message string, writes map[string]string) Version {
	t.Helper()

	contents := make(map[string][]byte, len(writes))
	for name, body := range writes {
		contents[name] = []byte(body)
	}

	version, err := repository.Save(t.Context(), Save{
		Message: message,
		Author:  editor(who),
		Writes:  contents,
	})
	if err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}
	return version
}

// emptyBareRemote is a remote with nothing in it, which is what a site
// created from an export is pushed to.
func emptyBareRemote(t *testing.T, cfg Config) string {
	t.Helper()

	remote := filepath.Join(t.TempDir(), "empty-remote.git")
	run(t, cfg, "", "init", "--bare", "--initial-branch=main", remote)
	return remote
}
