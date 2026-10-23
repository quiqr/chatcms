package content

import (
	"context"
	"io/fs"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/model"
)

// Section is a group in the content browser: the site's pages, or one
// collection of them.
type Section struct {
	Key   string
	Title string
	// ItemName is the singular word for one of its items, used in "New
	// article". It is empty for the pages section.
	ItemName string
	Count    int
	Items    []Item
}

// Item is one page in the browser.
type Item struct {
	Key   string
	Title string
	// Waiting says this page has a change that has not been published.
	Waiting bool
	By      string
	At      time.Time
}

// Browse lists what a content creator can open, in the order the model asks
// for, with a mark on anything not yet published.
func Browse(ctx context.Context, site Site, waiting map[string]changeBy) []Section {
	m, err := site.Model()
	if err != nil {
		return nil
	}

	sections := []Section{}

	pages := Section{Key: "pages", Title: "Pages"}
	for _, single := range m.Singles {
		if single.Hidden {
			continue
		}
		pages.Items = append(pages.Items, itemOf(single.Key, titleOr(single.Title, single.Key), single.File, waiting))
	}
	pages.Count = len(pages.Items)
	if pages.Count > 0 {
		sections = append(sections, pages)
	}

	for _, collection := range m.Collections {
		if collection.Hidden {
			continue
		}
		section := Section{
			Key:      collection.Key,
			Title:    titleOr(collection.Title, collection.Key),
			ItemName: itemNameOf(collection),
		}
		for _, name := range itemsIn(site.FS(), collection) {
			key := collection.Key + "/" + strings.TrimSuffix(name, path.Ext(name))
			file := path.Join(strings.TrimSuffix(collection.Folder, "/"), name)
			section.Items = append(section.Items, itemOf(key, readable(name), file, waiting))
		}
		section.Count = len(section.Items)
		sections = append(sections, section)
	}
	return sections
}

func itemNameOf(collection model.Collection) string {
	if collection.ItemTitle != "" {
		return collection.ItemTitle
	}
	return model.DefaultItemName
}

func itemOf(key, title, file string, waiting map[string]changeBy) Item {
	item := Item{Key: key, Title: title}
	if change, found := waiting[file]; found {
		item.Waiting = true
		item.By = change.By
		item.At = change.At
	}
	return item
}

// itemsIn lists a collection's files, most recently named first by name so
// that the order is the same every time.
func itemsIn(fsys fs.FS, collection model.Collection) []string {
	folder := strings.TrimSuffix(collection.Folder, "/")
	if folder == "" {
		return nil
	}

	entries, err := fs.ReadDir(fsys, folder)
	if err != nil {
		return nil
	}

	names := []string{}
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if collection.HideIndex && isIndex(entry.Name()) {
			continue
		}
		if !editableFile(entry.Name()) {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names
}

func isIndex(name string) bool {
	stem := strings.TrimSuffix(name, path.Ext(name))
	return stem == "index" || stem == "_index"
}

func editableFile(name string) bool {
	switch strings.ToLower(path.Ext(name)) {
	case ".md", ".markdown", ".html", ".yaml", ".yml", ".toml", ".json":
		return true
	default:
		return false
	}
}

// changeBy is who last changed a file and when, for the mark on an item.
type changeBy struct {
	By string
	At time.Time
}

func waitingFrom(changes []gitsync.Change) map[string]changeBy {
	waiting := map[string]changeBy{}
	for _, change := range changes {
		waiting[change.Path] = changeBy{By: change.By.Name, At: change.At}
	}
	return waiting
}
