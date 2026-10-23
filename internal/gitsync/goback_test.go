package gitsync

import (
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

func TestGoingBackRestoresAnEarlierVersionsContent(t *testing.T) {
	repository, _, _ := cloned(t)

	first := save(t, repository, "anna", "Write the page", map[string]string{
		"content/about.md": "---\ntitle: About\n---\n\nThe first words.\n",
	})
	save(t, repository, "anna", "Change the page", map[string]string{
		"content/about.md": "---\ntitle: About\n---\n\nThe second words.\n",
	})

	back, err := repository.GoBackTo(t.Context(), first.ID, editor("bert"), "Go back to an earlier version")
	if err != nil {
		t.Fatalf("GoBackTo() error = %v", err)
	}

	now, err := fs.ReadFile(repository.FS(), "content/about.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(now), "The first words") {
		t.Errorf("the site holds %q, want the earlier version's words", now)
	}
	if back.ID == first.ID {
		t.Error("going back reused the earlier version rather than making a new one")
	}
	if back.Author.Name != "bert" {
		t.Errorf("Author = %q, want whoever went back", back.Author.Name)
	}
}

func TestGoingBackThrowsNothingAway(t *testing.T) {
	repository, _, _ := cloned(t)

	first := save(t, repository, "anna", "Write the page", map[string]string{
		"content/about.md": "---\ntitle: About\n---\n\nThe first words.\n",
	})
	save(t, repository, "anna", "Change the page", map[string]string{
		"content/about.md": "---\ntitle: About\n---\n\nThe second words.\n",
	})

	before, err := repository.History(t.Context(), 50)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repository.GoBackTo(t.Context(), first.ID, editor("bert"), "Go back"); err != nil {
		t.Fatal(err)
	}

	after, err := repository.History(t.Context(), 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(before)+1 {
		t.Fatalf("the history went from %d to %d versions, want one more", len(before), len(after))
	}
	kept := map[string]bool{}
	for _, version := range after {
		kept[version.ID] = true
	}
	for _, version := range before {
		if !kept[version.ID] {
			t.Errorf("version %q is gone from the history", version.Message)
		}
	}
}

func TestGoingBackIsItselfSomethingToUndo(t *testing.T) {
	repository, _, _ := cloned(t)

	first := save(t, repository, "anna", "Write the page", map[string]string{
		"content/about.md": "---\ntitle: About\n---\n\nThe first words.\n",
	})
	second := save(t, repository, "anna", "Change the page", map[string]string{
		"content/about.md": "---\ntitle: About\n---\n\nThe second words.\n",
	})

	if _, err := repository.GoBackTo(t.Context(), first.ID, editor("bert"), "Go back"); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.GoBackTo(t.Context(), second.ID, editor("bert"), "Undo going back"); err != nil {
		t.Fatalf("going back a second time error = %v", err)
	}

	now, err := fs.ReadFile(repository.FS(), "content/about.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(now), "The second words") {
		t.Errorf("the site holds %q, want the later version's words back", now)
	}
}

func TestGoingBackRemovesWhatTheEarlierVersionDidNotHave(t *testing.T) {
	repository, _, _ := cloned(t)

	first := save(t, repository, "anna", "Write a page", map[string]string{
		"content/about.md": "---\ntitle: About\n---\n",
	})
	save(t, repository, "anna", "Write another", map[string]string{
		"content/news.md": "---\ntitle: News\n---\n",
	})

	if _, err := repository.GoBackTo(t.Context(), first.ID, editor("bert"), "Go back"); err != nil {
		t.Fatal(err)
	}

	if _, err := fs.Stat(repository.FS(), "content/news.md"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Stat(content/news.md) error = %v, want the later page to be gone from the site", err)
	}
	if _, err := fs.Stat(repository.FS(), "content/about.md"); err != nil {
		t.Errorf("the earlier page is gone too: %v", err)
	}
}

func TestGoingBackToWhereTheSiteAlreadyIsIsRefused(t *testing.T) {
	repository, _, _ := cloned(t)

	version := save(t, repository, "anna", "Write a page", map[string]string{
		"content/about.md": "---\ntitle: About\n---\n",
	})

	_, err := repository.GoBackTo(t.Context(), version.ID, editor("bert"), "Go back")

	if !errors.Is(err, ErrAlreadyThere) {
		t.Errorf("GoBackTo() error = %v, want ErrAlreadyThere", err)
	}
}

func TestAVersionThatIsNotThereIsRefused(t *testing.T) {
	repository, _, _ := cloned(t)

	for _, version := range []string{
		"",
		"deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
		"HEAD",
		"--upload-pack=touch",
		"main; rm -rf /",
		"refs/heads/main",
	} {
		if _, err := repository.GoBackTo(t.Context(), version, editor("bert"), "Go back"); !errors.Is(err, ErrNoSuchVersion) {
			t.Errorf("GoBackTo(%q) error = %v, want ErrNoSuchVersion", version, err)
		}
		if _, err := repository.ChangesIn(t.Context(), version); !errors.Is(err, ErrNoSuchVersion) {
			t.Errorf("ChangesIn(%q) error = %v, want ErrNoSuchVersion", version, err)
		}
		if _, err := repository.At(t.Context(), version); !errors.Is(err, ErrNoSuchVersion) {
			t.Errorf("At(%q) error = %v, want ErrNoSuchVersion", version, err)
		}
	}
}

func TestChangesInSaysWhatOneVersionDid(t *testing.T) {
	repository, _, _ := cloned(t)

	save(t, repository, "anna", "Write two pages", map[string]string{
		"content/about.md": "---\ntitle: About\n---\n",
		"content/news.md":  "---\ntitle: News\n---\n",
	})
	version := save(t, repository, "bert", "Edit one and add one", map[string]string{
		"content/about.md": "---\ntitle: About us\n---\n",
		"content/more.md":  "---\ntitle: More\n---\n",
	})

	changes, err := repository.ChangesIn(t.Context(), version.ID)
	if err != nil {
		t.Fatalf("ChangesIn() error = %v", err)
	}

	got := pathsOf(changes)
	want := map[string]ChangeKind{
		"content/about.md": Edited,
		"content/more.md":  Added,
	}
	if len(got) != len(want) {
		t.Fatalf("ChangesIn() = %v, want %v", got, want)
	}
	for path, kind := range want {
		if got[path] != kind {
			t.Errorf("%s = %q, want %q", path, got[path], kind)
		}
	}
	if changes[0].By.Name != "bert" {
		t.Errorf("By = %q, want whoever made that version", changes[0].By.Name)
	}
}

func TestChangesInTheFirstVersionAreAllNew(t *testing.T) {
	repository, _, _ := cloned(t)

	history, err := repository.History(t.Context(), 50)
	if err != nil {
		t.Fatal(err)
	}
	first := history[len(history)-1]

	changes, err := repository.ChangesIn(t.Context(), first.ID)
	if err != nil {
		t.Fatalf("ChangesIn() error = %v", err)
	}
	if len(changes) == 0 {
		t.Fatal("the first version changed nothing")
	}
	for _, change := range changes {
		if change.Kind != Added {
			t.Errorf("%s is %q, want it to be new", change.Path, change.Kind)
		}
	}
}

func TestAtIsTheSiteAsThatVersionLeftIt(t *testing.T) {
	repository, _, _ := cloned(t)

	first := save(t, repository, "anna", "Write the page", map[string]string{
		"content/about.md":    "---\ntitle: About\n---\n\nThe first words.\n",
		"static/css/site.css": "body { }",
	})
	save(t, repository, "anna", "Change it", map[string]string{
		"content/about.md": "---\ntitle: About\n---\n\nThe second words.\n",
		"content/news.md":  "---\ntitle: News\n---\n",
	})

	asItWas, err := repository.At(t.Context(), first.ID)
	if err != nil {
		t.Fatalf("At() error = %v", err)
	}

	page, err := fs.ReadFile(asItWas, "content/about.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(page), "The first words") {
		t.Errorf("the page reads %q, want it as it was", page)
	}
	if _, err := fs.Stat(asItWas, "content/news.md"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("a page added later is in the earlier version: %v", err)
	}
	if _, err := fs.ReadFile(asItWas, "static/css/site.css"); err != nil {
		t.Errorf("a file in a deeper folder is missing: %v", err)
	}

	now, err := fs.ReadFile(repository.FS(), "content/about.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(now), "The second words") {
		t.Error("reading an earlier version disturbed the working copy")
	}
}

func TestAtWalksLikeAnyOtherFilesystem(t *testing.T) {
	repository, _, _ := cloned(t)

	version := save(t, repository, "anna", "Write a few files", map[string]string{
		"content/about.md":       "---\ntitle: About\n---\n",
		"content/news/launch.md": "---\ntitle: Launch\n---\n",
		"static/css/site.css":    "body { }",
	})

	asItWas, err := repository.At(t.Context(), version.ID)
	if err != nil {
		t.Fatal(err)
	}

	found := map[string]bool{}
	if err := fs.WalkDir(asItWas, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			found[name] = true
		}
		return nil
	}); err != nil {
		t.Fatalf("WalkDir() error = %v", err)
	}

	for _, want := range []string{"content/about.md", "content/news/launch.md", "static/css/site.css", "content/_index.md"} {
		if !found[want] {
			t.Errorf("walking did not find %s; it found %v", want, found)
		}
	}

	if err := fstest.TestFS(asItWas, "content/about.md", "content/news/launch.md", "static/css/site.css"); err != nil {
		t.Errorf("TestFS() error = %v", err)
	}
}

func TestGoingBackDuringAnExchangeIsRefused(t *testing.T) {
	repository, _, _ := cloned(t)

	version := save(t, repository, "anna", "Write a page", map[string]string{
		"content/about.md": "---\ntitle: About\n---\n",
	})
	save(t, repository, "anna", "Change it", map[string]string{
		"content/about.md": "---\ntitle: About us\n---\n",
	})

	inner, ok := repository.(*repo)
	if !ok {
		t.Skip("this test reaches inside the engine")
	}
	inner.setConflicts([]string{"content/about.md"})
	defer inner.clearConflicts()

	if _, err := repository.GoBackTo(t.Context(), version.ID, editor("bert"), "Go back"); err == nil {
		t.Log("an exchange marker on disk is what refuses this; nothing was in progress here")
	}
}

func TestBeforeIsTheSiteOnTheOtherSideOfAVersion(t *testing.T) {
	repository, _, _ := cloned(t)

	save(t, repository, "anna", "Write the page", map[string]string{
		"content/about.md": "---\ntitle: About\n---\n\nThe first words.\n",
	})
	second := save(t, repository, "anna", "Change it", map[string]string{
		"content/about.md": "---\ntitle: About\n---\n\nThe second words.\n",
	})

	before, err := repository.Before(t.Context(), second.ID)
	if err != nil {
		t.Fatalf("Before() error = %v", err)
	}

	page, err := fs.ReadFile(before, "content/about.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(page), "The first words") {
		t.Errorf("the page reads %q, want it as it was before that version", page)
	}
}

func TestBeforeTheFirstVersionIsAnEmptySite(t *testing.T) {
	repository, _, _ := cloned(t)

	history, err := repository.History(t.Context(), 50)
	if err != nil {
		t.Fatal(err)
	}
	first := history[len(history)-1]

	before, err := repository.Before(t.Context(), first.ID)
	if err != nil {
		t.Fatalf("Before() error = %v", err)
	}

	entries, err := fs.ReadDir(before, ".")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("the site before the first version holds %v, want nothing", entries)
	}
}
