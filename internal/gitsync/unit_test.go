package gitsync

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestOperationErrorSaysWhatFailed(t *testing.T) {
	cause := errors.New("exit status 128")
	err := &OperationError{Operation: "fetch", Detail: "could not read from remote", Err: cause}

	if !strings.Contains(err.Error(), "fetch") {
		t.Errorf("error = %q, want it to name the operation", err)
	}
	if !strings.Contains(err.Error(), "could not read from remote") {
		t.Errorf("error = %q, want it to carry the detail", err)
	}
	if !errors.Is(err, cause) {
		t.Error("the cause cannot be reached through the error")
	}

	bare := &OperationError{Operation: "push", Err: cause}
	if !strings.Contains(bare.Error(), "exit status 128") {
		t.Errorf("error without a detail = %q, want the cause", bare)
	}
}

func TestConfigMustBeComplete(t *testing.T) {
	cases := map[string]Config{
		"no binary":   {System: Identity{Name: "a", Email: "a@example.org"}},
		"no identity": {GitBin: "/usr/bin/git"},
		"half an identity": {
			GitBin: "/usr/bin/git",
			System: Identity{Name: "a"},
		},
	}

	for name, cfg := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Open(t.Context(), cfg, t.TempDir()); err == nil {
				t.Error("Open() error = nil, want a configuration error")
			}
			if _, err := Clone(t.Context(), cfg, "somewhere", t.TempDir()); err == nil {
				t.Error("Clone() error = nil, want a configuration error")
			}
		})
	}
}

func TestSSHCommandUsesOnlyTheConfiguredKey(t *testing.T) {
	withKey := sshCommand("/run/secrets/site-key")
	if !strings.Contains(withKey, "-i /run/secrets/site-key") {
		t.Errorf("sshCommand = %q, want the key", withKey)
	}
	if !strings.Contains(withKey, "IdentitiesOnly=yes") {
		t.Errorf("sshCommand = %q, want it to refuse other identities", withKey)
	}

	without := sshCommand("")
	if strings.Contains(without, "-i ") {
		t.Errorf("sshCommand = %q, want no key when none is configured", without)
	}
	for _, command := range []string{withKey, without} {
		if !strings.Contains(command, "BatchMode=yes") {
			t.Errorf("sshCommand = %q, want it never to ask a question", command)
		}
	}
}

func TestIdentityFormatsAsAPerson(t *testing.T) {
	who := Identity{Name: "Joris", Email: "joris@example.org"}
	if got := who.String(); got != "Joris <joris@example.org>" {
		t.Errorf("String() = %q", got)
	}
}

func TestSplitHelpers(t *testing.T) {
	if got := splitLines(""); got != nil {
		t.Errorf("splitLines(empty) = %v, want nil", got)
	}
	if got := splitLines("a\nb\n"); len(got) != 2 {
		t.Errorf("splitLines = %v, want two lines", got)
	}
	if got := splitNul(""); got != nil {
		t.Errorf("splitNul(empty) = %v, want nil", got)
	}
	if got := splitNul("a\x00b\x00"); len(got) != 2 {
		t.Errorf("splitNul = %v, want two fields", got)
	}
}

func TestSaveNeedsAMessageAndAPerson(t *testing.T) {
	repository, _, _ := cloned(t)

	_, err := repository.Save(t.Context(), Save{
		Author: editor("joris"),
		Writes: map[string][]byte{"content/a.md": []byte("a")},
	})
	if err == nil {
		t.Error("Save() without a message error = nil, want a rejection")
	}

	_, err = repository.Save(t.Context(), Save{
		Message: "An edit",
		Writes:  map[string][]byte{"content/a.md": []byte("a")},
	})
	if err == nil {
		t.Error("Save() without an author error = nil, want a rejection")
	}
}

func TestContentFSReadDir(t *testing.T) {
	repository, _, _ := cloned(t)
	save(t, repository, "joris", "Add pages", map[string]string{
		"content/a.md":     "a",
		"content/sub/b.md": "b",
	})

	entries, err := fs.ReadDir(repository.FS(), "content")
	if err != nil {
		t.Fatalf("ReadDir() error = %v, want nil", err)
	}

	names := map[string]bool{}
	for _, entry := range entries {
		names[entry.Name()] = true
	}
	for _, want := range []string{"a.md", "sub", "_index.md"} {
		if !names[want] {
			t.Errorf("ReadDir(content) = %v, want %q among them", names, want)
		}
	}

	top, err := fs.ReadDir(repository.FS(), ".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range top {
		if entry.Name() == versionControlDir {
			t.Error("ReadDir(.) shows version control's own directory")
		}
	}

	if _, err := fs.ReadDir(repository.FS(), ".git"); err == nil {
		t.Error("ReadDir(.git) succeeded, want it hidden")
	}
	if _, err := fs.ReadDir(repository.FS(), "../outside"); err == nil {
		t.Error("ReadDir on an invalid path succeeded")
	}
}

func TestContentFSRejectsAnInvalidPath(t *testing.T) {
	repository, _, _ := cloned(t)

	if _, err := repository.FS().Open("../escape"); err == nil {
		t.Error("Open on an escaping path succeeded")
	}
	if _, err := repository.FS().Open("content/absent.md"); err == nil {
		t.Error("Open on a missing file succeeded")
	}
}

func TestStatusOnARepositoryWithNoVersions(t *testing.T) {
	cfg := testConfig(t)
	dir := filepath.Join(t.TempDir(), "empty")
	run(t, cfg, "", "init", "--initial-branch=main", dir)

	repository, err := Open(t.Context(), cfg, dir)
	if err != nil {
		t.Fatalf("Open() error = %v, want nil", err)
	}
	defer repository.Close()

	status, err := repository.Status(t.Context())
	if err != nil {
		t.Fatalf("Status() error = %v, want nil", err)
	}
	if status.Unpublished != 0 || status.RemoteAhead != 0 {
		t.Errorf("status = %+v, want nothing counted on an empty site", status)
	}
}

func TestStatusCountsEverythingWhenNothingWasFetched(t *testing.T) {
	cfg := testConfig(t)
	dir := filepath.Join(t.TempDir(), "local-only")
	run(t, cfg, "", "init", "--initial-branch=main", dir)

	repository, err := Open(t.Context(), cfg, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()

	save(t, repository, "joris", "First", map[string]string{"content/a.md": "a"})
	save(t, repository, "joris", "Second", map[string]string{"content/b.md": "b"})

	status, err := repository.Status(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if status.Unpublished != 2 {
		t.Errorf("unpublished = %d, want every version when there is no remote state", status.Unpublished)
	}
}

func TestSitesOpenRejectsAMissingSite(t *testing.T) {
	sites := NewSites(testConfig(t), t.TempDir(), nil)

	if _, err := sites.Open(t.Context(), "absent"); !errors.Is(err, ErrNotAWorkingCopy) {
		t.Errorf("Open() error = %v, want ErrNotAWorkingCopy", err)
	}
	if _, err := sites.Open(t.Context(), "a/b"); !errors.Is(err, ErrUnusableSiteID) {
		t.Errorf("Open() error = %v, want ErrUnusableSiteID", err)
	}
	if _, err := sites.Clone(t.Context(), "a/b", "somewhere"); !errors.Is(err, ErrUnusableSiteID) {
		t.Errorf("Clone() error = %v, want ErrUnusableSiteID", err)
	}
}

func TestSitesListIgnoresStrayEntries(t *testing.T) {
	cfg := testConfig(t)
	root := filepath.Join(t.TempDir(), "sites")
	sites := NewSites(cfg, root, nil)

	repository, err := sites.Clone(t.Context(), "korenaar", bareRemote(t, cfg))
	if err != nil {
		t.Fatal(err)
	}
	repository.Close()

	if err := os.WriteFile(filepath.Join(root, "a-stray-file"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "half-made"), 0o700); err != nil {
		t.Fatal(err)
	}

	ids, err := sites.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != "korenaar" {
		t.Errorf("List() = %v, want only the site that has a working copy", ids)
	}
}

func TestRunForTestReachesGit(t *testing.T) {
	cfg := testConfig(t)
	dir := t.TempDir()

	if err := RunForTest(t.Context(), cfg, "", dir, "init", filepath.Join(dir, "new")); err != nil {
		t.Fatalf("RunForTest() error = %v, want nil", err)
	}
	if !isWorkingCopy(filepath.Join(dir, "new")) {
		t.Error("RunForTest did not create a working copy")
	}
}

func TestPublishOnASiteWithNothingToSend(t *testing.T) {
	repository, _, _ := cloned(t)

	result, err := repository.Publish(t.Context())
	if err != nil {
		t.Fatalf("Publish() error = %v, want nil", err)
	}
	if len(result.Published) != 0 {
		t.Errorf("published = %+v, want nothing", result.Published)
	}
}

func TestPublishReportsAnUnreachableRemote(t *testing.T) {
	repository, remote, _ := cloned(t)
	save(t, repository, "joris", "An edit", map[string]string{"content/a.md": "a"})

	if err := os.Rename(remote, remote+".gone"); err != nil {
		t.Fatal(err)
	}

	if _, err := repository.Publish(t.Context()); !errors.Is(err, ErrRemoteUnreachable) {
		t.Errorf("Publish() error = %v, want ErrRemoteUnreachable", err)
	}
}

func TestSyncIntoAnEmptyWorkingCopy(t *testing.T) {
	cfg := testConfig(t)
	remote := bareRemote(t, cfg)

	dir := filepath.Join(t.TempDir(), "empty")
	run(t, cfg, "", "init", "--initial-branch=main", dir)
	run(t, cfg, dir, "remote", "add", "origin", remote)

	repository, err := Open(t.Context(), cfg, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()

	result, err := repository.Sync(t.Context())
	if err != nil {
		t.Fatalf("Sync() error = %v, want nil", err)
	}
	if len(result.Arrived) == 0 {
		t.Fatal("nothing arrived into an empty working copy")
	}
	if _, err := fs.ReadFile(repository.FS(), "content/_index.md"); err != nil {
		t.Errorf("the remote's content is not readable: %v", err)
	}
}

func TestCreateMakesAWorkingCopyFromFiles(t *testing.T) {
	cfg := testConfig(t)
	remote := emptyBareRemote(t, cfg)
	sites := NewSites(cfg, filepath.Join(t.TempDir(), "sites"), nil)

	content := fstest.MapFS{
		"hugo.toml":              {Data: []byte("baseURL = 'https://example.org'\n")},
		"content/_index.md":      {Data: []byte("---\ntitle: Home\n---\n")},
		"quiqr/model/base.yaml":  {Data: []byte("singles: []\n")},
		".git/should-not-arrive": {Data: []byte("x")},
	}

	repository, err := sites.Create(t.Context(), "korenaar", remote, content,
		Identity{Name: "Quiqr ChatCMS", Email: "chatcms@example.org"}, "Import the site")
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	defer repository.Close()

	for _, name := range []string{"hugo.toml", "content/_index.md", "quiqr/model/base.yaml"} {
		if _, err := fs.ReadFile(repository.FS(), name); err != nil {
			t.Errorf("%s did not arrive: %v", name, err)
		}
	}
	if _, err := fs.ReadFile(repository.FS(), "should-not-arrive"); err == nil {
		t.Error("something from version control's own directory was copied in")
	}

	versions, err := repository.History(t.Context(), 10)
	if err != nil || len(versions) != 1 {
		t.Fatalf("history = %+v, %v, want one version", versions, err)
	}
	if versions[0].Message != "Import the site" {
		t.Errorf("message = %q", versions[0].Message)
	}

	// The remote was recorded, so publishing has somewhere to go.
	if _, err := repository.Publish(t.Context()); err != nil {
		t.Errorf("Publish() after Create error = %v, want nil", err)
	}
}

func TestCreateRefusesWhereAWorkingCopyExists(t *testing.T) {
	cfg := testConfig(t)
	sites := NewSites(cfg, filepath.Join(t.TempDir(), "sites"), nil)

	repository, err := sites.Clone(t.Context(), "korenaar", bareRemote(t, cfg))
	if err != nil {
		t.Fatal(err)
	}
	repository.Close()

	_, err = sites.Create(t.Context(), "korenaar", "", fstest.MapFS{}, testIdentity(), "Again")
	if !errors.Is(err, ErrAlreadyAWorkingCopy) {
		t.Errorf("Create() error = %v, want ErrAlreadyAWorkingCopy", err)
	}
}

func TestCreateWithNoRemote(t *testing.T) {
	cfg := testConfig(t)
	sites := NewSites(cfg, filepath.Join(t.TempDir(), "sites"), nil)

	repository, err := sites.Create(t.Context(), "korenaar", "",
		fstest.MapFS{"README.md": {Data: []byte("a site")}}, testIdentity(), "Import")
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	defer repository.Close()

	if _, err := fs.ReadFile(repository.FS(), "README.md"); err != nil {
		t.Errorf("the content did not arrive: %v", err)
	}
}

func TestReachableSaysWhetherTheRemoteAnswers(t *testing.T) {
	cfg := testConfig(t)
	root := filepath.Join(t.TempDir(), "sites")
	sites := NewSites(cfg, root, nil)

	remote := bareRemote(t, cfg)
	repository, err := sites.Clone(t.Context(), "korenaar", remote)
	if err != nil {
		t.Fatal(err)
	}
	repository.Close()

	if err := sites.Reachable(t.Context(), "korenaar", remote); err != nil {
		t.Errorf("Reachable() error = %v, want nil for a remote that is there", err)
	}

	err = sites.Reachable(t.Context(), "korenaar", filepath.Join(t.TempDir(), "absent.git"))
	if err == nil {
		t.Fatal("Reachable() error = nil for a remote that is not there")
	}
	if !errors.Is(err, ErrRemoteUnreachable) && !errors.Is(err, ErrKeyRefused) {
		t.Errorf("Reachable() error = %v, want it classified", err)
	}
}

func TestClassifyRemoteTellsRefusalFromUnreachable(t *testing.T) {
	refused := &OperationError{
		Operation: "ls-remote",
		Detail:    "git@example.org: Permission denied (publickey).",
	}
	if err := classifyRemote(refused); !errors.Is(err, ErrKeyRefused) {
		t.Errorf("a refused key was classified as %v", err)
	}

	unreachable := &OperationError{
		Operation: "ls-remote",
		Detail:    "ssh: Could not resolve hostname example.org",
	}
	if err := classifyRemote(unreachable); !errors.Is(err, ErrRemoteUnreachable) {
		t.Errorf("an unreachable remote was classified as %v", err)
	}

	// Anything that is not an operation failure is left as it was.
	plain := errors.New("something else entirely")
	if err := classifyRemote(plain); err != plain {
		t.Errorf("classifyRemote changed an unrelated error into %v", err)
	}
}

func TestSitesUseTheKeyTheLookupGives(t *testing.T) {
	cfg := testConfig(t)
	root := filepath.Join(t.TempDir(), "sites")

	asked := []string{}
	sites := NewSites(cfg, root, func(_ context.Context, id string) (string, error) {
		asked = append(asked, id)
		return "", nil
	})

	repository, err := sites.Clone(t.Context(), "korenaar", bareRemote(t, cfg))
	if err != nil {
		t.Fatal(err)
	}
	repository.Close()

	if len(asked) == 0 || asked[0] != "korenaar" {
		t.Errorf("the lookup was asked for %v, want the site's own identifier", asked)
	}
}

func TestSitesReportALookupFailure(t *testing.T) {
	cfg := testConfig(t)
	wanted := errors.New("this project has no key yet")

	sites := NewSites(cfg, filepath.Join(t.TempDir(), "sites"),
		func(context.Context, string) (string, error) { return "", wanted })

	if _, err := sites.Clone(t.Context(), "korenaar", "somewhere"); !errors.Is(err, wanted) {
		t.Errorf("Clone() error = %v, want the lookup's own error", err)
	}
	if _, err := sites.Open(t.Context(), "korenaar"); err == nil {
		t.Error("Open() error = nil, want a failure")
	}
	if err := sites.Reachable(t.Context(), "korenaar", "somewhere"); !errors.Is(err, wanted) {
		t.Errorf("Reachable() error = %v, want the lookup's own error", err)
	}
}

func testIdentity() Identity {
	return Identity{Name: "Quiqr ChatCMS", Email: "chatcms@example.org"}
}
