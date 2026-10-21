package model

import (
	"reflect"
	"strings"
	"testing"
)

func navigate(t *testing.T, files map[string]string) Navigation {
	t.Helper()

	fsys := site(files)
	m, err := Load(fsys)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	return m.Navigate(fsys)
}

func entryKeys(section NavSection) []string {
	keys := make([]string, 0, len(section.Entries))
	for _, entry := range section.Entries {
		keys = append(keys, entry.Key)
	}
	return keys
}

func TestNavigationListsPagesAndCollections(t *testing.T) {
	navigation := navigate(t, map[string]string{
		"quiqr/model/base.yaml": `singles:
  - key: home
    title: Homepage
  - key: about
    title: About us
collections:
  - key: posts
    title: Posts
    folder: content/posts/
  - key: team
    title: Team
    folder: content/team/
  - key: products
    title: Products
    folder: content/products/
`,
	})

	if len(navigation.Sections) != 2 {
		t.Fatalf("got %d sections, want pages and collections", len(navigation.Sections))
	}
	if got := entryKeys(navigation.Sections[0]); !reflect.DeepEqual(got, []string{"home", "about"}) {
		t.Errorf("pages = %v", got)
	}
	if got := entryKeys(navigation.Sections[1]); !reflect.DeepEqual(got, []string{"posts", "team", "products"}) {
		t.Errorf("collections = %v", got)
	}
}

func TestCollectionCountsItems(t *testing.T) {
	files := map[string]string{
		"quiqr/model/base.yaml": "collections:\n  - key: posts\n    folder: content/posts/\n    extension: md\n",
	}
	for _, name := range []string{"one", "two", "three", "four", "five", "six", "seven"} {
		files["content/posts/"+name+".md"] = "---\ntitle: " + name + "\n---\n"
	}
	files["content/posts/notes.txt"] = "not an item"
	files["content/posts/drafts/eight.md"] = "a directory is not an item"

	navigation := navigate(t, files)

	entry := navigation.Sections[0].Entries[0]
	if entry.Count == nil || *entry.Count != 7 {
		t.Errorf("count = %v, want 7", entry.Count)
	}
}

func TestCollectionCountSkipsTheIndexWhenAsked(t *testing.T) {
	files := map[string]string{
		"quiqr/model/base.yaml":   "collections:\n  - key: posts\n    folder: content/posts/\n    extension: md\n    hideIndex: true\n",
		"content/posts/_index.md": "---\ntitle: Posts\n---\n",
		"content/posts/first.md":  "---\ntitle: First\n---\n",
	}

	entry := navigate(t, files).Sections[0].Entries[0]
	if entry.Count == nil || *entry.Count != 1 {
		t.Errorf("count = %v, want 1 with the index hidden", entry.Count)
	}
}

func TestEmptyCollectionStillAppears(t *testing.T) {
	navigation := navigate(t, map[string]string{
		"quiqr/model/base.yaml": "collections:\n  - key: posts\n    folder: content/posts/\n",
	})

	if len(navigation.Sections) != 1 || len(navigation.Sections[0].Entries) != 1 {
		t.Fatalf("navigation = %+v, want the empty collection listed", navigation)
	}
	entry := navigation.Sections[0].Entries[0]
	if entry.Count == nil || *entry.Count != 0 {
		t.Errorf("count = %v, want 0 rather than absent", entry.Count)
	}
}

func TestCollectionCarriesItemName(t *testing.T) {
	navigation := navigate(t, map[string]string{
		"quiqr/model/base.yaml": "collections:\n  - key: posts\n    folder: content/posts/\n    itemtitle: article\n",
	})

	if got := navigation.Sections[0].Entries[0].ItemName; got != "article" {
		t.Errorf("itemName = %q, want article", got)
	}
}

func TestItemNameFallbackIsPlain(t *testing.T) {
	navigation := navigate(t, map[string]string{
		"quiqr/model/base.yaml": "collections:\n  - key: posts\n    folder: content/posts/\n",
	})

	name := navigation.Sections[0].Entries[0].ItemName
	if name != DefaultItemName {
		t.Errorf("itemName = %q, want %q", name, DefaultItemName)
	}

	banned := []string{"commit", "branch", "push", "pull", "merge", "repository", "node", "record", "entity"}
	for _, word := range banned {
		if strings.Contains(strings.ToLower(name), word) {
			t.Errorf("the fallback item name uses %q, which an editor should never see", word)
		}
	}
}

func TestMenuDecidesOrder(t *testing.T) {
	navigation := navigate(t, map[string]string{
		"quiqr/model/base.yaml": `singles:
  - key: home
    title: Homepage
  - key: about
    title: About us
collections:
  - key: posts
    title: Posts
    folder: content/posts/
menu:
  - title: Shop
    menuItems:
      - key: posts
      - key: about
  - title: Site
    menuItems:
      - key: home
`,
	})

	if len(navigation.Sections) != 2 {
		t.Fatalf("got %d sections, want the two the menu declares", len(navigation.Sections))
	}
	if navigation.Sections[0].Title != "Shop" || navigation.Sections[1].Title != "Site" {
		t.Errorf("sections = %q, %q, want Shop then Site",
			navigation.Sections[0].Title, navigation.Sections[1].Title)
	}
	if got := entryKeys(navigation.Sections[0]); !reflect.DeepEqual(got, []string{"posts", "about"}) {
		t.Errorf("first section = %v, want the menu order", got)
	}
}

func TestDisabledMenuSectionIsSkipped(t *testing.T) {
	navigation := navigate(t, map[string]string{
		"quiqr/model/base.yaml": `singles:
  - key: home
menu:
  - title: Hidden
    disabled: true
    menuItems:
      - key: home
  - title: Site
    menuItems:
      - key: home
`,
	})

	if len(navigation.Sections) != 1 || navigation.Sections[0].Title != "Site" {
		t.Errorf("sections = %+v, want only the enabled one", navigation.Sections)
	}
}

func TestMenuEntryWithoutATargetIsAWarning(t *testing.T) {
	navigation := navigate(t, map[string]string{
		"quiqr/model/base.yaml": `singles:
  - key: home
menu:
  - title: Site
    menuItems:
      - key: home
      - key: ghost
`,
	})

	if got := entryKeys(navigation.Sections[0]); !reflect.DeepEqual(got, []string{"home"}) {
		t.Errorf("entries = %v, want the missing target omitted", got)
	}
	if len(navigation.Warnings) != 1 || !strings.Contains(navigation.Warnings[0].String(), "ghost") {
		t.Errorf("warnings = %v, want one naming the missing target", navigation.Warnings)
	}
}

func TestWithoutMenuSinglesComeFirst(t *testing.T) {
	navigation := navigate(t, map[string]string{
		"quiqr/model/base.yaml": `collections:
  - key: posts
    folder: content/posts/
singles:
  - key: home
`,
	})

	if navigation.Sections[0].Title != sectionPages {
		t.Errorf("first section = %q, want %q", navigation.Sections[0].Title, sectionPages)
	}
	if navigation.Sections[1].Title != sectionCollections {
		t.Errorf("second section = %q, want %q", navigation.Sections[1].Title, sectionCollections)
	}
}

func TestHiddenCollectionIsNotInNavigation(t *testing.T) {
	files := map[string]string{
		"quiqr/model/base.yaml": `singles:
  - key: home
  - key: secret
    hidden: true
collections:
  - key: posts
    folder: content/posts/
  - key: internal
    folder: content/internal/
    hidden: true
`,
	}

	navigation := navigate(t, files)
	for _, section := range navigation.Sections {
		for _, entry := range section.Entries {
			if entry.Key == "secret" || entry.Key == "internal" {
				t.Errorf("%q is hidden but appears in the navigation", entry.Key)
			}
		}
	}

	m, _ := Load(site(files))
	if _, ok := m.Collection("internal"); !ok {
		t.Error("a hidden collection is not reachable by key")
	}
	if _, ok := m.Single("secret"); !ok {
		t.Error("a hidden single is not reachable by key")
	}
}

func TestHiddenEntriesAreAlsoHiddenFromAMenu(t *testing.T) {
	navigation := navigate(t, map[string]string{
		"quiqr/model/base.yaml": `singles:
  - key: home
  - key: secret
    hidden: true
menu:
  - title: Site
    menuItems:
      - key: home
      - key: secret
`,
	})

	if got := entryKeys(navigation.Sections[0]); !reflect.DeepEqual(got, []string{"home"}) {
		t.Errorf("entries = %v, want the hidden single omitted", got)
	}
}

func TestNavigateWithoutAFilesystemStillWorks(t *testing.T) {
	m, err := Load(site(map[string]string{
		"quiqr/model/base.yaml": "collections:\n  - key: posts\n    folder: content/posts/\n",
	}))
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	navigation := m.Navigate(nil)
	entry := navigation.Sections[0].Entries[0]
	if entry.Count == nil || *entry.Count != 0 {
		t.Errorf("count = %v, want 0 when there is nothing to count against", entry.Count)
	}
}

func TestEntryTitleFallsBackToTheKey(t *testing.T) {
	navigation := navigate(t, map[string]string{
		"quiqr/model/base.yaml": "singles:\n  - key: home\ncollections:\n  - key: posts\n    folder: content/posts/\n",
	})

	if navigation.Sections[0].Entries[0].Title != "home" {
		t.Errorf("page title = %q, want the key", navigation.Sections[0].Entries[0].Title)
	}
	if navigation.Sections[1].Entries[0].Title != "posts" {
		t.Errorf("collection title = %q, want the key", navigation.Sections[1].Entries[0].Title)
	}
}
