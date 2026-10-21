package frontmatter

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestMapKeepsOrderAndAllowsRemoval(t *testing.T) {
	m := NewMap()
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)
	m.Set("b", 20)

	if got := m.Keys(); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Errorf("keys = %v, want insertion order with no duplicate", got)
	}
	if value, _ := m.Get("b"); value != 20 {
		t.Errorf("b = %v, want the value it was set to last", value)
	}
	if m.Len() != 3 {
		t.Errorf("Len() = %d, want 3", m.Len())
	}

	m.Delete("b")
	if got := m.Keys(); !reflect.DeepEqual(got, []string{"a", "c"}) {
		t.Errorf("keys after removal = %v", got)
	}
	m.Delete("absent")
	if m.Len() != 2 {
		t.Errorf("removing a key that is not there changed the map")
	}
}

func TestCloneIsDeep(t *testing.T) {
	inner := NewMap()
	inner.Set("x", 1)

	original := NewMap()
	original.Set("nested", inner)
	original.Set("list", []any{1, inner})

	copied := original.Clone()

	copiedInner, _ := copied.Get("nested")
	copiedInner.(*Map).Set("x", 99)

	if value, _ := inner.Get("x"); value != 1 {
		t.Error("changing the copy changed the original")
	}
	if copied.Len() != original.Len() {
		t.Error("the copy lost a key")
	}
}

func TestSetAfterPlacesANewKey(t *testing.T) {
	m := NewMap()
	m.Set("title", 1)
	m.Set("weight", 2)

	m.SetAfter("summary", 3, []string{"title"})
	if got := m.Keys(); !reflect.DeepEqual(got, []string{"title", "summary", "weight"}) {
		t.Errorf("keys = %v, want the new key after the one it follows", got)
	}

	m.SetAfter("slug", 4, nil)
	if m.Keys()[0] != "slug" {
		t.Errorf("keys = %v, want a key with nothing to follow at the front", m.Keys())
	}

	m.SetAfter("title", 10, []string{"weight"})
	if value, _ := m.Get("title"); value != 10 {
		t.Error("SetAfter did not update an existing key")
	}
	if m.Keys()[1] != "title" {
		t.Errorf("keys = %v, want an existing key to stay where it was", m.Keys())
	}
}

type notAValue struct{ x int }

func TestEmitterRefusesAnUnsupportedShape(t *testing.T) {
	cases := map[string]*Map{
		"at the top level": func() *Map {
			m := NewMap()
			m.Set("odd", notAValue{})
			return m
		}(),
		"inside a group": func() *Map {
			inner := NewMap()
			inner.Set("odd", notAValue{})
			m := NewMap()
			m.Set("seo", inner)
			return m
		}(),
		"inside a list": func() *Map {
			m := NewMap()
			m.Set("things", []any{"fine", notAValue{}})
			return m
		}(),
		"inside a list inside a list": func() *Map {
			m := NewMap()
			m.Set("grid", []any{[]any{notAValue{}}})
			return m
		}(),
		"inside a group inside a list": func() *Map {
			inner := NewMap()
			inner.Set("odd", notAValue{})
			m := NewMap()
			m.Set("items", []any{inner})
			return m
		}(),
	}

	for name, value := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := EmitYAML(value)
			if err == nil {
				t.Fatal("EmitYAML() error = nil, want a refusal")
			}

			var unsupportedErr *UnsupportedError
			if !errors.As(err, &unsupportedErr) {
				t.Fatalf("error = %v, want an *UnsupportedError", err)
			}
			if unsupportedErr.Path == "" {
				t.Error("the refusal names no value")
			}
			if !strings.Contains(unsupportedErr.Error(), "notAValue") {
				t.Errorf("error = %q, want it to say what kind the value was", unsupportedErr)
			}
		})
	}
}

func TestEmitsListsInsideLists(t *testing.T) {
	m := NewMap()
	m.Set("grid", []any{[]any{"a", "b"}, []any{"c"}})

	out, err := EmitYAML(m)
	if err != nil {
		t.Fatalf("EmitYAML() error = %v, want nil", err)
	}
	want := "grid:\n  - - a\n    - b\n  - - c\n"
	if string(out) != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestEmitsEmptyThingsInsideLists(t *testing.T) {
	m := NewMap()
	m.Set("items", []any{NewMap(), []any{}})

	out, err := EmitYAML(m)
	if err != nil {
		t.Fatalf("EmitYAML() error = %v, want nil", err)
	}
	if !strings.Contains(string(out), "- {}") || !strings.Contains(string(out), "- []") {
		t.Errorf("out = %q, want an empty mapping and an empty list", out)
	}
}

func TestParseErrorCarriesTheCause(t *testing.T) {
	cause := errors.New("something went wrong")
	err := &ParseError{File: "page.md", Format: FormatYAML, Err: cause}

	if !strings.Contains(err.Error(), "page.md") {
		t.Errorf("error = %q, want it to name the file", err)
	}
	if !errors.Is(err, cause) {
		t.Error("the cause cannot be reached through the error")
	}
}

func TestTimestampsAreRead(t *testing.T) {
	cases := map[string]string{
		"a bare date":          "2024-03-01",
		"a date and time":      "2024-03-01T10:00:00Z",
		"a space separator":    "2024-03-01 10:00:00",
		"a fraction":           "2024-03-01T10:00:00.250Z",
		"an offset":            "2024-03-01T10:00:00+02:00",
		"an offset in minutes": "2024-03-01T10:00:00+02:30",
	}

	for name, value := range cases {
		t.Run(name, func(t *testing.T) {
			when, ok := asTimestamp(value)
			if !ok {
				t.Fatalf("asTimestamp(%q) did not recognise it", value)
			}
			if when.Year() != 2024 {
				t.Errorf("year = %d, want 2024", when.Year())
			}
			if when.Location() != time.UTC {
				t.Errorf("location = %v, want UTC", when.Location())
			}
		})
	}

	for _, value := range []string{"not a date", "2024-13-01x", "0.110.0", ""} {
		if _, ok := asTimestamp(value); ok {
			t.Errorf("asTimestamp(%q) recognised something that is not a moment in time", value)
		}
	}
}

func TestAnOffsetIsConvertedToUTC(t *testing.T) {
	when, ok := asTimestamp("2024-03-01T10:00:00+02:00")
	if !ok {
		t.Fatal("the timestamp was not recognised")
	}
	if when.Hour() != 8 {
		t.Errorf("hour = %d, want 8 after converting to UTC", when.Hour())
	}
}

func TestEmptyDocumentWritesOnlyItsBody(t *testing.T) {
	document := &Document{Format: FormatYAML, Fields: NewMap(), Body: "Just a body.\n"}

	out, err := document.Bytes()
	if err != nil {
		t.Fatalf("Bytes() error = %v, want nil", err)
	}
	if string(out) != "Just a body.\n" {
		t.Errorf("out = %q, want the body alone", out)
	}
}

func TestReadingAListAtTheTopLevelIsRefused(t *testing.T) {
	_, err := Parse("page.md", []byte("---\n- a\n- b\n---\n\nbody\n"))
	if err == nil {
		t.Error("Parse() error = nil, want front matter that is not a mapping to be refused")
	}
}

func TestIsArrayIndexKey(t *testing.T) {
	for _, key := range []string{"0", "1", "42"} {
		if !isArrayIndexKey(key) {
			t.Errorf("isArrayIndexKey(%q) = false, want true", key)
		}
	}
	for _, key := range []string{"", "01", "-1", "1.5", "4294967295", "title", "1a"} {
		if isArrayIndexKey(key) {
			t.Errorf("isArrayIndexKey(%q) = true, want false", key)
		}
	}
}

func TestFoldLinesBreaksAtSpaces(t *testing.T) {
	lines := foldLines("one two three four", 9)
	for _, line := range lines {
		if strings.HasPrefix(line, " ") || strings.HasSuffix(line, " ") {
			t.Errorf("line %q keeps a space at an edge", line)
		}
	}
	if strings.Join(lines, " ") != "one two three four" {
		t.Errorf("folding changed the text: %v", lines)
	}
}

func TestFoldedStringReadsBackUnchanged(t *testing.T) {
	long := "This is a fairly long summary that will certainly exceed the default line width " +
		"that js-yaml uses when it dumps a document, and then keep going for a while after that."

	m := NewMap()
	m.Set("summary", long)

	out, err := EmitYAML(m)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), ">-") {
		t.Fatalf("the long value was not folded:\n%s", out)
	}

	document, err := Parse("page.md", []byte("---\n"+string(out)+"---\n\nbody\n"))
	if err != nil {
		t.Fatalf("the folded output cannot be read back: %v", err)
	}
	value, _ := document.Fields.Get("summary")
	if value != long {
		t.Errorf("read back %q, want the original string", value)
	}
}
