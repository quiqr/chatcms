package publish

import (
	"path"
	"strings"
	"unicode"

	"github.com/quiqr/chatcms/internal/model"
)

// Kind is what sort of thing changed, in the words the designs use for the
// badge beside each change.
type Kind string

const (
	ChangedPage  Kind = "changed-page"
	NewPage      Kind = "new-page"
	RemovedPage  Kind = "removed-page"
	NewImage     Kind = "new-image"
	RemovedImage Kind = "removed-image"
	OtherFile    Kind = "other-file"
)

// Place is where in the site a changed file sits, said the way a content
// creator would say it rather than as a path.
type Place struct {
	Section string
	Page    string
	Fields  []model.Field
	Known   bool
}

// Locate works out which page of the site a file is, using the model. A file
// the model says nothing about is still reported, named by its own path,
// because a change nobody can see is worse than one that is named plainly.
func Locate(m *model.Model, name string) Place {
	if m != nil {
		for _, single := range m.Singles {
			if single.File == name {
				return Place{Section: "Pages", Page: titleOr(single.Title, single.Key), Fields: single.Fields, Known: true}
			}
		}
		for _, collection := range m.Collections {
			if within(collection.Folder, name) {
				return Place{
					Section: titleOr(collection.Title, collection.Key),
					Page:    readableName(name),
					Fields:  collection.Fields,
					Known:   true,
				}
			}
		}
	}
	return Place{Section: sectionOf(name), Page: readableName(name), Known: false}
}

func within(folder, name string) bool {
	if folder == "" {
		return false
	}
	folder = strings.TrimSuffix(folder, "/")
	return strings.HasPrefix(name, folder+"/")
}

func titleOr(title, key string) string {
	if title != "" {
		return title
	}
	return Readable(key)
}

func sectionOf(name string) string {
	switch {
	case isImage(name):
		return "Images"
	case strings.HasPrefix(name, "content/"):
		return "Pages"
	default:
		return "Other files"
	}
}

// readableName turns a path into something a person can recognise: the page
// name a Hugo site would use, which is the folder for an index file and the
// file name otherwise.
func readableName(name string) string {
	base := path.Base(name)
	stem := strings.TrimSuffix(base, path.Ext(base))
	if stem == "index" || stem == "_index" {
		if parent := path.Base(path.Dir(name)); parent != "." && parent != "/" {
			stem = parent
		}
	}
	return Readable(stem)
}

// Readable turns a key into words. Quiqr models are written in camelCase,
// kebab-case and snake_case alike, and an editor should see none of them.
func Readable(key string) string {
	if key == "" {
		return ""
	}

	spaced := strings.Builder{}
	previous := rune(0)
	for _, letter := range key {
		switch {
		case letter == '-' || letter == '_' || letter == '.':
			spaced.WriteRune(' ')
		case unicode.IsUpper(letter) && unicode.IsLower(previous):
			spaced.WriteRune(' ')
			spaced.WriteRune(unicode.ToLower(letter))
		default:
			spaced.WriteRune(letter)
		}
		previous = letter
	}

	words := strings.Fields(spaced.String())
	if len(words) == 0 {
		return key
	}
	words[0] = string(unicode.ToUpper(rune(words[0][0]))) + words[0][1:]
	return strings.Join(words, " ")
}

var imageExtensions = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true,
	".svg": true, ".avif": true, ".bmp": true, ".ico": true, ".tiff": true,
}

func isImage(name string) bool {
	return imageExtensions[strings.ToLower(path.Ext(name))]
}

func isContent(name string) bool {
	switch strings.ToLower(path.Ext(name)) {
	case ".md", ".markdown", ".html", ".yaml", ".yml", ".toml", ".json":
		return strings.HasPrefix(name, "content/") || strings.HasPrefix(name, "data/")
	default:
		return false
	}
}
