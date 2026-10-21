package model

import (
	"io/fs"
	"path"
	"strings"
)

// Navigation is what a content creator browses: the pages and collections of a
// site, in the order the model asks for.
type Navigation struct {
	Sections []NavSection `json:"sections"`
	Warnings []Warning    `json:"warnings,omitempty"`
}

type NavSection struct {
	Key     string     `json:"key,omitempty"`
	Title   string     `json:"title"`
	Entries []NavEntry `json:"entries"`
}

type NavEntry struct {
	Key   string `json:"key"`
	Title string `json:"title"`
	Kind  string `json:"kind"`
	// ItemName is the singular name used in labels such as "New article". It
	// is empty for a page.
	ItemName string `json:"itemName,omitempty"`
	// Count is how many items a collection holds. It is nil for a page.
	Count *int `json:"count,omitempty"`
}

const (
	KindPage       = "page"
	KindCollection = "collection"

	// DefaultItemName is what a collection is called when its model gives it
	// no singular name. It avoids both git vocabulary and developer
	// vocabulary.
	DefaultItemName = "item"

	sectionPages       = "Pages"
	sectionCollections = "Collections"
)

// Navigate derives the navigation from a model. Item counts are read through
// fsys; pass nil to build the navigation without counting.
func (m *Model) Navigate(fsys fs.FS) Navigation {
	navigation := Navigation{}
	counts := m.countItems(fsys)

	if len(m.Menu) > 0 {
		navigation.Sections, navigation.Warnings = m.menuSections(counts)
		return navigation
	}

	pages := NavSection{Title: sectionPages}
	for _, single := range m.Singles {
		if single.Hidden {
			continue
		}
		pages.Entries = append(pages.Entries, m.pageEntry(single))
	}

	collections := NavSection{Title: sectionCollections}
	for _, collection := range m.Collections {
		if collection.Hidden {
			continue
		}
		collections.Entries = append(collections.Entries, m.collectionEntry(collection, counts))
	}

	if len(pages.Entries) > 0 {
		navigation.Sections = append(navigation.Sections, pages)
	}
	if len(collections.Entries) > 0 {
		navigation.Sections = append(navigation.Sections, collections)
	}
	return navigation
}

func (m *Model) menuSections(counts map[string]int) ([]NavSection, []Warning) {
	var sections []NavSection
	var warnings []Warning

	for _, declared := range m.Menu {
		if declared.Disabled {
			continue
		}

		section := NavSection{Key: declared.Key, Title: declared.Title}
		for _, item := range declared.MenuItems {
			if single, ok := m.Single(item.Key); ok {
				if single.Hidden {
					continue
				}
				section.Entries = append(section.Entries, m.pageEntry(single))
				continue
			}
			if collection, ok := m.Collection(item.Key); ok {
				if collection.Hidden {
					continue
				}
				section.Entries = append(section.Entries, m.collectionEntry(collection, counts))
				continue
			}
			warnings = append(warnings, Warning{
				File:    m.BaseFile,
				Path:    "menu." + declared.Title,
				Message: "the menu names \"" + item.Key + "\", which this model does not declare",
			})
		}
		sections = append(sections, section)
	}
	return sections, warnings
}

func (m *Model) pageEntry(single Single) NavEntry {
	return NavEntry{
		Key:   single.Key,
		Title: titleOr(single.Title, single.Key),
		Kind:  KindPage,
	}
}

func (m *Model) collectionEntry(collection Collection, counts map[string]int) NavEntry {
	count := counts[collection.Key]
	return NavEntry{
		Key:      collection.Key,
		Title:    titleOr(collection.Title, collection.Key),
		Kind:     KindCollection,
		ItemName: titleOr(collection.ItemTitle, DefaultItemName),
		Count:    &count,
	}
}

func (m *Model) countItems(fsys fs.FS) map[string]int {
	counts := make(map[string]int, len(m.Collections))
	for _, collection := range m.Collections {
		counts[collection.Key] = m.countCollection(fsys, collection)
	}
	return counts
}

func (m *Model) countCollection(fsys fs.FS, collection Collection) int {
	if fsys == nil || collection.Folder == "" {
		return 0
	}

	entries, err := fs.ReadDir(fsys, path.Clean(strings.TrimSuffix(collection.Folder, "/")))
	if err != nil {
		return 0
	}

	extension := collection.Extension
	if extension != "" && !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if extension != "" && !strings.HasSuffix(name, extension) {
			continue
		}
		if collection.HideIndex && isIndexFile(name) {
			continue
		}
		count++
	}
	return count
}

func isIndexFile(name string) bool {
	base := strings.TrimSuffix(name, path.Ext(name))
	return base == "index" || base == "_index"
}

func titleOr(title, fallback string) string {
	if title != "" {
		return title
	}
	return fallback
}
