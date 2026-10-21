package model

// Model is a site's assembled Quiqr model.
type Model struct {
	SSGType     string         `json:"ssgType"`
	SSGVersion  string         `json:"ssgVersion,omitempty"`
	Singles     []Single       `json:"singles,omitempty"`
	Collections []Collection   `json:"collections,omitempty"`
	Dynamics    []any          `json:"dynamics,omitempty"`
	Menu        []MenuSection  `json:"menu,omitempty"`
	Build       []NamedConfig  `json:"build,omitempty"`
	Serve       []NamedConfig  `json:"serve,omitempty"`
	Extra       map[string]any `json:"extra,omitempty"`

	BaseFile string    `json:"baseFile"`
	Warnings []Warning `json:"warnings,omitempty"`
}

type NamedConfig struct {
	Key    string         `json:"key"`
	Config string         `json:"config,omitempty"`
	Extra  map[string]any `json:"extra,omitempty"`
}

type Single struct {
	Key                  string         `json:"key"`
	Title                string         `json:"title,omitempty"`
	File                 string         `json:"file,omitempty"`
	DataFormat           string         `json:"dataformat,omitempty"`
	PreviewURL           string         `json:"previewUrl,omitempty"`
	Description          string         `json:"description,omitempty"`
	PullOuterRootKey     string         `json:"pullOuterRootKey,omitempty"`
	Hidden               bool           `json:"hidden,omitempty"`
	HidePreviewIcon      bool           `json:"hidePreviewIcon,omitempty"`
	HideExternalEditIcon bool           `json:"hideExternalEditIcon,omitempty"`
	HideSaveButton       bool           `json:"hideSaveButton,omitempty"`
	Fields               []Field        `json:"fields,omitempty"`
	Extra                map[string]any `json:"extra,omitempty"`
}

type Collection struct {
	Key             string         `json:"key"`
	Title           string         `json:"title,omitempty"`
	Folder          string         `json:"folder"`
	Extension       string         `json:"extension,omitempty"`
	DataFormat      string         `json:"dataformat,omitempty"`
	ItemTitle       string         `json:"itemtitle,omitempty"`
	SortKey         string         `json:"sortkey,omitempty"`
	Hidden          bool           `json:"hidden,omitempty"`
	HideIndex       bool           `json:"hideIndex,omitempty"`
	PreviewURLBase  string         `json:"previewUrlBase,omitempty"`
	HidePreviewIcon bool           `json:"hidePreviewIcon,omitempty"`
	Fields          []Field        `json:"fields,omitempty"`
	Extra           map[string]any `json:"extra,omitempty"`
}

type MenuSection struct {
	Key       string         `json:"key,omitempty"`
	Title     string         `json:"title"`
	MatchRole string         `json:"matchRole,omitempty"`
	Disabled  bool           `json:"disabled,omitempty"`
	MenuItems []MenuItem     `json:"menuItems,omitempty"`
	Extra     map[string]any `json:"extra,omitempty"`
}

type MenuItem struct {
	Key   string         `json:"key"`
	Extra map[string]any `json:"extra,omitempty"`
}

type Field struct {
	Key                  string         `json:"key"`
	Type                 string         `json:"type"`
	Title                string         `json:"title,omitempty"`
	Tip                  string         `json:"tip,omitempty"`
	Name                 string         `json:"name,omitempty"`
	Group                string         `json:"group,omitempty"`
	Content              string         `json:"content,omitempty"`
	Theme                string         `json:"theme,omitempty"`
	Default              any            `json:"default,omitempty"`
	Hidden               bool           `json:"hidden,omitempty"`
	GroupData            bool           `json:"groupdata,omitempty"`
	ArrayTitle           bool           `json:"arrayTitle,omitempty"`
	AutoSave             bool           `json:"autoSave,omitempty"`
	FieldPromptTemplates []string       `json:"field_prompt_templates,omitempty"`
	Fields               []Field        `json:"fields,omitempty"`
	Config               FieldConfig    `json:"config,omitempty"`
	Extra                map[string]any `json:"extra,omitempty"`
}

// Single returns the single with this key.
func (m *Model) Single(key string) (Single, bool) {
	for _, single := range m.Singles {
		if single.Key == key {
			return single, true
		}
	}
	return Single{}, false
}

// Collection returns the collection with this key.
func (m *Model) Collection(key string) (Collection, bool) {
	for _, collection := range m.Collections {
		if collection.Key == key {
			return collection, true
		}
	}
	return Collection{}, false
}
