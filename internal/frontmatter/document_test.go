package frontmatter

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestSplitsEachFormat(t *testing.T) {
	cases := map[string]struct {
		file     string
		format   Format
		wantBody string
		wantKey  string
	}{
		"yaml": {
			file:     "---\ntitle: About\n---\n\nThe body.\n",
			format:   FormatYAML,
			wantBody: "The body.\n",
			wantKey:  "title",
		},
		"toml": {
			file:     "+++\ntitle = \"About\"\n+++\n\nThe body.\n",
			format:   FormatTOML,
			wantBody: "The body.\n",
			wantKey:  "title",
		},
		"json": {
			file:     "{\n  \"title\": \"About\"\n}\n\nThe body.\n",
			format:   FormatJSON,
			wantBody: "The body.\n",
			wantKey:  "title",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			document, err := Parse("page.md", []byte(tc.file))
			if err != nil {
				t.Fatalf("Parse() error = %v, want nil", err)
			}
			if document.Format != tc.format {
				t.Errorf("format = %q, want %q", document.Format, tc.format)
			}
			if document.Body != tc.wantBody {
				t.Errorf("body = %q, want %q", document.Body, tc.wantBody)
			}
			if value, ok := document.Fields.Get(tc.wantKey); !ok || value != "About" {
				t.Errorf("%s = %v, want About", tc.wantKey, value)
			}
		})
	}
}

func TestFileWithNoFrontMatter(t *testing.T) {
	document, err := Parse("page.md", []byte("Just a body.\n"))
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}
	if document.Format != FormatNone {
		t.Errorf("format = %q, want %q", document.Format, FormatNone)
	}
	if document.Fields.Len() != 0 {
		t.Errorf("fields = %v, want none", document.Fields.Keys())
	}
	if document.Body != "Just a body.\n" {
		t.Errorf("body = %q", document.Body)
	}
}

func TestFrontMatterWithNoBody(t *testing.T) {
	document, err := Parse("page.md", []byte("---\ntitle: About\n---\n"))
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}
	if document.Body != "\n" && document.Body != "" {
		t.Errorf("body = %q, want nothing after the front matter", document.Body)
	}
	if _, ok := document.Fields.Get("title"); !ok {
		t.Error("the front matter was not read")
	}
}

func TestBodyContainingADelimiter(t *testing.T) {
	body := "Before.\n\n---\n\nAfter.\n"
	document, err := Parse("page.md", []byte("---\ntitle: About\n---\n\n"+body))
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}
	if document.Body != body {
		t.Errorf("body = %q, want %q", document.Body, body)
	}

	written, err := document.Bytes()
	if err != nil {
		t.Fatalf("Bytes() error = %v, want nil", err)
	}
	if !strings.Contains(string(written), body) {
		t.Errorf("the body's own delimiter did not survive:\n%s", written)
	}
}

func TestUnparseableFrontMatter(t *testing.T) {
	_, err := Parse("page.md", []byte("---\ntitle: [unterminated\n---\n\nbody\n"))
	if err == nil {
		t.Fatal("Parse() error = nil, want a parse error")
	}

	var parseErr *ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("error = %v, want a *ParseError", err)
	}
	if parseErr.File != "page.md" {
		t.Errorf("file = %q, want the file named", parseErr.File)
	}
}

func TestKeyOrderIsKept(t *testing.T) {
	document, err := Parse("page.md", []byte("---\nzebra: 1\napple: 2\nmango: 3\n---\n\nbody\n"))
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"zebra", "apple", "mango"}
	got := document.Fields.Keys()
	for i, key := range want {
		if got[i] != key {
			t.Fatalf("key order = %v, want %v", got, want)
		}
	}
}

func TestWrittenLayoutMatchesTheReference(t *testing.T) {
	golden, err := os.ReadFile(filepath.Join(goldenDir, "scalar-plain.document.golden"))
	if err != nil {
		t.Fatalf("reading the golden failed: %v", err)
	}

	document := &Document{Format: FormatYAML, Fields: NewMap(), Body: "The body.\n"}
	document.Fields.Set("note", "About us")

	written, err := document.Bytes()
	if err != nil {
		t.Fatalf("Bytes() error = %v, want nil", err)
	}
	if string(written) != string(golden) {
		t.Errorf("written:\n%q\nreference:\n%q", written, golden)
	}
}

func TestEveryDocumentGoldenRoundTrips(t *testing.T) {
	index := loadIndex(t)

	for _, entry := range index.Cases {
		if !entry.Document {
			continue
		}

		t.Run(entry.Name, func(t *testing.T) {
			path := filepath.Join(goldenDir, entry.Name+".document.golden")
			original, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("reading the golden failed: %v", err)
			}

			document, err := Parse(entry.Name, original)
			if err != nil {
				t.Fatalf("Parse() error = %v, want nil", err)
			}

			written, err := document.Bytes()
			if err != nil {
				t.Fatalf("Bytes() error = %v, want nil", err)
			}

			if string(written) != string(original) {
				t.Errorf("a round trip changed the file.\noriginal:\n%q\nwritten:\n%q", original, written)
			}
		})
	}
}

func TestWritingTOMLIsRefused(t *testing.T) {
	document, err := Parse("page.md", []byte("+++\ntitle = \"About\"\n+++\n\nbody\n"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = document.Bytes()
	if !errors.Is(err, ErrUnwritableFormat) {
		t.Fatalf("Bytes() error = %v, want ErrUnwritableFormat", err)
	}
	if !strings.Contains(err.Error(), string(FormatTOML)) {
		t.Errorf("error = %q, want it to name the format", err)
	}
}

func TestWritingJSONIsRefused(t *testing.T) {
	document, err := Parse("page.md", []byte("{\n  \"title\": \"About\"\n}\n\nbody\n"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = document.Bytes()
	if !errors.Is(err, ErrUnwritableFormat) {
		t.Fatalf("Bytes() error = %v, want ErrUnwritableFormat", err)
	}
	if !strings.Contains(err.Error(), string(FormatJSON)) {
		t.Errorf("error = %q, want it to name the format", err)
	}
}

func TestReadsTOMLFrontMatter(t *testing.T) {
	document, err := Parse("page.md", []byte(`+++
title = "Korenaar"
weight = 3
draft = false
tags = ["bread", "sourdough"]

[seo]
description = "Fresh bread"
+++

The body.
`))
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	if value, _ := document.Fields.Get("title"); value != "Korenaar" {
		t.Errorf("title = %v", value)
	}
	if value, _ := document.Fields.Get("weight"); value != int64(3) {
		t.Errorf("weight = %#v, want int64(3)", value)
	}
	tags, _ := document.Fields.Get("tags")
	if list, ok := tags.([]any); !ok || len(list) != 2 {
		t.Errorf("tags = %#v, want two entries", tags)
	}
	seo, _ := document.Fields.Get("seo")
	if nested, ok := seo.(*Map); !ok || nested.Len() != 1 {
		t.Errorf("seo = %#v, want a nested mapping", seo)
	}
}

func TestReadsJSONFrontMatter(t *testing.T) {
	document, err := Parse("page.md", []byte(`{
  "title": "Korenaar",
  "weight": 3,
  "tags": ["bread"],
  "seo": { "description": "Fresh bread" }
}

The body.
`))
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}
	if value, _ := document.Fields.Get("weight"); value != int64(3) {
		t.Errorf("weight = %#v, want int64(3)", value)
	}
	if document.Body != "The body.\n" {
		t.Errorf("body = %q", document.Body)
	}
}

// TestReferenceTOMLOutputIsUnreadableByItsOwnReader records why writing TOML
// is refused. The expression is the one the reference implementation uses to
// find the closing delimiter; the input is what its own writer produces.
func TestReferenceTOMLOutputIsUnreadableByItsOwnReader(t *testing.T) {
	index := loadIndex(t)
	if index.Tomlify == "" {
		t.Skip("no tomlify output was recorded")
	}

	header, err := os.ReadFile(filepath.Join(goldenDir, "scalar-plain.toml.golden"))
	if err != nil {
		t.Skipf("no TOML golden to check: %v", err)
	}
	if strings.HasSuffix(string(header), "\n") {
		t.Fatalf("tomlify %s now ends its output with a newline; writing TOML may be reconsidered",
			index.Tomlify)
	}

	// `+++\n${header}+++\n\n${content}` is the reference's dumpContent.
	produced := "+++\n" + string(header) + "+++\n\nThe body.\n"

	// /^[+]{3} *(\r?\n|\r|^)/gm, as the reference's parseFromMdFileString uses.
	expression := regexp.MustCompile(`(?m)^[+]{3} *(\r?\n|\r|$)`)
	matches := expression.FindAllStringIndex(produced, -1)

	if len(matches) >= 2 {
		t.Fatalf("the reference can now read its own TOML output; writing TOML may be reconsidered")
	}
	t.Logf("tomlify %s: the reference finds %d delimiter(s) in its own output and needs two",
		index.Tomlify, len(matches))
}

func TestGoldenIndexRecordsTheReferenceVersions(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(goldenDir, "index.json"))
	if err != nil {
		t.Fatal(err)
	}

	var index map[string]any
	if err := json.Unmarshal(raw, &index); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"jsYaml", "tomlify"} {
		if value, ok := index[key].(string); !ok || value == "" {
			t.Errorf("the goldens do not record %s", key)
		}
	}
}
