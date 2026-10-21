package forms_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/quiqr/chatcms/internal/forms"
	"github.com/quiqr/chatcms/internal/frontmatter"
	"github.com/quiqr/chatcms/internal/model"
)

// fieldsFrom builds a field tree by parsing a model, so the tests exercise
// the same path the application does rather than a hand-built tree.
func fieldsFrom(t *testing.T, fieldsYAML string) []model.Field {
	t.Helper()

	site := fstest.MapFS{
		"quiqr/model/base.yaml": &fstest.MapFile{Data: []byte(
			"singles:\n  - key: page\n    file: content/page.md\n    fields:\n" + fieldsYAML)},
	}

	parsed, err := model.Load(site)
	if err != nil {
		t.Fatalf("the test model does not parse: %v", err)
	}
	return parsed.Singles[0].Fields
}

func load(t *testing.T, fieldsYAML, file string) *forms.Form {
	t.Helper()

	form, err := forms.Load(fieldsFrom(t, fieldsYAML), "content/page.md", []byte(file))
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	return form
}

func TestPlainFieldRoundTrips(t *testing.T) {
	form := load(t, `      - key: title
        type: string
`, "---\ntitle: About us\n---\n\nThe body.\n")

	value, ok := form.Get("title")
	if !ok || value != "About us" {
		t.Fatalf("title = %v, %v", value, ok)
	}

	if err := form.Set("title", "Our story"); err != nil {
		t.Fatal(err)
	}

	written, err := form.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(written), "title: Our story") {
		t.Errorf("written:\n%s", written)
	}
}

func TestSectionNests(t *testing.T) {
	form := load(t, `      - key: seo
        type: section
        fields:
          - key: description
            type: string
`, "---\nseo:\n  description: A page\n---\n\nbody\n")

	value, ok := form.Get("seo.description")
	if !ok || value != "A page" {
		t.Fatalf("seo.description = %v, %v", value, ok)
	}

	if err := form.Set("seo.description", "Another page"); err != nil {
		t.Fatal(err)
	}
	written, _ := form.Bytes()
	if !strings.Contains(string(written), "seo:\n  description: Another page") {
		t.Errorf("written:\n%s", written)
	}
}

func TestSectionWithoutGroupdataDoesNot(t *testing.T) {
	form := load(t, `      - key: seo
        type: section
        groupdata: false
        fields:
          - key: description
            type: string
`, "---\ndescription: A page\n---\n\nbody\n")

	value, ok := form.Get("seo.description")
	if !ok || value != "A page" {
		t.Fatalf("seo.description = %v, %v, want the value from beside the group", value, ok)
	}

	if err := form.Set("seo.description", "Another page"); err != nil {
		t.Fatal(err)
	}
	written, _ := form.Bytes()
	if strings.Contains(string(written), "seo:") {
		t.Errorf("the group was written as a key although it does not group its data:\n%s", written)
	}
	if !strings.Contains(string(written), "description: Another page") {
		t.Errorf("written:\n%s", written)
	}
}

func TestPullUsesItsGroup(t *testing.T) {
	form := load(t, `      - key: meta
        type: pull
        group: metadata
        fields:
          - key: author
            type: string
`, "---\nmetadata:\n  author: Joris\n---\n\nbody\n")

	value, ok := form.Get("meta.author")
	if !ok || value != "Joris" {
		t.Fatalf("meta.author = %v, %v", value, ok)
	}
}

func TestPullWithoutAGroupUsesItsKey(t *testing.T) {
	form := load(t, `      - key: meta
        type: pull
        fields:
          - key: author
            type: string
`, "---\nmeta:\n  author: Joris\n---\n\nbody\n")

	value, ok := form.Get("meta.author")
	if !ok || value != "Joris" {
		t.Fatalf("meta.author = %v, %v", value, ok)
	}
}

func TestAccordionIsAListOfObjects(t *testing.T) {
	form := load(t, `      - key: blocks
        type: accordion
        fields:
          - key: heading
            type: string
`, "---\nblocks:\n  - heading: One\n  - heading: Two\n---\n\nbody\n")

	value, ok := form.Get("blocks")
	if !ok {
		t.Fatal("blocks is not readable")
	}
	items, isList := value.([]any)
	if !isList || len(items) != 2 {
		t.Fatalf("blocks = %#v, want two entries", value)
	}
	first, isMap := items[0].(*frontmatter.Map)
	if !isMap {
		t.Fatalf("the first block is %#v, want a mapping", items[0])
	}
	if heading, _ := first.Get("heading"); heading != "One" {
		t.Errorf("heading = %v, want One", heading)
	}
}

func TestLeafArrayIsAListOfValues(t *testing.T) {
	form := load(t, `      - key: tags
        type: leaf-array
        field:
          key: tag
          type: string
`, "---\ntags:\n  - news\n  - events\n---\n\nbody\n")

	value, _ := form.Get("tags")
	items, isList := value.([]any)
	if !isList || len(items) != 2 || items[0] != "news" {
		t.Fatalf("tags = %#v, want a list of values", value)
	}
}

func TestNestingToAnyDepth(t *testing.T) {
	form := load(t, `      - key: outer
        type: section
        fields:
          - key: middle
            type: nest
            fields:
              - key: inner
                type: section
                fields:
                  - key: leaf
                    type: string
`, "---\nouter:\n  middle:\n    inner:\n      leaf: deep\n---\n\nbody\n")

	value, ok := form.Get("outer.middle.inner.leaf")
	if !ok || value != "deep" {
		t.Fatalf("the deep value = %v, %v", value, ok)
	}

	if err := form.Set("outer.middle.inner.leaf", "deeper"); err != nil {
		t.Fatal(err)
	}
	written, _ := form.Bytes()
	if !strings.Contains(string(written), "      leaf: deeper") {
		t.Errorf("written:\n%s", written)
	}
}

func TestBodyIsReadableAsAField(t *testing.T) {
	form := load(t, `      - key: title
        type: string
      - key: mainContent
        type: markdown
`, "---\ntitle: About\n---\n\nThe body text.\n")

	value, ok := form.Get("mainContent")
	if !ok || value != "The body text.\n" {
		t.Fatalf("mainContent = %q, %v", value, ok)
	}
}

func TestWritingTheBodyLeavesValuesAlone(t *testing.T) {
	original := "---\ntitle: About\nweight: 3\n---\n\nOld body.\n"
	form := load(t, `      - key: title
        type: string
      - key: mainContent
        type: markdown
`, original)

	if err := form.Set("mainContent", "New body.\n"); err != nil {
		t.Fatal(err)
	}

	written, _ := form.Bytes()
	if string(written) != "---\ntitle: About\nweight: 3\n---\n\nNew body.\n" {
		t.Errorf("written:\n%q", written)
	}
}

func TestUnclaimedValuesSurviveARoundTrip(t *testing.T) {
	original := "---\ntitle: About\naliases:\n  - /old-url/\nweight: 3\n---\n\nbody\n"
	form := load(t, `      - key: title
        type: string
`, original)

	written, err := form.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if string(written) != original {
		t.Errorf("written:\n%q\noriginal:\n%q", written, original)
	}
}

func TestUnclaimedValuesSurviveAnEdit(t *testing.T) {
	original := "---\ntitle: About\naliases:\n  - /old-url/\nweight: 3\n---\n\nbody\n"
	form := load(t, `      - key: title
        type: string
`, original)

	if err := form.Set("title", "Our story"); err != nil {
		t.Fatal(err)
	}

	written, _ := form.Bytes()
	want := "---\ntitle: Our story\naliases:\n  - /old-url/\nweight: 3\n---\n\nbody\n"
	if string(written) != want {
		t.Errorf("written:\n%q\nwant:\n%q", written, want)
	}
}

func TestANewValueLandsWhereTheModelImplies(t *testing.T) {
	form := load(t, `      - key: title
        type: string
      - key: summary
        type: string
      - key: weight
        type: number
`, "---\ntitle: About\nweight: 3\n---\n\nbody\n")

	if err := form.Set("summary", "A short summary"); err != nil {
		t.Fatal(err)
	}

	written, _ := form.Bytes()
	want := "---\ntitle: About\nsummary: A short summary\nweight: 3\n---\n\nbody\n"
	if string(written) != want {
		t.Errorf("written:\n%q\nwant:\n%q", written, want)
	}
}

func TestClearingWritesAnAbsentValue(t *testing.T) {
	form := load(t, `      - key: title
        type: string
      - key: summary
        type: string
`, "---\ntitle: About\nsummary: Something\n---\n\nbody\n")

	if err := form.Clear("summary"); err != nil {
		t.Fatal(err)
	}

	written, _ := form.Bytes()
	if !strings.Contains(string(written), "summary: null") {
		t.Errorf("written:\n%s\nwant the cleared value written the way the reference writes an absent one", written)
	}
}

func TestUnknownFieldIsRefused(t *testing.T) {
	form := load(t, `      - key: title
        type: string
`, "---\ntitle: About\n---\n\nbody\n")

	if _, ok := form.Get("absent"); ok {
		t.Error("Get on an unknown field found something")
	}
	if err := form.Set("absent", "x"); !errors.Is(err, forms.ErrNoSuchField) {
		t.Errorf("Set() error = %v, want ErrNoSuchField", err)
	}
	if err := form.Clear("absent"); !errors.Is(err, forms.ErrNoSuchField) {
		t.Errorf("Clear() error = %v, want ErrNoSuchField", err)
	}
}

func TestLayoutFieldsBindNothing(t *testing.T) {
	form := load(t, `      - key: title
        type: string
      - key: spacer
        type: empty-line
      - key: guidance
        type: info
        content: Fill in the title first.
`, "---\ntitle: About\n---\n\nbody\n")

	if got := form.FieldPaths(); !reflect.DeepEqual(got, []string{"title"}) {
		t.Errorf("field paths = %v, want only the field that stores something", got)
	}
}

func TestBindingsCarryTheFieldAndItsDataPath(t *testing.T) {
	form := load(t, `      - key: seo
        type: section
        fields:
          - key: description
            type: string
`, "---\nseo:\n  description: A page\n---\n\nbody\n")

	bindings := form.Bindings()
	if len(bindings) != 1 {
		t.Fatalf("got %d bindings, want 1", len(bindings))
	}
	if bindings[0].FieldPath != "seo.description" {
		t.Errorf("field path = %q", bindings[0].FieldPath)
	}
	if !reflect.DeepEqual(bindings[0].DataPath, []string{"seo", "description"}) {
		t.Errorf("data path = %v", bindings[0].DataPath)
	}
	if bindings[0].Field.Type != "string" {
		t.Errorf("field type = %q", bindings[0].Field.Type)
	}
}

func TestSettingThroughAGroupThatIsNotThereYet(t *testing.T) {
	form := load(t, `      - key: title
        type: string
      - key: seo
        type: section
        fields:
          - key: description
            type: string
`, "---\ntitle: About\n---\n\nbody\n")

	if err := form.Set("seo.description", "A page"); err != nil {
		t.Fatal(err)
	}

	written, _ := form.Bytes()
	want := "---\ntitle: About\nseo:\n  description: A page\n---\n\nbody\n"
	if string(written) != want {
		t.Errorf("written:\n%q\nwant:\n%q", written, want)
	}
}

func TestLoadRejectsAnUnreadableFile(t *testing.T) {
	_, err := forms.Load(fieldsFrom(t, "      - key: title\n        type: string\n"),
		"content/page.md", []byte("---\ntitle: [broken\n---\n\nbody\n"))
	if err == nil {
		t.Error("Load() error = nil, want a parse error")
	}
}
