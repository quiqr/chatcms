package gitsync

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func pathsOf(changes []Change) map[string]ChangeKind {
	out := map[string]ChangeKind{}
	for _, change := range changes {
		out[change.Path] = change.Kind
	}
	return out
}

func TestChangesSaysWhatDiffersFromThePublishedSite(t *testing.T) {
	repository, _, _ := cloned(t)

	save(t, repository, "anna", "Edit the home page", map[string]string{
		"content/_index.md": "---\ntitle: Home\n---\n\nNow with words.\n",
		"content/about.md":  "---\ntitle: About\n---\n",
		"static/logo.svg":   "<svg/>",
	})
	if _, err := repository.Save(t.Context(), Save{
		Message: "Remove the old page",
		Author:  editor("bert"),
		Deletes: []string{"content/about.md"},
	}); err != nil {
		t.Fatal(err)
	}

	changes, err := repository.Changes(t.Context())
	if err != nil {
		t.Fatalf("Changes() error = %v", err)
	}

	got := pathsOf(changes)
	want := map[string]ChangeKind{
		"content/_index.md": Edited,
		"static/logo.svg":   Added,
	}
	if len(got) != len(want) {
		t.Fatalf("Changes() = %v, want %v", got, want)
	}
	for path, kind := range want {
		if got[path] != kind {
			t.Errorf("%s = %q, want %q", path, got[path], kind)
		}
	}
}

func TestAChangeSaysWhoMadeItAndWhen(t *testing.T) {
	repository, _, _ := cloned(t)

	save(t, repository, "anna", "Edit the home page", map[string]string{
		"content/_index.md": "---\ntitle: Home\n---\n\nAnna wrote this.\n",
	})

	changes, err := repository.Changes(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 {
		t.Fatalf("Changes() = %v, want one", changes)
	}

	if changes[0].By.Name != "anna" {
		t.Errorf("By.Name = %q, want anna", changes[0].By.Name)
	}
	if changes[0].At.IsZero() {
		t.Error("At is not set")
	}
}

func TestAChangeThatRemovesAPageIsStillAttributed(t *testing.T) {
	repository, _, _ := cloned(t)

	if _, err := repository.Save(t.Context(), Save{
		Message: "Take the home page down",
		Author:  editor("bert"),
		Deletes: []string{"content/_index.md"},
	}); err != nil {
		t.Fatal(err)
	}

	changes, err := repository.Changes(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 || changes[0].Kind != Removed {
		t.Fatalf("Changes() = %v, want one removal", changes)
	}
	if changes[0].By.Name != "bert" {
		t.Errorf("By.Name = %q, want the person who removed it", changes[0].By.Name)
	}
}

func TestASiteThatWasNeverPublishedHasEverythingAsNew(t *testing.T) {
	cfg := testConfig(t)
	remote := emptyBareRemote(t, cfg)
	sites := NewSites(cfg, filepath.Join(t.TempDir(), "sites"), nil)

	repository, err := sites.Create(t.Context(), "korenaar", remote, fstest.MapFS{
		"hugo.toml":         {Data: []byte("title = \"A Site\"\n")},
		"content/_index.md": {Data: []byte("---\ntitle: Home\n---\n")},
	}, editor("anna"), "Bring the site in")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { repository.Close() })

	changes, err := repository.Changes(t.Context())
	if err != nil {
		t.Fatalf("Changes() error = %v", err)
	}
	if len(changes) != 2 {
		t.Fatalf("Changes() = %v, want every file as new", changes)
	}
	for _, change := range changes {
		if change.Kind != Added {
			t.Errorf("%s is %q, want it to be new", change.Path, change.Kind)
		}
	}

	if _, err := repository.Publish(t.Context()); err != nil {
		t.Fatal(err)
	}

	changes, err = repository.Changes(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 0 {
		t.Fatalf("Changes() = %v, want nothing once it has been published", changes)
	}

	save(t, repository, "anna", "Add a page", map[string]string{
		"content/news.md": "---\ntitle: News\n---\n",
	})
	changes, err = repository.Changes(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 || changes[0].Kind != Added || changes[0].Path != "content/news.md" {
		t.Errorf("Changes() = %v, want just the new page", changes)
	}
}

func TestNothingChangedMeansNoChanges(t *testing.T) {
	repository, _, _ := cloned(t)

	changes, err := repository.Changes(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 0 {
		t.Errorf("Changes() = %v, want nothing", changes)
	}
}

func TestReadPublishedReadsTheLiveFile(t *testing.T) {
	repository, _, _ := cloned(t)

	save(t, repository, "anna", "Edit the home page", map[string]string{
		"content/_index.md": "---\ntitle: Home\n---\n\nThe new words.\n",
	})

	live, err := repository.ReadPublished(t.Context(), "content/_index.md")
	if err != nil {
		t.Fatalf("ReadPublished() error = %v", err)
	}
	if string(live) != "---\ntitle: Home\n---\n" {
		t.Errorf("ReadPublished() = %q, want the file as the live site has it", live)
	}

	now, err := fs.ReadFile(repository.FS(), "content/_index.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(now), "The new words") {
		t.Errorf("the working copy is %q, want the edited file", now)
	}
}

func TestReadPublishedSaysWhenTheLiveSiteDoesNotHaveIt(t *testing.T) {
	repository, _, _ := cloned(t)

	save(t, repository, "anna", "Add a page", map[string]string{
		"content/news.md": "---\ntitle: News\n---\n",
	})

	_, err := repository.ReadPublished(t.Context(), "content/news.md")
	if !errors.Is(err, ErrNotPublished) {
		t.Errorf("ReadPublished() error = %v, want ErrNotPublished", err)
	}
}

func TestPublishToSendsTheSiteToALineOfItsOwn(t *testing.T) {
	repository, remote, cfg := cloned(t)

	save(t, repository, "anna", "Write a page", map[string]string{
		"content/news.md": "---\ntitle: News\n---\n",
	})

	result, err := repository.PublishTo(t.Context(), "quiqr/publish")
	if err != nil {
		t.Fatalf("PublishTo() error = %v", err)
	}
	if len(result.Published) == 0 {
		t.Error("PublishTo() reported nothing published")
	}

	onTheLine := run(t, cfg, remote, "ls-tree", "-r", "--name-only", "refs/heads/quiqr/publish")
	if !strings.Contains(onTheLine, "content/news.md") {
		t.Errorf("the line holds %q, want the site's own files", onTheLine)
	}

	stillThere := run(t, cfg, remote, "ls-tree", "-r", "--name-only", "refs/heads/main")
	if strings.Contains(stillThere, "content/news.md") {
		t.Error("publishing to a line of its own also changed the main line")
	}
}

func TestPublishToWithNoLineIsTheOrdinaryPublish(t *testing.T) {
	repository, remote, cfg := cloned(t)

	save(t, repository, "anna", "Write a page", map[string]string{
		"content/news.md": "---\ntitle: News\n---\n",
	})

	if _, err := repository.PublishTo(t.Context(), ""); err != nil {
		t.Fatalf("PublishTo() error = %v", err)
	}

	onMain := run(t, cfg, remote, "ls-tree", "-r", "--name-only", "refs/heads/main")
	if !strings.Contains(onMain, "content/news.md") {
		t.Errorf("the main line holds %q, want the saved page", onMain)
	}
}

func TestPublishPagesSendsBuiltPagesWithoutTheSource(t *testing.T) {
	repository, remote, cfg := cloned(t)

	pages := fstest.MapFS{
		"index.html":      {Data: []byte("<h1>Home</h1>")},
		"css/site.css":    {Data: []byte("body { }")},
		"news/index.html": {Data: []byte("<h1>News</h1>")},
	}

	result, err := repository.PublishPages(t.Context(), "live", pages, editor("anna"), "Publish the site")
	if err != nil {
		t.Fatalf("PublishPages() error = %v", err)
	}
	if len(result.Published) != 1 {
		t.Fatalf("PublishPages() published %d versions, want one", len(result.Published))
	}

	listed := run(t, cfg, remote, "ls-tree", "-r", "--name-only", "refs/heads/live")
	for _, want := range []string{"index.html", "css/site.css", "news/index.html"} {
		if !strings.Contains(listed, want) {
			t.Errorf("the published line holds %q, want %s in it", listed, want)
		}
	}
	if strings.Contains(listed, "content/_index.md") {
		t.Error("the site's own files were published alongside the built pages")
	}

	page := run(t, cfg, remote, "cat-file", "blob", "refs/heads/live:index.html")
	if page != "<h1>Home</h1>" {
		t.Errorf("the published page is %q", page)
	}
}

func TestPublishPagesDoesNotDisturbTheWorkingCopy(t *testing.T) {
	repository, _, _ := cloned(t)

	save(t, repository, "anna", "Write a page", map[string]string{
		"content/news.md": "---\ntitle: News\n---\n",
	})
	before, err := repository.Status(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	if _, err := repository.PublishPages(t.Context(), "live",
		fstest.MapFS{"index.html": {Data: []byte("<h1>Home</h1>")}}, editor("anna"), "Publish"); err != nil {
		t.Fatal(err)
	}

	page, err := fs.ReadFile(repository.FS(), "content/news.md")
	if err != nil {
		t.Fatalf("the working copy lost a file: %v", err)
	}
	if !strings.Contains(string(page), "News") {
		t.Errorf("the working copy holds %q", page)
	}

	after, err := repository.Status(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if after.Unpublished != before.Unpublished {
		t.Errorf("Unpublished went from %d to %d; publishing pages made a version of the source",
			before.Unpublished, after.Unpublished)
	}
}

func TestPublishingPagesTwiceContinuesTheSameLine(t *testing.T) {
	repository, remote, cfg := cloned(t)

	first, err := repository.PublishPages(t.Context(), "live",
		fstest.MapFS{"index.html": {Data: []byte("<h1>One</h1>")}}, editor("anna"), "First")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repository.PublishPages(t.Context(), "live",
		fstest.MapFS{"index.html": {Data: []byte("<h1>Two</h1>")}}, editor("anna"), "Second"); err != nil {
		t.Fatal(err)
	}

	history := run(t, cfg, remote, "log", "--format=%s", "refs/heads/live")
	if !strings.Contains(history, "First") || !strings.Contains(history, "Second") {
		t.Errorf("the line's history is %q, want both publishes", history)
	}

	page := run(t, cfg, remote, "cat-file", "blob", "refs/heads/live:index.html")
	if page != "<h1>Two</h1>" {
		t.Errorf("the published page is %q, want the second publish", page)
	}
	if first.Published[0].ID == "" {
		t.Error("the first publish reported no version")
	}
}

func TestAPageThatIsGoneIsGoneFromTheLine(t *testing.T) {
	repository, remote, cfg := cloned(t)

	if _, err := repository.PublishPages(t.Context(), "live", fstest.MapFS{
		"index.html":     {Data: []byte("<h1>Home</h1>")},
		"old/index.html": {Data: []byte("<h1>Old</h1>")},
	}, editor("anna"), "First"); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.PublishPages(t.Context(), "live", fstest.MapFS{
		"index.html": {Data: []byte("<h1>Home</h1>")},
	}, editor("anna"), "Second"); err != nil {
		t.Fatal(err)
	}

	listed := run(t, cfg, remote, "ls-tree", "-r", "--name-only", "refs/heads/live")
	if strings.Contains(listed, "old/index.html") {
		t.Errorf("the line still holds %q; a removed page was not removed", listed)
	}
}

func TestPublishingPagesOverTheSiteIsRefused(t *testing.T) {
	repository, _, _ := cloned(t)

	_, err := repository.PublishPages(t.Context(), "main",
		fstest.MapFS{"index.html": {Data: []byte("<h1>Home</h1>")}}, editor("anna"), "Publish")

	if !errors.Is(err, ErrUnsuitableLine) {
		t.Fatalf("PublishPages() error = %v, want ErrUnsuitableLine", err)
	}
	if !strings.Contains(err.Error(), "replace") {
		t.Errorf("error = %q, want it to say what would happen", err)
	}
}

func TestAnUnusableLineNameIsRefused(t *testing.T) {
	repository, _, _ := cloned(t)

	for _, line := range []string{"", "--upload-pack=touch", "a line", "a/../b"} {
		if _, err := repository.PublishTo(t.Context(), line); line != "" && !errors.Is(err, ErrUnsuitableLine) {
			t.Errorf("PublishTo(%q) error = %v, want ErrUnsuitableLine", line, err)
		}
	}
}

func TestPublishingNoPagesIsRefused(t *testing.T) {
	repository, _, _ := cloned(t)

	if _, err := repository.PublishPages(t.Context(), "live", nil, editor("anna"), "Publish"); err == nil {
		t.Error("PublishPages() with nothing error = nil, want an error")
	}
	if _, err := repository.PublishPages(t.Context(), "live", fstest.MapFS{}, editor("anna"), "Publish"); err == nil {
		t.Error("PublishPages() with an empty site error = nil, want an error")
	}
}

func TestPublishedPagesKeepTheirExecutableBit(t *testing.T) {
	repository, remote, cfg := cloned(t)

	if _, err := repository.PublishPages(t.Context(), "live", fstest.MapFS{
		"index.html":   {Data: []byte("<h1>Home</h1>")},
		"cgi/hello.sh": {Data: []byte("#!/bin/sh\necho hello\n"), Mode: 0o755},
	}, editor("anna"), "Publish"); err != nil {
		t.Fatal(err)
	}

	listed := run(t, cfg, remote, "ls-tree", "-r", "refs/heads/live")
	if !strings.Contains(listed, "100755\tblob") && !strings.Contains(listed, "100755 blob") {
		t.Errorf("the line holds %q, want the script to still be runnable", listed)
	}
}
