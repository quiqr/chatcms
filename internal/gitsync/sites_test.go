package gitsync

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSiteIDMustBeUsableAsADirectory(t *testing.T) {
	sites := NewSites(testConfig(t), t.TempDir())

	for _, id := range []string{"", ".", "..", "a/b", `a\b`, ".hidden", "a:b"} {
		if _, err := sites.Dir(id); !errors.Is(err, ErrUnusableSiteID) {
			t.Errorf("Dir(%q) error = %v, want ErrUnusableSiteID", id, err)
		}
	}

	dir, err := sites.Dir("korenaar")
	if err != nil {
		t.Fatalf("Dir() error = %v, want nil", err)
	}
	if filepath.Base(filepath.Dir(dir)) != "korenaar" {
		t.Errorf("Dir() = %q, want it under the site's own directory", dir)
	}
}

func TestSitesCloneAndOpen(t *testing.T) {
	cfg := testConfig(t)
	remote := bareRemote(t, cfg)
	sites := NewSites(cfg, filepath.Join(t.TempDir(), "sites"))

	repository, err := sites.Clone(t.Context(), "korenaar", remote)
	if err != nil {
		t.Fatalf("Clone() error = %v, want nil", err)
	}
	repository.Close()

	reopened, err := sites.Open(t.Context(), "korenaar")
	if err != nil {
		t.Fatalf("Open() error = %v, want nil", err)
	}
	defer reopened.Close()

	if _, err := reopened.History(t.Context(), 1); err != nil {
		t.Errorf("History() error = %v, want nil", err)
	}
}

func TestSitesListsWorkingCopies(t *testing.T) {
	cfg := testConfig(t)
	root := filepath.Join(t.TempDir(), "sites")
	sites := NewSites(cfg, root)

	if ids, err := sites.List(); err != nil || len(ids) != 0 {
		t.Fatalf("List() = %v, %v, want nothing before any site exists", ids, err)
	}

	for _, id := range []string{"studio-noord", "korenaar"} {
		repository, err := sites.Clone(t.Context(), id, bareRemote(t, cfg))
		if err != nil {
			t.Fatal(err)
		}
		repository.Close()
	}

	ids, err := sites.List()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ids, []string{"korenaar", "studio-noord"}) {
		t.Errorf("List() = %v, want both sites in a stable order", ids)
	}
}
