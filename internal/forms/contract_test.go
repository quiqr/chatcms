package forms_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/quiqr/chatcms/internal/forms"
)

const goldenDir = "../../test/fixtures/content/golden"

// golden reads the bytes js-yaml produced for a case, which is the front
// matter this form is expected to write.
func golden(t *testing.T, name string) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(goldenDir, name+".yaml.golden"))
	if err != nil {
		t.Fatalf("reading the golden failed: %v; run scripts/generate-goldens.sh", err)
	}
	return "---\n" + string(data) + "---\n\n"
}

// edit is one change made through a form, and the state the reference
// implementation would have written afterwards.
type edit struct {
	fields string
	start  string
	path   string
	value  any
	clear  bool
	golden string
}

func TestEditedDocumentMatchesTheReference(t *testing.T) {
	body := "The body.\n"

	cases := map[string]edit{
		"a string": {
			fields: "      - key: title\n        type: string\n      - key: weight\n        type: number\n",
			start:  "---\ntitle: About\nweight: 3\n---\n\n" + body,
			path:   "title", value: "Our story", golden: "edit-string",
		},
		"a number": {
			fields: "      - key: title\n        type: string\n      - key: weight\n        type: number\n",
			start:  "---\ntitle: About\nweight: 3\n---\n\n" + body,
			path:   "weight", value: int64(7), golden: "edit-number",
		},
		"a boolean": {
			fields: "      - key: title\n        type: string\n      - key: draft\n        type: boolean\n",
			start:  "---\ntitle: About\ndraft: false\n---\n\n" + body,
			path:   "draft", value: true, golden: "edit-boolean",
		},
		"a date": {
			fields: "      - key: title\n        type: string\n      - key: date\n        type: date\n",
			start:  "---\ntitle: About\ndate: 2024-03-01T10:00:00.000Z\n---\n\n" + body,
			path:   "date", value: time.Date(2025, 1, 15, 9, 30, 0, 0, time.UTC), golden: "edit-date",
		},
		"a select": {
			fields: "      - key: title\n        type: string\n      - key: layout\n        type: select\n        options:\n          - plain\n          - wide\n",
			start:  "---\ntitle: About\nlayout: plain\n---\n\n" + body,
			path:   "layout", value: "wide", golden: "edit-select",
		},
		"a list of values": {
			fields: "      - key: title\n        type: string\n      - key: tags\n        type: leaf-array\n        field:\n          key: tag\n          type: string\n",
			start:  "---\ntitle: About\ntags:\n  - news\n  - events\n---\n\n" + body,
			path:   "tags", value: []any{"news", "events", "bakery"}, golden: "edit-leaf-array",
		},
		"a value inside a group": {
			fields: "      - key: title\n        type: string\n      - key: seo\n        type: section\n        fields:\n          - key: description\n            type: string\n          - key: noindex\n            type: boolean\n",
			start:  "---\ntitle: About\nseo:\n  description: A page\n  noindex: false\n---\n\n" + body,
			path:   "seo.description", value: "A new description", golden: "edit-section",
		},
		"a value inside a pull": {
			fields: "      - key: title\n        type: string\n      - key: meta\n        type: pull\n        group: metadata\n        fields:\n          - key: author\n            type: string\n",
			start:  "---\ntitle: About\nmetadata:\n  author: Joris\n---\n\n" + body,
			path:   "meta.author", value: "Marieke", golden: "edit-pull",
		},
		"a long value that folds": {
			fields: "      - key: title\n        type: string\n      - key: summary\n        type: string\n",
			start:  "---\ntitle: About\nsummary: Short\n---\n\n" + body,
			path:   "summary",
			value:  "This is a fairly long summary that will certainly exceed the default line width that js-yaml uses when it dumps a document.",
			golden: "edit-long-text",
		},
		"a value with more than one line": {
			fields: "      - key: title\n        type: string\n      - key: intro\n        type: string\n",
			start:  "---\ntitle: About\nintro: Short\n---\n\n" + body,
			path:   "intro", value: "first line\nsecond line\n", golden: "edit-multiline",
		},
		"a value that is added": {
			fields: "      - key: title\n        type: string\n      - key: summary\n        type: string\n      - key: weight\n        type: number\n",
			start:  "---\ntitle: About\nweight: 3\n---\n\n" + body,
			path:   "summary", value: "A short summary", golden: "edit-added",
		},
		"a value that is cleared": {
			fields: "      - key: title\n        type: string\n      - key: summary\n        type: string\n",
			start:  "---\ntitle: About\nsummary: Something\n---\n\n" + body,
			path:   "summary", clear: true, golden: "edit-cleared",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			form := load(t, tc.fields, tc.start)

			var err error
			if tc.clear {
				err = form.Clear(tc.path)
			} else {
				err = form.Set(tc.path, tc.value)
			}
			if err != nil {
				t.Fatalf("editing %s failed: %v", tc.path, err)
			}

			written, err := form.Bytes()
			if err != nil {
				t.Fatalf("Bytes() error = %v, want nil", err)
			}

			want := golden(t, tc.golden) + body
			if string(written) != want {
				t.Errorf("after editing %s this form wrote:\n%q\nthe reference would have written:\n%q",
					tc.path, written, want)
			}
		})
	}
}

func TestAnEditChangesOnlyWhatWasEdited(t *testing.T) {
	fields := "      - key: title\n        type: string\n      - key: summary\n        type: string\n"
	start := "---\ntitle: About\nsummary: Something\naliases:\n  - /old/\nweight: 3\n---\n\nThe body.\n"

	form := load(t, fields, start)
	if err := form.Set("summary", "Something else"); err != nil {
		t.Fatal(err)
	}

	written, err := form.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	before := strings.Split(start, "\n")
	after := strings.Split(string(written), "\n")

	if len(before) != len(after) {
		t.Fatalf("the file gained or lost lines:\n%q\n%q", start, written)
	}

	var changed []int
	for i := range before {
		if before[i] != after[i] {
			changed = append(changed, i)
		}
	}

	if len(changed) != 1 {
		t.Fatalf("%d lines changed, want 1:\n%q\n%q", len(changed), start, written)
	}
	if !strings.HasPrefix(after[changed[0]], "summary:") {
		t.Errorf("the changed line is %q, want the edited value", after[changed[0]])
	}
}

func TestRoundTripIsByteIdenticalForEveryFixture(t *testing.T) {
	entries, err := os.ReadDir(goldenDir)
	if err != nil {
		t.Fatalf("reading the goldens failed: %v", err)
	}

	checked := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".document.golden") {
			continue
		}

		name := entry.Name()
		t.Run(name, func(t *testing.T) {
			original, err := os.ReadFile(filepath.Join(goldenDir, name))
			if err != nil {
				t.Fatal(err)
			}

			// No model: every value is unclaimed, which is the strictest
			// version of the promise that nothing is lost.
			form, err := forms.Load(nil, name, original)
			if err != nil {
				t.Fatalf("Load() error = %v, want nil", err)
			}

			written, err := form.Bytes()
			if err != nil {
				t.Fatalf("Bytes() error = %v, want nil", err)
			}
			if string(written) != string(original) {
				t.Errorf("a round trip changed the file:\noriginal:\n%q\nwritten:\n%q", original, written)
			}
		})
		checked++
	}

	if checked == 0 {
		t.Fatal("no document goldens were checked")
	}
	t.Logf("%d documents round tripped byte for byte", checked)
}
