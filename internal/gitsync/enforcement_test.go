package gitsync

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// gitWord matches a mention of git that is not part of a longer word, so
// "gitsync" and "digit" do not trip it.
var gitWord = regexp.MustCompile(`(?i)\bgit\b`)

const runsACommand = "exec.Command"

// TestOnlyGitsyncRunsGit enforces the rule in docs/ARCHITECTURE.md: every git
// invocation originates here, so a different version control system could be
// put behind the same interface.
//
// The check is a heuristic by design: a file that both starts a process and
// talks about git is almost certainly running git, and the cost of a false
// positive is a sentence in a comment, while the cost of a miss is a second
// place that knows how git works.
func TestOnlyGitsyncRunsGit(t *testing.T) {
	root := repositoryRoot(t)

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if skipDir(entry.Name()) {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(entry.Name(), ".go") {
			return nil
		}

		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if strings.HasPrefix(filepath.ToSlash(relative), "internal/gitsync/") {
			return nil
		}

		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(source)

		if strings.Contains(text, runsACommand) && gitWord.MatchString(text) {
			t.Errorf("%s starts a process and mentions git; every git invocation belongs in internal/gitsync", relative)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking the sources failed: %v", err)
	}
}

func TestOnlyGitsyncRunsGitCatchesAnOffender(t *testing.T) {
	// The rule is worth nothing if it cannot fail, so prove the check does.
	offending := "package worker\n\nimport \"os/exec\"\n\nfunc pull() { exec.Command(\"git\", \"pull\").Run() }\n"

	if !(strings.Contains(offending, runsACommand) && gitWord.MatchString(offending)) {
		t.Error("the check would not notice a package shelling out to git")
	}

	innocent := "package model\n\nfunc digits() int { return 2 }\n"
	if strings.Contains(innocent, runsACommand) && gitWord.MatchString(innocent) {
		t.Error("the check trips on a file that neither runs a process nor mentions git")
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()

	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("finding the repository root failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Skipf("the repository root is not readable from here: %v", err)
	}
	return root
}

func skipDir(name string) bool {
	switch name {
	case ".git", ".jj", "node_modules", "vendor", "result", ".direnv", ".pgdata":
		return true
	}
	return false
}

// TestTheGateProvidesThePinnedGit proves the engine is exercised against the
// binary a deployment will use, not whatever happens to be on a developer's
// machine. In the gate CHATCMS_GIT_BIN names the pinned binary; locally it is
// unset and this test says so rather than failing.
func TestTheGateProvidesThePinnedGit(t *testing.T) {
	bin := os.Getenv("CHATCMS_GIT_BIN")
	if bin == "" {
		t.Skip("CHATCMS_GIT_BIN is not set, so these tests ran against the git on PATH")
	}

	if !strings.HasPrefix(bin, "/nix/store/") {
		t.Errorf("CHATCMS_GIT_BIN = %q, want the pinned binary from the flake", bin)
	}

	cfg := Config{GitBin: bin, System: Identity{Name: "t", Email: "t@example.org"}}
	out, err := newRunner(cfg, t.TempDir(), t.TempDir()).run(t.Context(), "--version")
	if err != nil {
		t.Fatalf("running the pinned git failed: %v", err)
	}
	if !strings.HasPrefix(out, "git version") {
		t.Errorf("git --version said %q", out)
	}
}
