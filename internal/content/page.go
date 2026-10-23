package content

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/quiqr/chatcms/internal/forms"
	"github.com/quiqr/chatcms/internal/frontmatter"
	"github.com/quiqr/chatcms/internal/model"
)

// Page is one editable thing in a site: a single page the model names, or one
// item of a collection.
type Page struct {
	// Key addresses the page: a single's own key, or a collection's key and
	// the item's name joined by a slash.
	Key     string
	Title   string
	Section string
	// File is where it lives. It is never shown to a content creator.
	File   string
	Fields []model.Field
}

// Locate works out which page a key names, using the model.
func Locate(m *model.Model, key string) (Page, bool) {
	if m == nil || key == "" {
		return Page{}, false
	}

	section, item, nested := strings.Cut(key, "/")
	if !nested {
		for _, single := range m.Singles {
			if single.Key == key {
				return Page{
					Key:     key,
					Title:   titleOr(single.Title, single.Key),
					Section: "Pages",
					File:    single.File,
					Fields:  single.Fields,
				}, true
			}
		}
		return Page{}, false
	}

	for _, collection := range m.Collections {
		if collection.Key != section {
			continue
		}
		name := itemFile(item, collection.Extension)
		if name == "" {
			return Page{}, false
		}
		return Page{
			Key:     key,
			Title:   readable(item),
			Section: titleOr(collection.Title, collection.Key),
			File:    path.Join(strings.TrimSuffix(collection.Folder, "/"), name),
			Fields:  collection.Fields,
		}, true
	}
	return Page{}, false
}

// itemFile turns an item's name into a file name, refusing anything that
// could address a file outside the collection.
func itemFile(item, extension string) string {
	if item == "" || item == "." || item == ".." ||
		strings.ContainsAny(item, `/\:`) || strings.Contains(item, "..") {
		return ""
	}
	if extension == "" {
		extension = "md"
	}
	if strings.HasSuffix(item, "."+extension) {
		return item
	}
	return item + "." + extension
}

func titleOr(title, key string) string {
	if title != "" {
		return title
	}
	return readable(key)
}

func readable(key string) string {
	spaced := strings.NewReplacer("-", " ", "_", " ").Replace(key)
	spaced = strings.TrimSuffix(spaced, path.Ext(spaced))
	if spaced == "" {
		return key
	}
	return strings.ToUpper(spaced[:1]) + spaced[1:]
}

// Control is one thing a person can change, with its value. Groups carry the
// controls inside them; a list of items carries the shape of one item.
type Control struct {
	Path     string         `json:"path"`
	Type     string         `json:"type"`
	Title    string         `json:"title"`
	Tip      string         `json:"tip,omitempty"`
	Value    any            `json:"value,omitempty"`
	Options  []Option       `json:"options,omitempty"`
	Controls []Control      `json:"controls,omitempty"`
	Config   map[string]any `json:"config,omitempty"`
	// Unknown says the model asked for a kind of field this version does not
	// render. It is shown as plain text rather than left out, so that nothing
	// a site holds can be lost by being invisible.
	Unknown bool `json:"unknown,omitempty"`
}

type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// skip says a field is not drawn as a control of its own. A hidden field is
// not for a person at all, and the page's own text is drawn by the editor
// for it rather than twice.
func skip(field model.Field) bool {
	return field.Hidden || field.Type == "hidden" || field.Key == forms.BodyKey
}

// presentational fields carry no value and are shown as they are.
var presentational = map[string]bool{"info": true, "empty-line": true}

// grouping fields hold other fields rather than a value of their own.
var grouping = map[string]bool{"section": true, "nest": true, "pull": true}

// rendered is every field type this version draws a control for. Anything
// else is still shown, as text, with a note.
var rendered = map[string]bool{
	"string": true, "markdown": true, "easymde": true, "readonly": true,
	"number": true, "slider": true, "date": true, "select": true,
	"boolean": true, "color": true, "chips": true, "leaf-array": true,
	"accordion": true, "hidden": true, "uniq": true,
	"image-select": true, "bundle-image-thumbnail": true,
}

// Describe turns a page's fields and its current values into the controls an
// interface draws.
func Describe(fields []model.Field, form *forms.Form) []Control {
	return describe(fields, "", form)
}

func describe(fields []model.Field, prefix string, form *forms.Form) []Control {
	controls := []Control{}

	for _, field := range fields {
		if skip(field) {
			continue
		}

		where := field.Key
		if prefix != "" && where != "" {
			where = prefix + "." + where
		}

		control := Control{
			Path:   where,
			Type:   field.Type,
			Title:  titleOr(field.Title, field.Key),
			Tip:    field.Tip,
			Config: configOf(field),
		}

		switch {
		case presentational[field.Type]:
			control.Title = field.Title
			if control.Title == "" {
				control.Title = field.Content
			}
		case grouping[field.Type]:
			control.Controls = describe(field.Fields, where, form)
		case field.Type == "accordion":
			control.Controls = shapeOf(field.Fields)
			control.Value = valueOf(form, where)
		default:
			control.Unknown = !rendered[field.Type]
			control.Value = valueOf(form, where)
			control.Options = optionsOf(field)
		}

		controls = append(controls, control)
	}
	return controls
}

// shapeOf is what one item of a list of items looks like. Its paths are
// relative to the item, because every item has the same shape.
func shapeOf(fields []model.Field) []Control {
	controls := []Control{}
	for _, field := range fields {
		if skip(field) {
			continue
		}
		controls = append(controls, Control{
			Path:    field.Key,
			Type:    field.Type,
			Title:   titleOr(field.Title, field.Key),
			Tip:     field.Tip,
			Options: optionsOf(field),
			Config:  configOf(field),
			Unknown: !rendered[field.Type],
		})
	}
	return controls
}

func valueOf(form *forms.Form, where string) any {
	if form == nil {
		return nil
	}
	held, found := form.Get(where)
	if !found {
		return nil
	}
	return Plain(held)
}

func optionsOf(field model.Field) []Option {
	config, ok := field.Config.(model.SelectConfig)
	if !ok {
		return nil
	}
	options := make([]Option, 0, len(config.Options))
	for _, option := range config.Options {
		value := fmt.Sprintf("%v", option.Value)
		label := option.Text
		if label == "" {
			label = value
		}
		options = append(options, Option{Value: value, Label: label})
	}
	return options
}

// configOf is the handful of settings a control needs to draw itself. It is
// deliberately small: anything the interface does not use is left out rather
// than passed through and forgotten about.
func configOf(field model.Field) map[string]any {
	config := map[string]any{}

	switch held := field.Config.(type) {
	case model.StringConfig:
		if held.MultiLine {
			config["multiLine"] = true
		}
	case model.MultiLineConfig:
		if held.MultiLine {
			config["multiLine"] = true
		}
	case model.NumberConfig:
		put(config, "min", held.Min)
		put(config, "max", held.Max)
		put(config, "step", held.Step)
	case model.SliderConfig:
		config["min"] = held.Min
		config["max"] = held.Max
		put(config, "step", held.Step)
	case model.DateConfig:
		if held.DateFormat != "" {
			config["dateFormat"] = held.DateFormat
		}
	}

	if len(config) == 0 {
		return nil
	}
	return config
}

func put(config map[string]any, name string, value *float64) {
	if value != nil {
		config[name] = *value
	}
}

// Plain turns what the document holds into something that can be sent as
// JSON, losing the ordering the document keeps. Order is never read back from
// here: a value is written by its own path, into the document, which keeps it.
func Plain(value any) any {
	switch held := value.(type) {
	case nil:
		return nil
	case []any:
		out := make([]any, 0, len(held))
		for _, item := range held {
			out = append(out, Plain(item))
		}
		return out
	default:
		if asMap, ok := value.(*frontmatter.Map); ok && asMap != nil {
			out := map[string]any{}
			for _, key := range asMap.Keys() {
				inner, _ := asMap.Get(key)
				out[key] = Plain(inner)
			}
			return out
		}
		return value
	}
}

// Paths is every control's path, deepest last, which is what a save walks.
func Paths(controls []Control) []string {
	found := []string{}
	for _, control := range controls {
		if len(control.Controls) > 0 && control.Type != "accordion" {
			found = append(found, Paths(control.Controls)...)
			continue
		}
		if presentational[control.Type] {
			continue
		}
		found = append(found, control.Path)
	}
	sort.Strings(found)
	return found
}
