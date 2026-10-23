// Package sitesync turns what the repository engine reports about a site into
// something a content creator can act on: whether everything is saved, what
// arrived from somebody else, and, when two people changed the same thing,
// which of the two to keep.
package sitesync

import (
	"fmt"
	"sort"
	"strings"

	"github.com/quiqr/chatcms/internal/content"
	"github.com/quiqr/chatcms/internal/forms"
	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/model"
	"github.com/quiqr/chatcms/internal/publish"
)

// FieldConflict is one field two people changed, with both answers.
type FieldConflict struct {
	// Path is the field as the model addresses it, which is what a choice
	// names.
	Path string
	Name string
	// Mine is what the person editing here wrote; Theirs is what arrived.
	Mine   string
	Theirs string
	// Words compares the two, so that "Compare" can show what differs
	// rather than two blocks of text.
	Words []publish.Word
}

// PageConflict is one page two people changed, said as pages and fields.
type PageConflict struct {
	// Page is the key the content browser uses, so this can be a link.
	Page string
	// Title is what the page is called.
	Title string
	// File is where it lives. It is what a resolution is written back to and
	// is never shown.
	File   string
	Who    string
	Fields []FieldConflict
	// WholeFile says the two versions differ in ways this cannot put as
	// fields, so the choice is between one whole version and the other.
	WholeFile bool
}

// BodyField is what a page's own text is called when it is the thing two
// people changed.
const BodyField = publish.BodyField

// Describe turns the files a sync could not reconcile into pages and fields.
// A file it cannot read as content is still reported, as a whole-file choice,
// because a change nobody can act on is worse than a blunt one.
func Describe(conflict *gitsync.Conflict, m *model.Model, who string) []PageConflict {
	if conflict == nil {
		return nil
	}

	pages := []PageConflict{}
	for _, path := range conflict.Paths {
		pages = append(pages, describeOne(path, m, who))
	}
	sort.Slice(pages, func(a, b int) bool { return pages[a].File < pages[b].File })
	return pages
}

func describeOne(path gitsync.ConflictedPath, m *model.Model, who string) PageConflict {
	place := publish.Locate(m, path.Path)
	page := PageConflict{
		Page:  keyOf(m, path.Path),
		Title: place.Page,
		File:  path.Path,
		Who:   who,
	}

	// Only a file that holds words is worth asking about field by field. A
	// picture parses as a document with everything in its body, and showing
	// that to a person would be a screen of mojibake.
	if !place.Known && !publish.IsContent(path.Path) {
		page.WholeFile = true
		return page
	}

	mine, mineErr := forms.Load(place.Fields, path.Path, path.Local)
	theirs, theirsErr := forms.Load(place.Fields, path.Path, path.Remote)
	if mineErr != nil || theirsErr != nil {
		page.WholeFile = true
		return page
	}

	for _, field := range differing(mine, theirs, place.Fields) {
		page.Fields = append(page.Fields, field)
	}
	if len(page.Fields) == 0 {
		// The two differ somewhere this cannot name: a key the model does not
		// declare, or spacing. Choosing a whole version is then the only
		// honest offer.
		page.WholeFile = true
	}
	return page
}

// keyOf is how the content browser addresses a file, so that a conflict can
// link to the page it is about.
func keyOf(m *model.Model, file string) string {
	if m == nil {
		return ""
	}
	for _, single := range m.Singles {
		if single.File == file {
			return single.Key
		}
	}
	for _, collection := range m.Collections {
		folder := strings.TrimSuffix(collection.Folder, "/")
		if folder != "" && strings.HasPrefix(file, folder+"/") {
			name := strings.TrimPrefix(file, folder+"/")
			return collection.Key + "/" + strings.TrimSuffix(name, extensionOf(name))
		}
	}
	return ""
}

func extensionOf(name string) string {
	if at := strings.LastIndex(name, "."); at > 0 {
		return name[at:]
	}
	return ""
}

func differing(mine, theirs *forms.Form, fields []model.Field) []FieldConflict {
	found := []FieldConflict{}

	for _, binding := range mine.Bindings() {
		ours := said(mine, binding.FieldPath)
		yours := said(theirs, binding.FieldPath)
		if ours == yours {
			continue
		}
		found = append(found, FieldConflict{
			Path:   binding.FieldPath,
			Name:   nameOf(fields, binding.FieldPath),
			Mine:   ours,
			Theirs: yours,
			Words:  publish.WordDiff(ours, yours),
		})
	}

	if mine.Body() != theirs.Body() {
		found = append(found, FieldConflict{
			Path:   forms.BodyKey,
			Name:   BodyField,
			Mine:   mine.Body(),
			Theirs: theirs.Body(),
			Words:  publish.WordDiff(mine.Body(), theirs.Body()),
		})
	}
	return found
}

func said(form *forms.Form, path string) string {
	held, found := form.Get(path)
	if !found {
		return ""
	}
	return plainly(content.Plain(held))
}

func plainly(value any) string {
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
	case []any:
		parts := make([]string, 0, len(held))
		for _, item := range held {
			parts = append(parts, plainly(item))
		}
		return strings.Join(parts, ", ")
	default:
		return fmt.Sprintf("%v", held)
	}
}

func nameOf(fields []model.Field, path string) string {
	parts := strings.Split(path, ".")
	named := fields
	title := ""

	for _, part := range parts {
		for _, field := range named {
			if field.Key != part {
				continue
			}
			said := field.Title
			if said == "" {
				said = publish.Readable(field.Key)
			}
			if title == "" {
				title = said
			} else {
				title += ", " + said
			}
			named = field.Fields
			break
		}
	}

	if title == "" {
		return publish.Readable(strings.ReplaceAll(path, ".", " "))
	}
	return title
}
