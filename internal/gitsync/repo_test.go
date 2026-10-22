package gitsync

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/quiqr/chatcms/internal/redact"
)

func TestRunnerUsesTheConfiguredBinary(t *testing.T) {
	cfg := testConfig(t)
	cfg.GitBin = filepath.Join(t.TempDir(), "definitely-not-git")

	_, err := Open(t.Context(), cfg, t.TempDir())
	if err == nil {
		t.Fatal("Open() error = nil, want a failure from the configured binary")
	}
	if errors.Is(err, ErrNotAWorkingCopy) {
		return
	}
	if !strings.Contains(err.Error(), "definitely-not-git") && !strings.Contains(err.Error(), "git") {
		t.Errorf("error = %v, want it to come from running the configured binary", err)
	}
}

func TestRunnerIsNonInteractive(t *testing.T) {
	repository, _, cfg := cloned(t)

	dir := repository.(*repo).dir
	hooks := filepath.Join(dir, ".git", "hooks")
	if err := os.MkdirAll(hooks, 0o755); err != nil {
		t.Fatal(err)
	}
	hook := filepath.Join(hooks, "pre-commit")
	if err := os.WriteFile(hook, []byte("#!/bin/sh\necho a hook ran >&2\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := repository.Save(t.Context(), Save{
		Message: "A save that a hook would block",
		Author:  editor("joris"),
		Writes:  map[string][]byte{"content/about.md": []byte("about")},
	}); err != nil {
		t.Fatalf("Save() error = %v, want the hook to be ignored", err)
	}

	_ = cfg
}

func TestRunnerIgnoresGlobalConfig(t *testing.T) {
	repository, _, cfg := cloned(t)

	home := filepath.Join(filepath.Dir(repository.(*repo).dir), ".githome")
	if err := os.WriteFile(filepath.Join(home, ".gitconfig"),
		[]byte("[user]\n\tname = Should Not Be Used\n\temail = wrong@example.org\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	version := save(t, repository, "joris", "A save", map[string]string{"content/a.md": "a"})
	if version.Author.Name != "joris" {
		t.Errorf("author = %q, want the editor rather than any configured identity", version.Author.Name)
	}

	_ = cfg
}

func TestErrorScrubsCredentials(t *testing.T) {
	cases := map[string]struct {
		in   string
		gone string
	}{
		"a token in a URL": {
			in:   "fatal: could not read from https://user:ghp_secret123@github.com/a/b.git",
			gone: "ghp_secret123",
		},
		"a named token": {
			in:   "remote: rejected; token=abc123def",
			gone: "abc123def",
		},
		"a named password": {
			in:   "fatal: password: hunter2",
			gone: "hunter2",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := redact.Secrets(tc.in)
			if strings.Contains(got, tc.gone) {
				t.Errorf("scrub(%q) = %q, want %q removed", tc.in, got, tc.gone)
			}
			if !strings.Contains(got, "redacted") {
				t.Errorf("scrub(%q) = %q, want it to say something was removed", tc.in, got)
			}
		})
	}
}

func TestScrubKeepsOrdinaryMessages(t *testing.T) {
	message := "fatal: couldn't find remote ref refs/heads/main"
	if got := redact.Secrets(message); got != message {
		t.Errorf("scrub(%q) = %q, want it untouched", message, got)
	}
}

func TestCloneBringsContent(t *testing.T) {
	repository, _, _ := cloned(t)

	contents, err := fs.ReadFile(repository.FS(), "content/_index.md")
	if err != nil {
		t.Fatalf("reading the cloned content failed: %v", err)
	}
	if !strings.Contains(string(contents), "title: Home") {
		t.Errorf("content = %q, want the remote's content", contents)
	}
}

func TestOpenTouchesNothing(t *testing.T) {
	repository, remote, cfg := cloned(t)
	dir := repository.(*repo).dir

	// Move the remote out of the way: opening must not need it.
	if err := os.Rename(remote, remote+".moved"); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(t.Context(), cfg, dir)
	if err != nil {
		t.Fatalf("Open() error = %v, want nil with no remote present", err)
	}
	defer reopened.Close()

	if _, err := fs.ReadFile(reopened.FS(), "content/_index.md"); err != nil {
		t.Errorf("reading after opening failed: %v", err)
	}
}

func TestOpenRejectsAPlainDirectory(t *testing.T) {
	cfg := testConfig(t)
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "notes.md"), "just a file")

	_, err := Open(t.Context(), cfg, dir)
	if !errors.Is(err, ErrNotAWorkingCopy) {
		t.Fatalf("Open() error = %v, want ErrNotAWorkingCopy", err)
	}

	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Errorf("the directory now holds %d entries, want the one file it had", len(entries))
	}
}

func TestCloneRefusesToOverwrite(t *testing.T) {
	repository, remote, cfg := cloned(t)
	dir := repository.(*repo).dir

	_, err := Clone(t.Context(), cfg, remote, dir)
	if !errors.Is(err, ErrAlreadyAWorkingCopy) {
		t.Errorf("Clone() error = %v, want ErrAlreadyAWorkingCopy", err)
	}
}

func TestContentFSReadsFiles(t *testing.T) {
	repository, _, _ := cloned(t)
	save(t, repository, "joris", "Add a page", map[string]string{
		"content/about.md": "---\ntitle: About\n---\nHello.\n",
	})

	contents, err := fs.ReadFile(repository.FS(), "content/about.md")
	if err != nil {
		t.Fatalf("reading failed: %v", err)
	}
	if !strings.Contains(string(contents), "Hello.") {
		t.Errorf("content = %q", contents)
	}
}

func TestContentFSHidesVersionControl(t *testing.T) {
	repository, _, _ := cloned(t)

	entries, err := fs.ReadDir(repository.FS(), ".")
	if err != nil {
		t.Fatalf("listing failed: %v", err)
	}
	for _, entry := range entries {
		if entry.Name() == ".git" {
			t.Fatal("version control's own directory is visible as content")
		}
	}

	if _, err := fs.ReadFile(repository.FS(), ".git/config"); err == nil {
		t.Error("version control's configuration is readable as content")
	}

	err = fs.WalkDir(repository.FS(), ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasPrefix(path, ".git") {
			return errors.New("walked into version control's directory: " + path)
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}

func TestSaveMakesOneVersion(t *testing.T) {
	repository, _, _ := cloned(t)

	before, err := repository.History(t.Context(), 100)
	if err != nil {
		t.Fatal(err)
	}

	save(t, repository, "joris", "Add two pages", map[string]string{
		"content/a.md": "a",
		"content/b.md": "b",
	})

	after, err := repository.History(t.Context(), 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(before)+1 {
		t.Fatalf("history grew by %d, want 1", len(after)-len(before))
	}
	if len(after[0].Paths) != 2 {
		t.Errorf("the version changed %v, want both files", after[0].Paths)
	}
}

func TestSaveIsAttributedToTheEditor(t *testing.T) {
	repository, _, _ := cloned(t)

	version := save(t, repository, "marieke", "Edit the homepage", map[string]string{
		"content/_index.md": "---\ntitle: Welcome\n---\n",
	})

	if version.Author.Name != "marieke" || version.Author.Email != "marieke@example.org" {
		t.Errorf("author = %+v, want the editor", version.Author)
	}
	if version.Message != "Edit the homepage" {
		t.Errorf("message = %q", version.Message)
	}
}

func TestSaveWithNoChangeMakesNoVersion(t *testing.T) {
	repository, _, _ := cloned(t)
	save(t, repository, "joris", "Add a page", map[string]string{"content/a.md": "a"})

	_, err := repository.Save(t.Context(), Save{
		Message: "The same content again",
		Author:  editor("joris"),
		Writes:  map[string][]byte{"content/a.md": []byte("a")},
	})
	if !errors.Is(err, ErrNothingChanged) {
		t.Errorf("Save() error = %v, want ErrNothingChanged", err)
	}
}

func TestSaveCanDelete(t *testing.T) {
	repository, _, _ := cloned(t)
	save(t, repository, "joris", "Add a page", map[string]string{"content/a.md": "a"})

	version, err := repository.Save(t.Context(), Save{
		Message: "Remove the page",
		Author:  editor("joris"),
		Deletes: []string{"content/a.md"},
	})
	if err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}
	if len(version.Paths) != 1 || version.Paths[0] != "content/a.md" {
		t.Errorf("version paths = %v, want the removed file", version.Paths)
	}
	if _, err := fs.ReadFile(repository.FS(), "content/a.md"); err == nil {
		t.Error("the file is still readable after being removed")
	}
}

func TestSaveRejectsAnEscapingPath(t *testing.T) {
	repository, _, _ := cloned(t)
	dir := repository.(*repo).dir

	outside := filepath.Join(filepath.Dir(dir), "outside")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "escape")); err != nil {
		t.Skipf("symlinks are not available: %v", err)
	}

	cases := map[string]string{
		"a parent traversal": "../outside/a.md",
		"an absolute path":   "/etc/passwd",
		"version control":    ".git/config",
		"through a symlink":  "escape/a.md",
	}

	for name, path := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := repository.Save(t.Context(), Save{
				Message: "Should not be written",
				Author:  editor("joris"),
				Writes:  map[string][]byte{path: []byte("no")},
			})
			if !errors.Is(err, ErrEscapingPath) {
				t.Errorf("Save(%q) error = %v, want ErrEscapingPath", path, err)
			}
		})
	}

	if _, err := os.Stat(filepath.Join(outside, "a.md")); err == nil {
		t.Error("a file was written outside the site")
	}
}

func TestConcurrentSavesToOneSite(t *testing.T) {
	repository, _, _ := cloned(t)

	before, _ := repository.History(t.Context(), 100)

	var wait sync.WaitGroup
	errs := make([]error, 2)
	for i, name := range []string{"content/one.md", "content/two.md"} {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, errs[i] = repository.Save(t.Context(), Save{
				Message: "Save " + name,
				Author:  editor("joris"),
				Writes:  map[string][]byte{name: []byte(name)},
			})
		}()
	}
	wait.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("save %d failed: %v", i, err)
		}
	}

	after, _ := repository.History(t.Context(), 100)
	if len(after) != len(before)+2 {
		t.Fatalf("history grew by %d, want 2", len(after)-len(before))
	}
	for _, version := range after[:2] {
		if len(version.Paths) != 1 {
			t.Errorf("version %q changed %v, want one file", version.Message, version.Paths)
		}
	}
}

func TestConcurrentSavesToTwoSitesDoNotBlock(t *testing.T) {
	first, _, _ := cloned(t)
	second, _, _ := cloned(t)

	started := make(chan struct{})
	release := make(chan struct{})

	firstLock := siteLock(first.(*repo).dir)
	go func() {
		firstLock.Lock()
		close(started)
		<-release
		firstLock.Unlock()
	}()
	<-started

	done := make(chan error, 1)
	go func() {
		_, err := second.Save(t.Context(), Save{
			Message: "A save on another site",
			Author:  editor("joris"),
			Writes:  map[string][]byte{"content/other.md": []byte("other")},
		})
		done <- err
	}()

	if err := <-done; err != nil {
		t.Errorf("the second site's save failed while the first was locked: %v", err)
	}
	close(release)
}

func TestHistoryIsNewestFirst(t *testing.T) {
	repository, _, _ := cloned(t)
	save(t, repository, "joris", "First edit", map[string]string{"content/a.md": "a"})
	save(t, repository, "joris", "Second edit", map[string]string{"content/a.md": "aa"})

	versions, err := repository.History(t.Context(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 2 {
		t.Fatalf("got %d versions, want 2", len(versions))
	}
	if versions[0].Message != "Second edit" || versions[1].Message != "First edit" {
		t.Errorf("order = %q, %q, want newest first", versions[0].Message, versions[1].Message)
	}
	if versions[0].Time.Before(versions[1].Time) {
		t.Error("the newest version is older than the one before it")
	}
}

func TestVersionNamesWhatItChanged(t *testing.T) {
	repository, _, _ := cloned(t)
	save(t, repository, "joris", "An edit", map[string]string{
		"content/a.md":     "a",
		"content/sub/b.md": "b",
	})

	versions, _ := repository.History(t.Context(), 1)
	got := map[string]bool{}
	for _, path := range versions[0].Paths {
		got[path] = true
	}
	for _, want := range []string{"content/a.md", "content/sub/b.md"} {
		if !got[want] {
			t.Errorf("version paths = %v, want %q among them", versions[0].Paths, want)
		}
	}
}

func TestHistoryOnAnEmptyRepository(t *testing.T) {
	cfg := testConfig(t)
	dir := filepath.Join(t.TempDir(), "empty")
	run(t, cfg, "", "init", "--initial-branch=main", dir)

	repository, err := Open(t.Context(), cfg, dir)
	if err != nil {
		t.Fatalf("Open() error = %v, want nil", err)
	}
	defer repository.Close()

	versions, err := repository.History(t.Context(), 10)
	if err != nil {
		t.Fatalf("History() error = %v, want nil", err)
	}
	if len(versions) != 0 {
		t.Errorf("got %d versions, want none", len(versions))
	}
}

func TestHistoryLimit(t *testing.T) {
	repository, _, _ := cloned(t)
	for i := range 5 {
		save(t, repository, "joris", "Edit", map[string]string{
			"content/a.md": strings.Repeat("a", i+1),
		})
	}

	versions, err := repository.History(t.Context(), 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 3 {
		t.Errorf("got %d versions, want the limit of 3", len(versions))
	}
}

func TestSaveWithNoPathsSavesEverything(t *testing.T) {
	repository, _, _ := cloned(t)
	dir := repository.(*repo).dir

	writeFile(t, filepath.Join(dir, "content", "one.md"), "one")
	writeFile(t, filepath.Join(dir, "content", "two.md"), "two")

	version, err := repository.Save(t.Context(), Save{
		Message: "Import the site",
		Author:  editor("joris"),
	})
	if err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}

	paths := map[string]bool{}
	for _, path := range version.Paths {
		paths[path] = true
	}
	for _, want := range []string{"content/one.md", "content/two.md"} {
		if !paths[want] {
			t.Errorf("version paths = %v, want %q among them", version.Paths, want)
		}
	}

	status, _ := repository.Status(t.Context())
	if status.UnsavedWork {
		t.Error("something was left unsaved")
	}
}

func TestSaveWithNoPathsAndNothingToSave(t *testing.T) {
	repository, _, _ := cloned(t)

	_, err := repository.Save(t.Context(), Save{
		Message: "Nothing at all",
		Author:  editor("joris"),
	})
	if !errors.Is(err, ErrNothingChanged) {
		t.Errorf("Save() error = %v, want ErrNothingChanged", err)
	}
}
