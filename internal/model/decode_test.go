package model

import (
	"strings"
	"testing"
	"testing/fstest"
)

func load(t *testing.T, files map[string]string) *Model {
	t.Helper()

	m, err := Load(site(files))
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	return m
}

func withFields(fields string) map[string]string {
	return map[string]string{
		"quiqr/model/base.yaml": "singles:\n  - key: home\n    file: content/_index.md\n    fields:\n" + fields,
	}
}

func TestRootIsTyped(t *testing.T) {
	m := load(t, map[string]string{
		"quiqr/model/base.yaml": `ssgType: hugo
ssgVersion: 0.110.0
singles:
  - key: home
    title: Homepage
    file: content/_index.md
collections:
  - key: posts
    title: Posts
    folder: content/posts/
    extension: md
    dataformat: yaml
    itemtitle: article
menu:
  - title: Site
    menuItems:
      - key: home
build:
  - key: default
    config: config.toml
serve:
  - key: preview
    config: config.toml
`,
	})

	if m.SSGType != "hugo" || m.SSGVersion != "0.110.0" {
		t.Errorf("generator = %s %s, want hugo 0.110.0", m.SSGType, m.SSGVersion)
	}
	if len(m.Singles) != 1 || m.Singles[0].File != "content/_index.md" {
		t.Errorf("singles = %+v", m.Singles)
	}
	if len(m.Collections) != 1 || m.Collections[0].ItemTitle != "article" {
		t.Errorf("collections = %+v", m.Collections)
	}
	if len(m.Menu) != 1 || len(m.Menu[0].MenuItems) != 1 {
		t.Errorf("menu = %+v", m.Menu)
	}
	if len(m.Build) != 1 || len(m.Serve) != 1 {
		t.Errorf("build = %+v, serve = %+v", m.Build, m.Serve)
	}
}

func TestHugoverIsMigratedAndPreserved(t *testing.T) {
	m := load(t, map[string]string{
		"quiqr/model/base.yaml": "hugover: 0.110.0\nsingles:\n  - key: home\n",
	})

	if m.SSGType != "hugo" {
		t.Errorf("ssgType = %q, want hugo", m.SSGType)
	}
	if m.SSGVersion != "0.110.0" {
		t.Errorf("ssgVersion = %q, want the hugover value", m.SSGVersion)
	}
	if m.Extra["hugover"] != "0.110.0" {
		t.Errorf("extra = %v, want the original hugover key preserved", m.Extra)
	}
}

func TestExplicitGeneratorWinsOverHugover(t *testing.T) {
	m := load(t, map[string]string{
		"quiqr/model/base.yaml": "hugover: 0.110.0\nssgVersion: 0.120.0\nsingles:\n  - key: home\n",
	})

	if m.SSGVersion != "0.120.0" {
		t.Errorf("ssgVersion = %q, want the explicit key to win", m.SSGVersion)
	}
}

func TestEveryFieldTypeIsTyped(t *testing.T) {
	// One minimal declaration per type in docs/QUIQR-MODEL.md. A type missing
	// from this table fails the test below, so a new type cannot be added to
	// KnownFieldTypes without being exercised here.
	declarations := map[string]string{
		"string":                 "    multiLine: true\n",
		"markdown":               "",
		"easymde":                "",
		"readonly":               "",
		"uniq":                   "",
		"number":                 "    min: 1\n    max: 10\n",
		"slider":                 "    min: 0\n    max: 100\n",
		"boolean":                "",
		"date":                   "    time: true\n",
		"select":                 "    options:\n      - a\n      - value: b\n        text: Bee\n",
		"select-from-query":      "    query_glob: content/**/*.md\n    query_string: .title\n",
		"chips":                  "    options:\n      - a\n",
		"color":                  "",
		"font-picker":            "    limit: 20\n",
		"fonticon-picker":        "    multiple: true\n",
		"image-select":           "    path: static/images\n",
		"eisenhouwer":            "    dataSetsPath: data\n",
		"section":                "    fields:\n      - key: inner\n        type: string\n",
		"nest":                   "    max_depth: 3\n    fields:\n      - key: inner\n        type: string\n",
		"accordion":              "    disableSort: true\n    fields:\n      - key: inner\n        type: string\n",
		"pull":                   "    group: meta\n    fields:\n      - key: inner\n        type: string\n",
		"leaf-array":             "    field:\n      key: item\n      type: string\n",
		"bundle-manager":         "    path: bundles\n    fields:\n      - key: inner\n        type: string\n",
		"hidden":                 "    default: fixed\n",
		"empty-line":             "    amount: 2\n",
		"info":                   "    content: Read this first\n",
		"bundle-image-thumbnail": "    src: thumb.png\n",
	}

	for _, fieldType := range KnownFieldTypes {
		extra, declared := declarations[fieldType]
		if !declared {
			t.Errorf("%s is a known field type with no case in this table", fieldType)
			continue
		}

		t.Run(fieldType, func(t *testing.T) {
			m := load(t, withFields("      - key: f\n        type: "+fieldType+"\n"+indent(extra)))

			fields := m.Singles[0].Fields
			if len(fields) != 1 {
				t.Fatalf("got %d fields, want 1", len(fields))
			}
			if fields[0].Type != fieldType {
				t.Errorf("type = %q, want %q", fields[0].Type, fieldType)
			}
			if fields[0].Config != nil && fields[0].Config.fieldType() != fieldType {
				t.Errorf("config is for %q, want %q", fields[0].Config.fieldType(), fieldType)
			}
			if len(fields[0].Extra) != 0 {
				t.Errorf("extra = %v, want every declared property consumed", fields[0].Extra)
			}
		})
	}
}

func indent(block string) string {
	if block == "" {
		return ""
	}
	var out strings.Builder
	for _, line := range strings.Split(strings.TrimRight(block, "\n"), "\n") {
		out.WriteString("    " + line + "\n")
	}
	return out.String()
}

func TestSelectOptionsAcceptBothShapes(t *testing.T) {
	m := load(t, withFields(`      - key: layout
        type: select
        options:
          - plain
          - 3
          - value: wide
            text: Wide
`))

	config := m.Singles[0].Fields[0].Config.(SelectConfig)
	if len(config.Options) != 3 {
		t.Fatalf("got %d options, want 3", len(config.Options))
	}
	if config.Options[0].Value != "plain" || config.Options[1].Value != int64(3) {
		t.Errorf("scalar options = %+v", config.Options[:2])
	}
	if config.Options[2].Value != "wide" || config.Options[2].Text != "Wide" {
		t.Errorf("object option = %+v", config.Options[2])
	}
}

func TestContainersNestToAnyDepth(t *testing.T) {
	m := load(t, withFields(`      - key: outer
        type: accordion
        fields:
          - key: middle
            type: section
            fields:
              - key: inner
                type: nest
                fields:
                  - key: leaf
                    type: string
`))

	field := m.Singles[0].Fields[0]
	for _, want := range []string{"middle", "inner", "leaf"} {
		if len(field.Fields) != 1 {
			t.Fatalf("expected one child on the way to %q, got %d", want, len(field.Fields))
		}
		field = field.Fields[0]
		if field.Key != want {
			t.Fatalf("child key = %q, want %q", field.Key, want)
		}
	}
}

func TestLeafArrayCarriesOneChildField(t *testing.T) {
	m := load(t, withFields(`      - key: tags
        type: leaf-array
        field:
          key: tag
          type: string
`))

	config := m.Singles[0].Fields[0].Config.(LeafArrayConfig)
	if config.Field == nil || config.Field.Key != "tag" {
		t.Errorf("leaf-array child = %+v, want the single field named tag", config.Field)
	}
	if len(m.Singles[0].Fields[0].Fields) != 0 {
		t.Error("a leaf-array gained a fields list it should not have")
	}
}

func TestUnknownFieldTypeSurvives(t *testing.T) {
	m := load(t, withFields(`      - key: rating
        type: star-rating
        maximum: 5
`))

	field := m.Singles[0].Fields[0]
	if field.Type != TypeUnknown {
		t.Errorf("type = %q, want %q", field.Type, TypeUnknown)
	}

	config, ok := field.Config.(UnknownConfig)
	if !ok {
		t.Fatalf("config = %#v, want an UnknownConfig", field.Config)
	}
	if config.DeclaredType != "star-rating" {
		t.Errorf("declaredType = %q, want star-rating", config.DeclaredType)
	}
	if config.Properties["maximum"] != int64(5) {
		t.Errorf("properties = %v, want the raw properties kept", config.Properties)
	}
}

func TestUnknownKeysArePreserved(t *testing.T) {
	m := load(t, map[string]string{
		"quiqr/model/base.yaml": `futureRootKey: kept
singles:
  - key: home
    futureSingleKey: kept
    fields:
      - key: title
        type: string
        futureFieldKey: kept
collections:
  - key: posts
    folder: content/posts/
    futureCollectionKey: kept
`,
	})

	if m.Extra["futureRootKey"] != "kept" {
		t.Errorf("root extra = %v", m.Extra)
	}
	if m.Singles[0].Extra["futureSingleKey"] != "kept" {
		t.Errorf("single extra = %v", m.Singles[0].Extra)
	}
	if m.Singles[0].Fields[0].Extra["futureFieldKey"] != "kept" {
		t.Errorf("field extra = %v", m.Singles[0].Fields[0].Extra)
	}
	if m.Collections[0].Extra["futureCollectionKey"] != "kept" {
		t.Errorf("collection extra = %v", m.Collections[0].Extra)
	}
}

func TestErrorNamesFileAndPath(t *testing.T) {
	cases := map[string]struct {
		files    map[string]string
		wantPath string
	}{
		"a collection without a folder": {
			files:    map[string]string{"quiqr/model/base.yaml": "collections:\n  - key: posts\n"},
			wantPath: "collections[0].folder",
		},
		"a nested field without a key": {
			files: map[string]string{"quiqr/model/base.yaml": `singles:
  - key: home
    fields:
      - key: outer
        type: section
        fields:
          - type: string
`},
			wantPath: "singles[0].fields[0].fields[0].key",
		},
		"a field without a type": {
			files:    map[string]string{"quiqr/model/base.yaml": "singles:\n  - key: home\n    fields:\n      - key: title\n"},
			wantPath: "singles[0].fields[0].type",
		},
		"two singles sharing a key": {
			files:    map[string]string{"quiqr/model/base.yaml": "singles:\n  - key: home\n  - key: home\n"},
			wantPath: "singles",
		},
		"two collections sharing a key": {
			files: map[string]string{"quiqr/model/base.yaml": `collections:
  - key: posts
    folder: a/
  - key: posts
    folder: b/
`},
			wantPath: "collections",
		},
		"a slider without a maximum": {
			files: map[string]string{"quiqr/model/base.yaml": `singles:
  - key: home
    fields:
      - key: weight
        type: slider
        min: 0
`},
			wantPath: "singles[0].fields[0].max",
		},
		"a select without options": {
			files: map[string]string{"quiqr/model/base.yaml": `singles:
  - key: home
    fields:
      - key: layout
        type: select
`},
			wantPath: "singles[0].fields[0].options",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := Load(site(tc.files))
			if err == nil {
				t.Fatal("Load() error = nil, want a model error")
			}

			var modelErr *Error
			if !asModelError(err, &modelErr) {
				t.Fatalf("error = %v, want a *model.Error", err)
			}
			if modelErr.File != "quiqr/model/base.yaml" {
				t.Errorf("file = %q, want the base file", modelErr.File)
			}
			if modelErr.Path != tc.wantPath {
				t.Errorf("path = %q, want %q", modelErr.Path, tc.wantPath)
			}
		})
	}
}

func asModelError(err error, target **Error) bool {
	typed, ok := err.(*Error)
	if ok {
		*target = typed
	}
	return ok
}

func TestWrongTypeForAKnownKeyIsRejected(t *testing.T) {
	cases := map[string]string{
		"a title that is a number":  "singles:\n  - key: home\n    title: 42\n",
		"a hidden that is a string": "singles:\n  - key: home\n    hidden: yes-please\n",
		"a fields list of scalars":  "singles:\n  - key: home\n    fields:\n      - just a string\n",
	}

	for name, document := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Load(site(map[string]string{"quiqr/model/base.yaml": document})); err == nil {
				t.Error("Load() error = nil, want a rejection")
			}
		})
	}
}

func TestAListKeyThatIsNotAListIsAWarning(t *testing.T) {
	m := load(t, map[string]string{
		"quiqr/model/base.yaml": "singles:\n  home:\n    file: a.md\n",
	})

	if len(m.Singles) != 0 {
		t.Errorf("singles = %+v, want none", m.Singles)
	}
	if len(m.Warnings) != 1 || !strings.Contains(m.Warnings[0].String(), "singles") {
		t.Errorf("warnings = %v, want one naming singles", m.Warnings)
	}
}

func TestLoadReportsNoModel(t *testing.T) {
	_, err := Load(fstest.MapFS{})
	if err != ErrNoModel {
		t.Errorf("Load() error = %v, want ErrNoModel", err)
	}
}

func TestSingleAndCollectionLookup(t *testing.T) {
	m := load(t, map[string]string{
		"quiqr/model/base.yaml": "singles:\n  - key: home\ncollections:\n  - key: posts\n    folder: content/posts/\n",
	})

	if _, ok := m.Single("home"); !ok {
		t.Error("Single(home) not found")
	}
	if _, ok := m.Single("absent"); ok {
		t.Error("Single(absent) found something")
	}
	if _, ok := m.Collection("posts"); !ok {
		t.Error("Collection(posts) not found")
	}
	if _, ok := m.Collection("absent"); ok {
		t.Error("Collection(absent) found something")
	}
}

func TestEveryConfigNamesItsType(t *testing.T) {
	configs := []FieldConfig{
		StringConfig{},
		MultiLineConfig{Type: "easymde"},
		NumberConfig{},
		SliderConfig{},
		DateConfig{},
		SelectConfig{},
		SelectFromQueryConfig{},
		ChipsConfig{},
		FontPickerConfig{},
		FontIconPickerConfig{},
		ImageSelectConfig{},
		EisenhouwerConfig{},
		SectionConfig{Type: "section"},
		NestConfig{},
		AccordionConfig{},
		PullConfig{},
		LeafArrayConfig{},
		BundleManagerConfig{},
		BundleImageThumbnailConfig{},
		EmptyLineConfig{},
		InfoConfig{},
		UnknownConfig{},
	}

	for _, config := range configs {
		name := config.fieldType()
		if name == "" {
			t.Errorf("%T names no field type", config)
		}
		if name != TypeUnknown && !IsKnownFieldType(name) {
			t.Errorf("%T names %q, which is not a known field type", config, name)
		}
	}
}
