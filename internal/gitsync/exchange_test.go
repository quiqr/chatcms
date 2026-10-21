package gitsync

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncBringsRemoteVersions(t *testing.T) {
	repository, remote, cfg := cloned(t)
	other := elsewhere(t, cfg, remote)
	pushFrom(t, cfg, other, "content/news.md", "---\ntitle: News\n---\n", "Add news")

	result, err := repository.Sync(t.Context())
	if err != nil {
		t.Fatalf("Sync() error = %v, want nil", err)
	}
	if len(result.Arrived) != 1 || result.Arrived[0].Message != "Add news" {
		t.Fatalf("arrived = %+v, want the one pushed version", result.Arrived)
	}

	contents, err := fs.ReadFile(repository.FS(), "content/news.md")
	if err != nil {
		t.Fatalf("the arrived content is not readable: %v", err)
	}
	if !strings.Contains(string(contents), "title: News") {
		t.Errorf("content = %q", contents)
	}

	status, _ := repository.Status(t.Context())
	if status.RemoteAhead != 0 || status.Unpublished != 0 {
		t.Errorf("status = %+v, want level with the remote", status)
	}
}

func TestSyncReportsNothingNew(t *testing.T) {
	repository, _, _ := cloned(t)

	result, err := repository.Sync(t.Context())
	if err != nil {
		t.Fatalf("Sync() error = %v, want nil", err)
	}
	if len(result.Arrived) != 0 || result.Conflict != nil {
		t.Errorf("result = %+v, want nothing to have arrived", result)
	}
}

func TestSyncSurvivesAnUnreachableRemote(t *testing.T) {
	repository, remote, _ := cloned(t)
	if err := os.Rename(remote, remote+".gone"); err != nil {
		t.Fatal(err)
	}

	_, err := repository.Sync(t.Context())
	if !errors.Is(err, ErrRemoteUnreachable) {
		t.Fatalf("Sync() error = %v, want ErrRemoteUnreachable", err)
	}
	if strings.Contains(err.Error(), "password") || strings.Contains(err.Error(), "token") {
		t.Errorf("error = %v, want no credential in the message", err)
	}

	if _, err := fs.ReadFile(repository.FS(), "content/_index.md"); err != nil {
		t.Errorf("the site is unusable after a failed sync: %v", err)
	}
}

func TestReplayKeepsHistoryLinear(t *testing.T) {
	repository, remote, cfg := cloned(t)

	save(t, repository, "joris", "Local one", map[string]string{"content/local1.md": "1"})
	save(t, repository, "joris", "Local two", map[string]string{"content/local2.md": "2"})

	other := elsewhere(t, cfg, remote)
	pushFrom(t, cfg, other, "content/remote.md", "remote", "Remote one")

	result, err := repository.Sync(t.Context())
	if err != nil {
		t.Fatalf("Sync() error = %v, want nil", err)
	}
	if result.Conflict != nil {
		t.Fatalf("unexpected conflict: %+v", result.Conflict)
	}
	if result.Replayed != 2 {
		t.Errorf("replayed = %d, want 2", result.Replayed)
	}

	versions, _ := repository.History(t.Context(), 4)
	messages := []string{}
	for _, version := range versions {
		messages = append(messages, version.Message)
	}
	want := []string{"Local two", "Local one", "Remote one", "First version"}
	for i, message := range want {
		if i >= len(messages) || messages[i] != message {
			t.Fatalf("history = %v, want %v", messages, want)
		}
	}

	for _, message := range messages {
		if strings.HasPrefix(strings.ToLower(message), "merge") {
			t.Errorf("history contains a merge version: %q", message)
		}
	}
}

func TestPublishSendsVersions(t *testing.T) {
	repository, remote, cfg := cloned(t)
	save(t, repository, "joris", "Add a page", map[string]string{"content/a.md": "a"})

	result, err := repository.Publish(t.Context())
	if err != nil {
		t.Fatalf("Publish() error = %v, want nil", err)
	}
	if len(result.Published) != 1 || result.Published[0].Message != "Add a page" {
		t.Errorf("published = %+v, want the one local version", result.Published)
	}

	check := elsewhere(t, cfg, remote)
	contents, err := os.ReadFile(filepath.Join(check, "content", "a.md"))
	if err != nil {
		t.Fatalf("the version did not reach the remote: %v", err)
	}
	if string(contents) != "a" {
		t.Errorf("remote content = %q, want a", contents)
	}

	status, _ := repository.Status(t.Context())
	if status.Unpublished != 0 {
		t.Errorf("unpublished = %d after publishing, want 0", status.Unpublished)
	}
}

func TestRejectedPushLosesNothing(t *testing.T) {
	repository, remote, cfg := cloned(t)
	save(t, repository, "joris", "Local edit", map[string]string{"content/local.md": "local"})

	other := elsewhere(t, cfg, remote)
	pushFrom(t, cfg, other, "content/remote.md", "remote", "Remote edit")

	_, err := repository.Publish(t.Context())
	if !errors.Is(err, ErrRemoteMoved) {
		t.Fatalf("Publish() error = %v, want ErrRemoteMoved", err)
	}

	versions, _ := repository.History(t.Context(), 10)
	if versions[0].Message != "Local edit" {
		t.Fatalf("the local version is gone: %+v", versions)
	}

	if _, err := repository.Sync(t.Context()); err != nil {
		t.Fatalf("Sync() after a rejected publish failed: %v", err)
	}
	if _, err := repository.Publish(t.Context()); err != nil {
		t.Fatalf("Publish() after syncing failed: %v", err)
	}

	check := elsewhere(t, cfg, remote)
	for _, name := range []string{"content/local.md", "content/remote.md"} {
		if _, err := os.ReadFile(filepath.Join(check, filepath.FromSlash(name))); err != nil {
			t.Errorf("%s is not on the remote: %v", name, err)
		}
	}
}

// conflicting sets up one file changed differently on both sides.
func conflicting(t *testing.T) (Repo, string, Config) {
	t.Helper()

	repository, remote, cfg := cloned(t)
	other := elsewhere(t, cfg, remote)
	pushFrom(t, cfg, other, "content/_index.md", "---\ntitle: Remote title\n---\n", "Remote edit")
	save(t, repository, "joris", "Local edit", map[string]string{
		"content/_index.md": "---\ntitle: Local title\n---\n",
	})

	return repository, remote, cfg
}

func TestConflictCarriesThreeVersions(t *testing.T) {
	repository, _, _ := conflicting(t)

	result, err := repository.Sync(t.Context())
	if err != nil {
		t.Fatalf("Sync() error = %v, want a reported conflict rather than a failure", err)
	}
	if result.Conflict == nil {
		t.Fatal("Sync() reported no conflict")
	}
	if len(result.Conflict.Paths) != 1 || result.Conflict.Paths[0].Path != "content/_index.md" {
		t.Fatalf("conflict = %+v, want the one path", result.Conflict.Paths)
	}

	path := result.Conflict.Paths[0]
	if !strings.Contains(string(path.Base), "title: Home") {
		t.Errorf("base = %q, want the content both sides started from", path.Base)
	}
	if !strings.Contains(string(path.Local), "Local title") {
		t.Errorf("local = %q, want the version made here", path.Local)
	}
	if !strings.Contains(string(path.Remote), "Remote title") {
		t.Errorf("remote = %q, want the version made elsewhere", path.Remote)
	}
}

func TestNoConflictMarkerReachesContent(t *testing.T) {
	repository, _, _ := conflicting(t)

	result, err := repository.Sync(t.Context())
	if err != nil || result.Conflict == nil {
		t.Fatalf("Sync() = %+v, %v, want a conflict", result, err)
	}

	// The file on disk must not hold markers, whatever reads it.
	dir := repository.(*repo).dir
	raw, err := os.ReadFile(filepath.Join(dir, "content", "_index.md"))
	if err != nil {
		t.Fatalf("reading the file on disk failed: %v", err)
	}
	for _, marker := range []string{"<<<<<<<", "=======", ">>>>>>>"} {
		if strings.Contains(string(raw), marker) {
			t.Fatalf("the file on disk holds a conflict marker:\n%s", raw)
		}
	}

	// Through the content filesystem the path is refused rather than served.
	_, err = fs.ReadFile(repository.FS(), "content/_index.md")
	if !errors.Is(err, ErrPathInConflict) {
		t.Errorf("reading a conflicted path gave %v, want ErrPathInConflict", err)
	}

	// Everything else still reads, and nothing readable holds a marker.
	err = fs.WalkDir(repository.FS(), ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return nil
		}
		contents, readErr := fs.ReadFile(repository.FS(), path)
		if readErr != nil {
			return nil
		}
		if strings.Contains(string(contents), "<<<<<<<") {
			t.Errorf("%s holds a conflict marker", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDifferentFilesDoNotConflict(t *testing.T) {
	repository, remote, cfg := cloned(t)
	other := elsewhere(t, cfg, remote)
	pushFrom(t, cfg, other, "content/remote.md", "remote", "Remote edit")
	save(t, repository, "joris", "Local edit", map[string]string{"content/local.md": "local"})

	result, err := repository.Sync(t.Context())
	if err != nil {
		t.Fatalf("Sync() error = %v, want nil", err)
	}
	if result.Conflict != nil {
		t.Fatalf("unexpected conflict: %+v", result.Conflict)
	}

	for _, name := range []string{"content/local.md", "content/remote.md"} {
		if _, err := fs.ReadFile(repository.FS(), name); err != nil {
			t.Errorf("%s is missing after the sync: %v", name, err)
		}
	}
}

func resolveWith(t *testing.T, repository Repo, chooser func(ConflictedPath) []byte) SyncResult {
	t.Helper()

	result, err := repository.Sync(t.Context())
	if err != nil || result.Conflict == nil {
		t.Fatalf("Sync() = %+v, %v, want a conflict", result, err)
	}

	resolutions := map[string][]byte{}
	for _, path := range result.Conflict.Paths {
		resolutions[path.Path] = chooser(path)
	}

	resolved, err := repository.Resolve(t.Context(), resolutions)
	if err != nil {
		t.Fatalf("Resolve() error = %v, want nil", err)
	}
	if resolved.Conflict != nil {
		t.Fatalf("still conflicted after resolving: %+v", resolved.Conflict)
	}
	return resolved
}

func TestResolveKeepingLocal(t *testing.T) {
	repository, _, _ := conflicting(t)

	resolveWith(t, repository, func(path ConflictedPath) []byte { return path.Local })

	contents, err := fs.ReadFile(repository.FS(), "content/_index.md")
	if err != nil {
		t.Fatalf("reading after resolving failed: %v", err)
	}
	if !strings.Contains(string(contents), "Local title") {
		t.Errorf("content = %q, want the local version", contents)
	}

	status, _ := repository.Status(t.Context())
	if status.ExchangeInProgress {
		t.Error("an exchange is still in progress after resolving")
	}
}

func TestResolveKeepingRemote(t *testing.T) {
	repository, _, _ := conflicting(t)

	resolveWith(t, repository, func(path ConflictedPath) []byte { return path.Remote })

	contents, _ := fs.ReadFile(repository.FS(), "content/_index.md")
	if !strings.Contains(string(contents), "Remote title") {
		t.Errorf("content = %q, want the remote version", contents)
	}
}

func TestResolveWithNewContent(t *testing.T) {
	repository, _, _ := conflicting(t)

	resolveWith(t, repository, func(ConflictedPath) []byte {
		return []byte("---\ntitle: Something both of us can live with\n---\n")
	})

	contents, _ := fs.ReadFile(repository.FS(), "content/_index.md")
	if !strings.Contains(string(contents), "both of us") {
		t.Errorf("content = %q, want the content that was chosen", contents)
	}
}

func TestResolveRejectsAnUnconflictedPath(t *testing.T) {
	repository, _, _ := conflicting(t)

	result, err := repository.Sync(t.Context())
	if err != nil || result.Conflict == nil {
		t.Fatalf("Sync() = %+v, %v, want a conflict", result, err)
	}

	_, err = repository.Resolve(t.Context(), map[string][]byte{
		"content/not-conflicted.md": []byte("nope"),
	})
	if !errors.Is(err, ErrNotConflicted) {
		t.Fatalf("Resolve() error = %v, want ErrNotConflicted", err)
	}

	status, _ := repository.Status(t.Context())
	if !status.ExchangeInProgress {
		t.Error("the exchange was ended by a rejected resolution")
	}
}

func TestAbandonRestoresTheWorkingCopy(t *testing.T) {
	repository, _, _ := conflicting(t)

	before, _ := repository.History(t.Context(), 10)

	result, err := repository.Sync(t.Context())
	if err != nil || result.Conflict == nil {
		t.Fatalf("Sync() = %+v, %v, want a conflict", result, err)
	}

	if err := repository.Abandon(t.Context()); err != nil {
		t.Fatalf("Abandon() error = %v, want nil", err)
	}

	after, _ := repository.History(t.Context(), 10)
	if len(after) != len(before) || after[0].Message != before[0].Message {
		t.Errorf("history after abandoning = %+v, want it as it was", after)
	}

	contents, err := fs.ReadFile(repository.FS(), "content/_index.md")
	if err != nil {
		t.Fatalf("reading after abandoning failed: %v", err)
	}
	if !strings.Contains(string(contents), "Local title") {
		t.Errorf("content = %q, want the local version back", contents)
	}

	status, _ := repository.Status(t.Context())
	if status.ExchangeInProgress {
		t.Error("an exchange is still in progress after abandoning")
	}
}

func TestAbandonWithoutAnExchange(t *testing.T) {
	repository, _, _ := cloned(t)

	if err := repository.Abandon(t.Context()); !errors.Is(err, ErrNoExchange) {
		t.Errorf("Abandon() error = %v, want ErrNoExchange", err)
	}
	if _, err := repository.Resolve(t.Context(), nil); !errors.Is(err, ErrNoExchange) {
		t.Errorf("Resolve() error = %v, want ErrNoExchange", err)
	}
}

func TestReplayDoesNotAskTwiceForTheSameConflict(t *testing.T) {
	repository, remote, cfg := cloned(t)

	// Two local versions both touching the file the remote also changed, so
	// the replay would otherwise stop twice on the same path.
	save(t, repository, "joris", "Local one", map[string]string{
		"content/_index.md": "---\ntitle: Local one\n---\n",
	})
	save(t, repository, "joris", "Local two", map[string]string{
		"content/_index.md": "---\ntitle: Local two\n---\n",
	})

	other := elsewhere(t, cfg, remote)
	pushFrom(t, cfg, other, "content/_index.md", "---\ntitle: Remote\n---\n", "Remote edit")

	result, err := repository.Sync(t.Context())
	if err != nil || result.Conflict == nil {
		t.Fatalf("Sync() = %+v, %v, want a conflict", result, err)
	}

	chosen := []byte("---\ntitle: Agreed\n---\n")
	resolved, err := repository.Resolve(t.Context(), map[string][]byte{
		"content/_index.md": chosen,
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v, want nil", err)
	}
	if resolved.Conflict != nil {
		t.Fatalf("asked again about the same file: %+v", resolved.Conflict)
	}

	contents, _ := fs.ReadFile(repository.FS(), "content/_index.md")
	if string(contents) != string(chosen) {
		t.Errorf("content = %q, want the answer that was given", contents)
	}
}

func TestOpenReportsAnExchangeInProgress(t *testing.T) {
	repository, _, cfg := conflicting(t)
	dir := repository.(*repo).dir

	result, err := repository.Sync(t.Context())
	if err != nil || result.Conflict == nil {
		t.Fatalf("Sync() = %+v, %v, want a conflict", result, err)
	}

	// Throw the handle away, as a process that died would.
	repository.Close()

	reopened, err := Open(t.Context(), cfg, dir)
	if err != nil {
		t.Fatalf("Open() error = %v, want nil", err)
	}
	defer reopened.Close()

	status, err := reopened.Status(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if !status.ExchangeInProgress {
		t.Fatal("the reopened working copy does not report the unfinished exchange")
	}

	if _, err := fs.ReadFile(reopened.FS(), "content/_index.md"); !errors.Is(err, ErrPathInConflict) {
		t.Errorf("reading the conflicted path gave %v, want ErrPathInConflict", err)
	}

	if err := reopened.Abandon(t.Context()); err != nil {
		t.Fatalf("Abandon() after reopening failed: %v", err)
	}

	versions, _ := reopened.History(t.Context(), 10)
	if versions[0].Message != "Local edit" {
		t.Errorf("history after abandoning = %+v, want the local version back", versions)
	}
}

func TestSyncRefusesWhileAnExchangeIsInProgress(t *testing.T) {
	repository, _, _ := conflicting(t)

	if result, err := repository.Sync(t.Context()); err != nil || result.Conflict == nil {
		t.Fatalf("Sync() = %+v, %v, want a conflict", result, err)
	}
	if _, err := repository.Sync(t.Context()); !errors.Is(err, ErrExchangeInProgress) {
		t.Errorf("Sync() error = %v, want ErrExchangeInProgress", err)
	}
	if _, err := repository.Publish(t.Context()); !errors.Is(err, ErrExchangeInProgress) {
		t.Errorf("Publish() error = %v, want ErrExchangeInProgress", err)
	}
}

func TestStatusAfterSave(t *testing.T) {
	repository, _, _ := cloned(t)

	save(t, repository, "joris", "An edit", map[string]string{"content/a.md": "a"})

	status, err := repository.Status(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if status.UnsavedWork {
		t.Error("status reports unsaved work right after a save")
	}
}

func TestStatusSeesUnsavedWork(t *testing.T) {
	repository, _, _ := cloned(t)
	dir := repository.(*repo).dir
	writeFile(t, filepath.Join(dir, "content", "stray.md"), "written behind our back")

	status, _ := repository.Status(t.Context())
	if !status.UnsavedWork {
		t.Error("status does not notice a file that was written but not saved")
	}
}

func TestStatusCountsUnpublished(t *testing.T) {
	repository, _, _ := cloned(t)
	for i, name := range []string{"content/a.md", "content/b.md", "content/c.md"} {
		save(t, repository, "joris", "Edit", map[string]string{name: strings.Repeat("x", i+1)})
	}

	status, _ := repository.Status(t.Context())
	if status.Unpublished != 3 {
		t.Errorf("unpublished = %d, want 3", status.Unpublished)
	}
}

func TestStatusSeesTheRemoteMove(t *testing.T) {
	repository, remote, cfg := cloned(t)
	other := elsewhere(t, cfg, remote)
	pushFrom(t, cfg, other, "content/news.md", "news", "Remote edit")

	if _, err := repository.(*repo).git.run(t.Context(), "fetch", "origin"); err != nil {
		t.Fatal(err)
	}

	status, _ := repository.Status(t.Context())
	if status.RemoteAhead != 1 {
		t.Errorf("remoteAhead = %d, want 1", status.RemoteAhead)
	}
}

func TestStatusIsNotRemembered(t *testing.T) {
	repository, _, cfg := cloned(t)
	save(t, repository, "joris", "An edit", map[string]string{"content/a.md": "a"})

	before, _ := repository.Status(t.Context())

	// Throw away everything the engine holds in memory and read it again.
	dir := repository.(*repo).dir
	repository.Close()

	reopened, err := Open(t.Context(), cfg, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()

	after, _ := reopened.Status(t.Context())
	if before != after {
		t.Errorf("status changed from %+v to %+v with nothing but a reopen between", before, after)
	}
}
