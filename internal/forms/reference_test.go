package forms_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/quiqr/chatcms/internal/forms"
	"github.com/quiqr/chatcms/internal/model"
)

const recordingsDir = "../../test/fixtures/reference"

// recording is one edit made in the old Quiqr itself, through its own form,
// and the file it wrote afterwards. The serialisers settle how a value is
// written down; only the application settles what the value has become by
// the time they see it, which is why this case comes from the application
// and not from a library.
type recording struct {
	Single string `json:"single"`
	Field  string `json:"field"`
	Value  string `json:"value"`
	File   string `json:"file"`
	Quiqr  string `json:"quiqr"`
	Site   string `json:"site"`
	// Divergences are the places where ChatCMS deliberately does not do what
	// the old Quiqr does. Every byte of difference is either absent or named
	// here with a reason: an unnamed one fails, and a named one that has
	// stopped happening fails too, so the list cannot quietly rot.
	Divergences []divergence `json:"divergences"`
}

type divergence struct {
	Kind  string `json:"kind"`
	Line  string `json:"line,omitempty"`
	Count int    `json:"count,omitempty"`
	Why   string `json:"why"`
}

func TestRecordedEditsMatchTheReference(t *testing.T) {
	entries, err := os.ReadDir(recordingsDir)
	if err != nil {
		t.Fatalf("reading %s failed: %v", recordingsDir, err)
	}

	found := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		found++
		t.Run(entry.Name(), func(t *testing.T) {
			replay(t, filepath.Join(recordingsDir, entry.Name()))
		})
	}

	if found == 0 {
		t.Fatalf("no recording in %s; make one with scripts/reference.sh record", recordingsDir)
	}
}

func replay(t *testing.T, dir string) {
	t.Helper()

	var recorded recording
	read(t, filepath.Join(dir, "case.json"), &recorded)

	site := os.DirFS(filepath.Join(dir, "site"))
	loaded, err := model.Load(site)
	if err != nil {
		t.Fatalf("the recorded model could not be read: %v", err)
	}

	fields, file := single(t, loaded, recorded.Single)
	if file != recorded.File {
		t.Fatalf("the model says %q keeps %q, the recording says %q",
			recorded.Single, file, recorded.File)
	}

	before, err := os.ReadFile(filepath.Join(dir, "site", file))
	if err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(filepath.Join(dir, "after", file))
	if err != nil {
		t.Fatal(err)
	}

	form, err := forms.Load(fields, filepath.Base(file), before)
	if err != nil {
		t.Fatalf("loading the recorded file failed: %v", err)
	}
	if err := form.Set(recorded.Field, recorded.Value); err != nil {
		t.Fatalf("setting %s failed: %v", recorded.Field, err)
	}
	written, err := form.Bytes()
	if err != nil {
		t.Fatalf("writing failed: %v", err)
	}

	expected := without(t, string(after), recorded.Divergences)
	if string(written) != expected {
		t.Errorf("ChatCMS and the old Quiqr %s do not agree on %s, in a way nothing in case.json names.\n%s",
			short(recorded.Quiqr), file, where(expected, string(written)))
	}

	// A case that only checked the edited value would pass while a
	// neighbour was quietly rewritten, which is the failure this recording
	// exists to catch: the file holds twenty values of other kinds around
	// the one that was edited.
	untouched(t, fields, before, written, recorded.Field)
}

// where says what the first difference is, because two files that differ in
// one byte out of five hundred are unreadable side by side.
func where(expected, written string) string {
	at := 0
	for at < len(expected) && at < len(written) && expected[at] == written[at] {
		at++
	}

	from := at - 40
	if from < 0 {
		from = 0
	}
	return fmt.Sprintf("they are the same for %d bytes, then part ways.\n"+
		"the old Quiqr, less its named divergences (%d bytes): %q\n"+
		"ChatCMS (%d bytes):                                  %q",
		at, len(expected), expected[from:min(len(expected), at+40)],
		len(written), written[from:min(len(written), at+40)])
}

// without takes the named divergences back out of what the old Quiqr wrote,
// so that what is left is what ChatCMS is expected to write exactly. A
// divergence that is not there to take out is a divergence that has stopped
// happening, and the recording is then describing an application that no
// longer behaves that way.
func without(t *testing.T, reference string, divergences []divergence) string {
	t.Helper()

	for _, named := range divergences {
		if named.Why == "" {
			t.Fatalf("the divergence %q says no reason; a difference nobody can explain is not a decision", named.Kind)
		}

		switch named.Kind {
		case "line-added":
			line := named.Line + "\n"
			if !strings.Contains(reference, line) {
				t.Fatalf("case.json says the old Quiqr adds the line %q, and it no longer does", named.Line)
			}
			reference = strings.Replace(reference, line, "", 1)

		case "trailing-newlines":
			if named.Count <= 0 {
				t.Fatalf("the divergence %q says it adds %d newlines", named.Kind, named.Count)
			}
			tail := strings.Repeat("\n", named.Count)
			if !strings.HasSuffix(reference, tail) {
				t.Fatalf("case.json says the old Quiqr adds %d newlines at the end, and it no longer does", named.Count)
			}
			reference = strings.TrimSuffix(reference, tail)

		default:
			t.Fatalf("case.json names a kind of divergence this test does not know: %q", named.Kind)
		}
	}
	return reference
}

// untouched says every value the edit did not name is what it was.
func untouched(t *testing.T, fields []model.Field, before, written []byte, edited string) {
	t.Helper()

	was, err := forms.Load(fields, "before.md", before)
	if err != nil {
		t.Fatal(err)
	}
	is, err := forms.Load(fields, "written.md", written)
	if err != nil {
		t.Fatalf("what we wrote cannot be read back: %v", err)
	}

	for _, path := range was.FieldPaths() {
		if path == edited {
			continue
		}
		left, hadIt := was.Get(path)
		right, hasIt := is.Get(path)
		if hadIt != hasIt {
			t.Errorf("%s was %v before and %v after, though the edit named %s",
				path, hadIt, hasIt, edited)
			continue
		}
		if hadIt && !sameValue(left, right) {
			t.Errorf("%s changed from %#v to %#v, though the edit named %s",
				path, left, right, edited)
		}
	}
}

func sameValue(left, right any) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	if leftErr != nil || rightErr != nil {
		return leftErr == nil && rightErr == nil
	}
	return string(leftJSON) == string(rightJSON)
}

func single(t *testing.T, loaded *model.Model, key string) ([]model.Field, string) {
	t.Helper()

	for _, one := range loaded.Singles {
		if one.Key == key {
			return one.Fields, one.File
		}
	}
	t.Fatalf("the recorded model has no single called %q", key)
	return nil, ""
}

func read(t *testing.T, path string, into any) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s failed: %v; make a recording with scripts/reference.sh record", path, err)
	}
	if err := json.Unmarshal(data, into); err != nil {
		t.Fatalf("%s is not what this test knows: %v", path, err)
	}
}

func short(commit string) string {
	if len(commit) > 8 {
		return "at " + commit[:8]
	}
	return ""
}
