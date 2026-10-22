package publish_test

import (
	"strings"
	"testing"

	"github.com/quiqr/chatcms/internal/publish"
)

func rebuilt(parts []publish.Word, want publish.Part) string {
	out := strings.Builder{}
	for _, part := range parts {
		if part.Part == publish.Kept || part.Part == want {
			out.WriteString(part.Text)
		}
	}
	return out.String()
}

func TestAWordDiffCanBeReadBackAsBothTexts(t *testing.T) {
	cases := [][2]string{
		{"The quick brown fox", "The quick red fox"},
		{"", "A new page."},
		{"An old page.", ""},
		{"One two three four five", "One three five"},
		{"Hello", "Hello world"},
		{"Hello world", "Hello"},
		{"a b c", "c b a"},
		{"Line one\nLine two\n", "Line one\nLine two changed\n"},
		{"  leading space", "leading space"},
		{"same", "same"},
	}

	for _, test := range cases {
		was, now := test[0], test[1]
		parts := publish.WordDiff(was, now)

		if got := rebuilt(parts, publish.Gone); was != "" && got != was && len(parts) > 0 {
			t.Errorf("WordDiff(%q, %q): reading the removals back gives %q, want %q", was, now, got, was)
		}
		if got := rebuilt(parts, publish.Added); now != "" && got != now && len(parts) > 0 {
			t.Errorf("WordDiff(%q, %q): reading the additions back gives %q, want %q", was, now, got, now)
		}
	}
}

func TestTextThatDidNotChangeHasNoDiff(t *testing.T) {
	if parts := publish.WordDiff("the same words", "the same words"); parts != nil {
		t.Errorf("WordDiff() = %v, want nothing when nothing changed", parts)
	}
}

func TestOnlyTheChangedWordsAreMarked(t *testing.T) {
	parts := publish.WordDiff("The quick brown fox jumps", "The quick red fox jumps")

	gone, added := "", ""
	for _, part := range parts {
		switch part.Part {
		case publish.Gone:
			gone += part.Text
		case publish.Added:
			added += part.Text
		}
	}

	if strings.TrimSpace(gone) != "brown" {
		t.Errorf("removed %q, want just the word that went", gone)
	}
	if strings.TrimSpace(added) != "red" {
		t.Errorf("added %q, want just the word that arrived", added)
	}
}

func TestNeighbouringChangesAreOnePhrase(t *testing.T) {
	parts := publish.WordDiff("We sell apples and pears", "We sell plums and pears")

	runs := 0
	for _, part := range parts {
		if part.Part != publish.Kept {
			runs++
		}
	}
	if runs > 2 {
		t.Errorf("WordDiff() = %v, want the change as one removal and one addition", parts)
	}
}

func TestAVeryLongTextIsNotComparedWordByWord(t *testing.T) {
	was := strings.Repeat("word ", 5000)
	now := strings.Repeat("word ", 5000) + "and one more"

	parts := publish.WordDiff(was, now)

	if len(parts) != 2 {
		t.Fatalf("WordDiff() gave %d parts, want the whole text replaced", len(parts))
	}
	if parts[0].Part != publish.Gone || parts[1].Part != publish.Added {
		t.Errorf("WordDiff() = %v, want one removal and one addition", parts)
	}
}

func TestReadableTurnsAKeyIntoWords(t *testing.T) {
	cases := map[string]string{
		"heroTitle":      "Hero title",
		"hero_title":     "Hero title",
		"hero-title":     "Hero title",
		"title":          "Title",
		"someLongerName": "Some longer name",
		"":               "",
	}

	for key, want := range cases {
		if got := publish.Readable(key); got != want {
			t.Errorf("Readable(%q) = %q, want %q", key, got, want)
		}
	}
}
