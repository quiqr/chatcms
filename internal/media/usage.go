package media

import (
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/quiqr/chatcms/internal/model"
)

// UsedOn finds the pages that refer to a picture. A site writes a picture's
// address in more than one shape, so every shape a Hugo site uses is looked
// for rather than only the one ChatCMS would have written.
func UsedOn(fsys fs.FS, m *model.Model, folder, name string) []string {
	if m == nil || name == "" {
		return nil
	}

	wanted := shapesOf(folder, name)
	pages := map[string]bool{}

	for _, single := range m.Singles {
		if single.File != "" && refers(fsys, single.File, wanted) {
			pages[single.Key] = true
		}
	}

	for _, collection := range m.Collections {
		holder := strings.TrimSuffix(collection.Folder, "/")
		if holder == "" {
			continue
		}
		entries, err := fs.ReadDir(fsys, holder)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if refers(fsys, path.Join(holder, entry.Name()), wanted) {
				stem := strings.TrimSuffix(entry.Name(), path.Ext(entry.Name()))
				pages[collection.Key+"/"+stem] = true
			}
		}
	}

	found := make([]string, 0, len(pages))
	for key := range pages {
		found = append(found, key)
	}
	sort.Strings(found)
	return found
}

// shapesOf is every way a site might write one picture's address: as Hugo
// serves it, as it sits in the repository, and bare.
func shapesOf(folder, name string) []string {
	served := strings.TrimPrefix(folder, "static/")
	return []string{
		"/" + path.Join(served, name),
		path.Join(served, name),
		path.Join(folder, name),
		"/" + path.Join(folder, name),
	}
}

func refers(fsys fs.FS, file string, wanted []string) bool {
	content, err := fs.ReadFile(fsys, file)
	if err != nil {
		return false
	}
	said := string(content)
	for _, shape := range wanted {
		if strings.Contains(said, shape) {
			return true
		}
	}
	return false
}
