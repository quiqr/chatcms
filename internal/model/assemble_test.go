package model

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

func site(files map[string]string) fstest.MapFS {
	out := fstest.MapFS{}
	for name, contents := range files {
		out[name] = &fstest.MapFile{Data: []byte(contents)}
	}
	return out
}

func assemble(t *testing.T, files map[string]string) *Assembly {
	t.Helper()

	assembly, err := Assemble(site(files))
	if err != nil {
		t.Fatalf("Assemble() error = %v, want nil", err)
	}
	return assembly
}

func TestLocateBaseFile(t *testing.T) {
	cases := map[string]struct {
		files map[string]string
		want  string
	}{
		"the current location": {
			files: map[string]string{"quiqr/model/base.yaml": "title: a\n"},
			want:  "quiqr/model/base.yaml",
		},
		"the legacy location": {
			files: map[string]string{"sukoh.yaml": "title: a\n"},
			want:  "sukoh.yaml",
		},
		"the current location wins over the legacy one": {
			files: map[string]string{
				"quiqr/model/base.yaml": "title: a\n",
				"sukoh.yaml":            "title: b\n",
			},
			want: "quiqr/model/base.yaml",
		},
		"json wins over yaml when both are present": {
			files: map[string]string{
				"quiqr/model/base.json": `{"title":"a"}`,
				"quiqr/model/base.yaml": "title: b\n",
			},
			want: "quiqr/model/base.json",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := assemble(t, tc.files).BaseFile; got != tc.want {
				t.Errorf("BaseFile = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNoModelIsDistinguishableFromAParseFailure(t *testing.T) {
	_, err := Assemble(site(map[string]string{"content/index.md": "hello"}))
	if !errors.Is(err, ErrNoModel) {
		t.Fatalf("Assemble() error = %v, want ErrNoModel", err)
	}

	_, err = Assemble(site(map[string]string{"quiqr/model/base.yaml": "title: [broken\n"}))
	if errors.Is(err, ErrNoModel) {
		t.Errorf("a broken model reported as ErrNoModel: %v", err)
	}
	if err == nil {
		t.Error("a broken model parsed without error")
	}
}

func TestIncludesAppendToLists(t *testing.T) {
	assembly := assemble(t, map[string]string{
		"quiqr/model/base.yaml":                 "collections:\n  - key: base\n    folder: content/base/\n",
		"quiqr/model/includes/collections.yaml": "- key: included\n  folder: content/included/\n",
	})

	collections := assembly.Tree["collections"].([]any)
	if len(collections) != 2 {
		t.Fatalf("got %d collections, want 2", len(collections))
	}
	if collections[0].(map[string]any)["key"] != "base" {
		t.Errorf("first collection = %v, want the base file's entry first", collections[0])
	}
	if collections[1].(map[string]any)["key"] != "included" {
		t.Errorf("second collection = %v, want the include's entry", collections[1])
	}
}

func TestIncludesMergeObjectsUnderTheirFilename(t *testing.T) {
	assembly := assemble(t, map[string]string{
		"quiqr/model/base.yaml":           "title: Korenaar\n",
		"quiqr/model/includes/build.yaml": "command: hugo\nflags:\n  - --minify\n",
	})

	build, ok := assembly.Tree["build"].(map[string]any)
	if !ok {
		t.Fatalf("build = %#v, want the include's object", assembly.Tree["build"])
	}
	if build["command"] != "hugo" {
		t.Errorf("build.command = %v, want hugo", build["command"])
	}
}

func TestBaseWinsOverInclude(t *testing.T) {
	assembly := assemble(t, map[string]string{
		"quiqr/model/base.yaml":              "settings:\n  theme: base\n  shared: from-base\n",
		"quiqr/model/includes/settings.yaml": "theme: include\nonly: from-include\n",
	})

	settings := assembly.Tree["settings"].(map[string]any)
	if settings["theme"] != "base" {
		t.Errorf("theme = %v, want the base file to win", settings["theme"])
	}
	if settings["only"] != "from-include" {
		t.Errorf("only = %v, want the include's key to survive", settings["only"])
	}
}

func TestIncludesAreMergedInFilenameOrder(t *testing.T) {
	assembly := assemble(t, map[string]string{
		"quiqr/model/base.yaml":               "title: a\n",
		"quiqr/model/includes/a-singles.yaml": "key: first\n",
		"quiqr/model/includes/b-singles.yaml": "key: second\n",
	})

	if assembly.Tree["a-singles"].(map[string]any)["key"] != "first" {
		t.Error("the first include did not land under its own filename")
	}
	if assembly.Tree["b-singles"].(map[string]any)["key"] != "second" {
		t.Error("the second include did not land under its own filename")
	}
}

func TestSubIncludesAddOneItemEach(t *testing.T) {
	assembly := assemble(t, map[string]string{
		"quiqr/model/base.yaml":                       "title: a\n",
		"quiqr/model/includes/singles/about.yaml":     "key: about\nfile: content/about.md\n",
		"quiqr/model/includes/singles/contact.yaml":   "key: contact\nfile: content/contact.md\n",
		"quiqr/model/includes/collections/posts.yaml": "key: posts\nfolder: content/posts/\n",
		"quiqr/model/includes/menus/site.yaml":        "title: Site\nmenuItems:\n  - key: about\n",
	})

	if got := len(assembly.Tree["singles"].([]any)); got != 2 {
		t.Errorf("got %d singles, want 2", got)
	}
	if got := len(assembly.Tree["collections"].([]any)); got != 1 {
		t.Errorf("got %d collections, want 1", got)
	}
	if got := len(assembly.Tree["menu"].([]any)); got != 1 {
		t.Errorf("got %d menu sections, want 1", got)
	}
}

func TestBothIncludeStylesAtOnce(t *testing.T) {
	assembly := assemble(t, map[string]string{
		"quiqr/model/base.yaml":                     "title: a\n",
		"quiqr/model/includes/collections.yaml":     "- key: listed\n  folder: content/listed/\n",
		"quiqr/model/includes/collections/own.yaml": "key: own\nfolder: content/own/\n",
	})

	collections := assembly.Tree["collections"].([]any)
	keys := []string{}
	for _, entry := range collections {
		keys = append(keys, entry.(map[string]any)["key"].(string))
	}
	if !reflect.DeepEqual(keys, []string{"listed", "own"}) {
		t.Errorf("keys = %v, want each collection exactly once", keys)
	}
}

func TestPartialSuppliesFields(t *testing.T) {
	assembly := assemble(t, map[string]string{
		"quiqr/model/base.yaml":             "collections:\n  - key: posts\n    folder: content/posts/\n    _mergePartial: article\n",
		"quiqr/model/partials/article.yaml": "title: Article\nfields:\n  - key: title\n    type: string\n",
	})

	posts := assembly.Tree["collections"].([]any)[0].(map[string]any)
	if posts["title"] != "Article" {
		t.Errorf("title = %v, want the partial's title", posts["title"])
	}
	if got := len(posts["fields"].([]any)); got != 1 {
		t.Errorf("got %d fields, want the partial's field", got)
	}
	if _, present := posts[mergePartial]; present {
		t.Error("_mergePartial survived the merge")
	}
}

func TestItemOverridesPartial(t *testing.T) {
	assembly := assemble(t, map[string]string{
		"quiqr/model/base.yaml":          "singles:\n  - key: about\n    title: Our story\n    _mergePartial: page\n",
		"quiqr/model/partials/page.yaml": "title: Page\nfile: content/page.md\n",
	})

	about := assembly.Tree["singles"].([]any)[0].(map[string]any)
	if about["title"] != "Our story" {
		t.Errorf("title = %v, want the item to win", about["title"])
	}
	if about["file"] != "content/page.md" {
		t.Errorf("file = %v, want the partial's value to survive", about["file"])
	}
}

func TestPartialFieldOrderMatchesReference(t *testing.T) {
	assembly := assemble(t, map[string]string{
		"quiqr/model/base.yaml": `collections:
  - key: posts
    folder: content/posts/
    _mergePartial: article
    fields:
      - key: title
        type: string
        title: Own title
      - key: extra
        type: string
`,
		"quiqr/model/partials/article.yaml": `fields:
  - key: title
    type: string
    title: Partial title
  - key: body
    type: markdown
`,
	})

	posts := assembly.Tree["collections"].([]any)[0].(map[string]any)
	fields := posts["fields"].([]any)

	keys := []string{}
	for _, field := range fields {
		keys = append(keys, field.(map[string]any)["key"].(string))
	}
	if !reflect.DeepEqual(keys, []string{"body", "title", "extra"}) {
		t.Errorf("field order = %v, want [body title extra] as the reference application produces", keys)
	}

	title := fields[1].(map[string]any)
	if title["title"] != "Own title" {
		t.Errorf("the surviving title field came from the partial: %v", title)
	}
}

func TestRemotePartialFromCache(t *testing.T) {
	assembly := assemble(t, map[string]string{
		"quiqr/model/base.yaml": "singles:\n  - key: about\n    _mergePartial: https://example.com/page.yaml\n",
		"quiqr/model/partialsRemoteCache/https%3A%2F%2Fexample.com%2Fpage.yaml": "file: content/about.md\n",
	})

	about := assembly.Tree["singles"].([]any)[0].(map[string]any)
	if about["file"] != "content/about.md" {
		t.Errorf("file = %v, want the cached partial to be used", about["file"])
	}
}

func TestUncachedRemotePartialIsAnError(t *testing.T) {
	_, err := Assemble(site(map[string]string{
		"quiqr/model/base.yaml": "singles:\n  - key: about\n    _mergePartial: https://example.com/page.yaml\n",
	}))

	if err == nil {
		t.Fatal("Assemble() error = nil, want an error about the uncached partial")
	}
	if !strings.Contains(err.Error(), "https://example.com/page.yaml") {
		t.Errorf("error = %q, want it to name the partial", err)
	}
}

func TestDogfoodPartialIsAWarning(t *testing.T) {
	assembly := assemble(t, map[string]string{
		"quiqr/model/base.yaml": "singles:\n  - key: dogfoodMenu\n    _mergePartial: dogfood_site://cmsDashboardMd\n",
	})

	if len(assembly.Warnings) != 1 {
		t.Fatalf("got %d warnings, want 1", len(assembly.Warnings))
	}
	if !strings.Contains(assembly.Warnings[0].String(), "dogfood_site://cmsDashboardMd") {
		t.Errorf("warning = %q, want it to name the partial", assembly.Warnings[0])
	}

	about := assembly.Tree["singles"].([]any)[0].(map[string]any)
	if about["key"] != "dogfoodMenu" {
		t.Error("the single was dropped rather than kept unmerged")
	}
}

func TestMissingLocalPartialIsAWarning(t *testing.T) {
	assembly := assemble(t, map[string]string{
		"quiqr/model/base.yaml": "singles:\n  - key: about\n    _mergePartial: absent\n",
	})

	if len(assembly.Warnings) != 1 {
		t.Fatalf("got %d warnings, want 1", len(assembly.Warnings))
	}
	if !strings.Contains(assembly.Warnings[0].String(), "absent") {
		t.Errorf("warning = %q, want it to name the partial", assembly.Warnings[0])
	}
}

func TestAssemblyWritesNothing(t *testing.T) {
	files := map[string]string{
		"quiqr/model/base.yaml":                 "collections:\n  - key: posts\n    folder: content/posts/\n",
		"quiqr/model/includes/collections.yaml": "- key: pages\n  folder: content/pages/\n",
	}

	fsys := site(files)
	before := len(fsys)

	if _, err := Assemble(fsys); err != nil {
		t.Fatalf("Assemble() error = %v, want nil", err)
	}
	if len(fsys) != before {
		t.Errorf("the site gained %d files during assembly", len(fsys)-before)
	}
}

func TestAssemblyIsDeterministic(t *testing.T) {
	files := map[string]string{
		"quiqr/model/base.yaml":                   "title: a\ncollections:\n  - key: posts\n    folder: content/posts/\n",
		"quiqr/model/includes/collections.yaml":   "- key: pages\n  folder: content/pages/\n",
		"quiqr/model/includes/singles/about.yaml": "key: about\n",
	}

	first := assemble(t, files)
	second := assemble(t, files)

	if !reflect.DeepEqual(first.Tree, second.Tree) {
		t.Error("two assemblies of the same site differ")
	}
}

func TestEmptyListsAreDropped(t *testing.T) {
	assembly := assemble(t, map[string]string{"quiqr/model/base.yaml": "title: a\n"})

	for _, key := range listKeys {
		if _, present := assembly.Tree[key]; present {
			t.Errorf("%s is present but empty", key)
		}
	}
}

func TestABrokenIncludeFailsAssembly(t *testing.T) {
	_, err := Assemble(site(map[string]string{
		"quiqr/model/base.yaml":                 "title: a\n",
		"quiqr/model/includes/collections.yaml": "- key: [broken\n",
	}))

	if err == nil {
		t.Fatal("Assemble() error = nil, want the include's parse error")
	}
	if !strings.Contains(err.Error(), "includes/collections.yaml") {
		t.Errorf("error = %q, want it to name the include", err)
	}
}

func TestEncodeReferenceMatchesEncodeURIComponent(t *testing.T) {
	cases := map[string]string{
		"https://example.com/a.yaml": "https%3A%2F%2Fexample.com%2Fa.yaml",
		"file:///tmp/a.yaml":         "file%3A%2F%2F%2Ftmp%2Fa.yaml",
		"a-b_c.d~e":                  "a-b_c.d~e",
		"a b":                        "a%20b",
	}

	for reference, want := range cases {
		if got := encodeReference(reference); got != want {
			t.Errorf("encodeReference(%q) = %q, want %q", reference, got, want)
		}
	}
}
