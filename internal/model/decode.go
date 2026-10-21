package model

import "fmt"

// Decode turns an assembled generic tree into the typed model.
func Decode(assembly *Assembly) (*Model, error) {
	root := newReader(assembly.BaseFile, "", assembly.Tree)

	m := &Model{
		BaseFile: assembly.BaseFile,
		Warnings: assembly.Warnings,
	}

	m.SSGType = root.str("ssgType")
	m.SSGVersion = root.str("ssgVersion")

	if legacy, ok := assembly.Tree["hugover"]; ok {
		version, ok := legacy.(string)
		if !ok {
			return nil, at(assembly.BaseFile, "hugover", "a version string")
		}
		if m.SSGType == "" {
			m.SSGType = "hugo"
		}
		if m.SSGVersion == "" {
			m.SSGVersion = version
		}
	}
	if m.SSGType == "" {
		m.SSGType = "hugo"
	}

	for _, entry := range root.mapList("singles") {
		single, err := decodeSingle(assembly.BaseFile, len(m.Singles), entry)
		if err != nil {
			return nil, err
		}
		m.Singles = append(m.Singles, single)
	}

	for _, entry := range root.mapList("collections") {
		collection, err := decodeCollection(assembly.BaseFile, len(m.Collections), entry)
		if err != nil {
			return nil, err
		}
		m.Collections = append(m.Collections, collection)
	}

	m.Dynamics = root.list("dynamics")

	for index, entry := range root.mapList("menu") {
		section, err := decodeMenuSection(assembly.BaseFile, index, entry)
		if err != nil {
			return nil, err
		}
		m.Menu = append(m.Menu, section)
	}

	for index, entry := range root.mapList("build") {
		m.Build = append(m.Build, decodeNamedConfig(assembly.BaseFile, "build", index, entry))
	}
	for index, entry := range root.mapList("serve") {
		m.Serve = append(m.Serve, decodeNamedConfig(assembly.BaseFile, "serve", index, entry))
	}

	if root.err != nil {
		return nil, root.err
	}
	m.Extra = root.extra()

	if err := rejectDuplicateKeys(assembly.BaseFile, m); err != nil {
		return nil, err
	}
	return m, nil
}

func rejectDuplicateKeys(file string, m *Model) error {
	seen := map[string]bool{}
	for _, single := range m.Singles {
		if seen[single.Key] {
			return atf(file, "singles", "two singles share the key %q", single.Key)
		}
		seen[single.Key] = true
	}

	seen = map[string]bool{}
	for _, collection := range m.Collections {
		if seen[collection.Key] {
			return atf(file, "collections", "two collections share the key %q", collection.Key)
		}
		seen[collection.Key] = true
	}
	return nil
}

func decodeSingle(file string, index int, tree map[string]any) (Single, error) {
	where := pathOf("singles", index)
	r := newReader(file, where, tree)

	single := Single{
		Key:                  r.requiredStr("key"),
		Title:                r.str("title"),
		File:                 r.str("file"),
		DataFormat:           r.str("dataformat"),
		PreviewURL:           r.str("previewUrl"),
		Description:          r.str("description"),
		PullOuterRootKey:     r.str("pullOuterRootKey"),
		Hidden:               r.boolean("hidden"),
		HidePreviewIcon:      r.boolean("hidePreviewIcon"),
		HideExternalEditIcon: r.boolean("hideExternalEditIcon"),
		HideSaveButton:       r.boolean("hideSaveButton"),
	}

	fields, err := decodeFields(file, join(where, "fields"), r.list("fields"))
	if err != nil {
		return Single{}, err
	}
	single.Fields = fields

	if r.err != nil {
		return Single{}, r.err
	}
	single.Extra = r.extra()
	return single, nil
}

func decodeCollection(file string, index int, tree map[string]any) (Collection, error) {
	where := pathOf("collections", index)
	r := newReader(file, where, tree)

	collection := Collection{
		Key:             r.requiredStr("key"),
		Title:           r.str("title"),
		Folder:          r.requiredStr("folder"),
		Extension:       r.str("extension"),
		DataFormat:      r.str("dataformat"),
		ItemTitle:       r.str("itemtitle"),
		SortKey:         r.str("sortkey"),
		Hidden:          r.boolean("hidden"),
		HideIndex:       r.boolean("hideIndex"),
		PreviewURLBase:  r.str("previewUrlBase"),
		HidePreviewIcon: r.boolean("hidePreviewIcon"),
	}

	fields, err := decodeFields(file, join(where, "fields"), r.list("fields"))
	if err != nil {
		return Collection{}, err
	}
	collection.Fields = fields

	if r.err != nil {
		return Collection{}, r.err
	}
	collection.Extra = r.extra()
	return collection, nil
}

func decodeMenuSection(file string, index int, tree map[string]any) (MenuSection, error) {
	where := pathOf("menu", index)
	r := newReader(file, where, tree)

	section := MenuSection{
		Key:       r.str("key"),
		Title:     r.str("title"),
		MatchRole: r.str("matchRole"),
		Disabled:  r.boolean("disabled"),
	}

	for itemIndex, item := range r.mapList("menuItems") {
		itemReader := newReader(file, fmt.Sprintf("%s.menuItems[%d]", where, itemIndex), item)
		menuItem := MenuItem{Key: itemReader.requiredStr("key")}
		if itemReader.err != nil {
			return MenuSection{}, itemReader.err
		}
		menuItem.Extra = itemReader.extra()
		section.MenuItems = append(section.MenuItems, menuItem)
	}

	if r.err != nil {
		return MenuSection{}, r.err
	}
	section.Extra = r.extra()
	return section, nil
}

func decodeNamedConfig(file, key string, index int, tree map[string]any) NamedConfig {
	r := newReader(file, pathOf(key, index), tree)
	config := NamedConfig{Key: r.str("key"), Config: r.str("config")}
	config.Extra = r.extra()
	return config
}
