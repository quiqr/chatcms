package publish

import (
	"fmt"
	"sort"
	"strings"

	"github.com/quiqr/chatcms/internal/frontmatter"
	"github.com/quiqr/chatcms/internal/model"
)

// FieldChange is one field of one page that differs from the published site.
type FieldChange struct {
	Name  string `json:"name"`
	Was   string `json:"was,omitempty"`
	Now   string `json:"now,omitempty"`
	Words []Word `json:"words,omitempty"`
}

// BodyField is what the page's own text is called. It is not a field in the
// model, so it needs a name of its own that reads like one.
const BodyField = "Page text"

// Compare says which fields of a page differ, naming each one the way the
// model names it.
func Compare(fields []model.Field, was, now *frontmatter.Document) []FieldChange {
	names := namesOf(fields, nil, map[string]string{})

	changes := []FieldChange{}
	for _, path := range union(was, now) {
		before, hadBefore := valueAt(was, path)
		after, hasAfter := valueAt(now, path)
		if !hadBefore && !hasAfter {
			continue
		}
		if before == after {
			continue
		}
		changes = append(changes, FieldChange{
			Name:  nameFor(names, path),
			Was:   before,
			Now:   after,
			Words: WordDiff(before, after),
		})
	}

	if bodyOf(was) != bodyOf(now) {
		changes = append(changes, FieldChange{
			Name:  BodyField,
			Was:   bodyOf(was),
			Now:   bodyOf(now),
			Words: WordDiff(bodyOf(was), bodyOf(now)),
		})
	}
	return changes
}

func bodyOf(document *frontmatter.Document) string {
	if document == nil {
		return ""
	}
	return document.Body
}

// namesOf walks the model's fields and remembers what each one is called, by
// the path its data sits at. Nested fields keep their parent's name in front
// so "Address, Street" reads as one thing.
func namesOf(fields []model.Field, prefix []string, into map[string]string) map[string]string {
	for _, field := range fields {
		if field.Key == "" {
			continue
		}
		here := append(append([]string{}, prefix...), field.Key)
		title := field.Title
		if title == "" {
			title = Readable(field.Key)
		}
		if parent, nested := into[strings.Join(prefix, ".")]; nested && parent != "" {
			title = parent + ", " + title
		}
		into[strings.Join(here, ".")] = title
		if len(field.Fields) > 0 {
			namesOf(field.Fields, here, into)
		}
	}
	return into
}

func nameFor(names map[string]string, path string) string {
	if title, known := names[path]; known {
		return title
	}
	// A list item's path carries its position; the field it belongs to is
	// what a person recognises, with the position said in words.
	parts := strings.Split(path, ".")
	for cut := len(parts) - 1; cut > 0; cut-- {
		if title, known := names[strings.Join(parts[:cut], ".")]; known {
			return title + " (" + strings.Join(parts[cut:], ", ") + ")"
		}
	}
	return Readable(strings.ReplaceAll(path, ".", " "))
}

func union(was, now *frontmatter.Document) []string {
	seen := map[string]bool{}
	for _, path := range paths(was) {
		seen[path] = true
	}
	for _, path := range paths(now) {
		seen[path] = true
	}
	out := make([]string, 0, len(seen))
	for path := range seen {
		out = append(out, path)
	}
	sort.Strings(out)
	return out
}

func paths(document *frontmatter.Document) []string {
	if document == nil || document.Fields == nil {
		return nil
	}
	return walk(document.Fields, nil, nil)
}

func walk(value any, at []string, into []string) []string {
	switch held := value.(type) {
	case *frontmatter.Map:
		if held == nil {
			return into
		}
		for _, key := range held.Keys() {
			inner, _ := held.Get(key)
			into = walk(inner, append(at, key), into)
		}
		return into
	case []any:
		for index, inner := range held {
			into = walk(inner, append(at, fmt.Sprintf("%d", index+1)), into)
		}
		return into
	default:
		if len(at) == 0 {
			return into
		}
		return append(into, strings.Join(at, "."))
	}
}

func valueAt(document *frontmatter.Document, path string) (string, bool) {
	if document == nil || document.Fields == nil {
		return "", false
	}

	var value any = document.Fields
	for _, step := range strings.Split(path, ".") {
		switch held := value.(type) {
		case *frontmatter.Map:
			inner, found := held.Get(step)
			if !found {
				return "", false
			}
			value = inner
		case []any:
			index := 0
			if _, err := fmt.Sscanf(step, "%d", &index); err != nil || index < 1 || index > len(held) {
				return "", false
			}
			value = held[index-1]
		default:
			return "", false
		}
	}
	return say(value), true
}

func say(value any) string {
	switch held := value.(type) {
	case nil:
		return ""
	case string:
		return held
	case bool:
		if held {
			return "yes"
		}
		return "no"
	default:
		return fmt.Sprintf("%v", held)
	}
}
