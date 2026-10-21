// Package forms maps a content file's values onto the fields a Quiqr model
// declares, and back again. Where a value lives is decided by the field tree,
// not by the shape of the data, so a value no field claims is left where it
// is rather than guessed at.
package forms

import (
	"strings"

	"github.com/quiqr/chatcms/internal/model"
)

// BodyKey is the name the reference implementation gives the Markdown body
// when it treats it as one more value. A model that declares a field with
// this key edits the body.
const BodyKey = "mainContent"

// Binding is one editable thing: where the model puts it, and where the file
// keeps it.
type Binding struct {
	// FieldPath addresses the field as the model declares it, with every
	// grouping field in the way, such as "seo.description".
	FieldPath string
	// DataPath addresses the value in the file, which is shorter whenever a
	// grouping field does not group its data.
	DataPath []string
	// Siblings holds, for each step of DataPath, the keys the model declares
	// before it at that level. A value the file does not have yet is placed
	// after the last of them that it does, so a new field lands where the
	// model implies rather than at the end.
	Siblings [][]string
	Field    model.Field
	// IsBody says this binding is the Markdown body rather than a value.
	IsBody bool
}

// Bind walks a field tree and lists everything a form can edit.
func Bind(fields []model.Field) []Binding {
	var bindings []Binding
	bind(fields, nil, nil, nil, nil, &bindings)
	return bindings
}

// bind walks one level. declared is the keys already declared at this data
// level, which a value added later is placed after; it is threaded through
// rather than restarted so that a group which does not group its data keeps
// its parent's ordering.
func bind(
	fields []model.Field,
	fieldPath []string,
	dataPath []string,
	siblings [][]string,
	declared []string,
	out *[]Binding,
) []string {
	for _, field := range fields {
		here := append(append([]string{}, fieldPath...), field.Key)

		switch field.Type {
		case "section", "nest":
			if !field.GroupsData() {
				declared = bind(field.Fields, here, dataPath, siblings, declared, out)
				continue
			}
			childData := append(append([]string{}, dataPath...), field.Key)
			childSiblings := append(cloneLevels(siblings), append([]string{}, declared...))
			bind(field.Fields, here, childData, childSiblings, nil, out)
			declared = append(declared, field.Key)

		case "pull":
			group := field.Key
			if config, ok := field.Config.(model.PullConfig); ok && config.Group != "" {
				group = config.Group
			}
			childData := append(append([]string{}, dataPath...), group)
			childSiblings := append(cloneLevels(siblings), append([]string{}, declared...))
			bind(field.Fields, here, childData, childSiblings, nil, out)
			declared = append(declared, group)

		case "empty-line", "info", "bundle-image-thumbnail":
			// Layout only: nothing is stored.

		default:
			*out = append(*out, Binding{
				FieldPath: strings.Join(here, "."),
				DataPath:  append(append([]string{}, dataPath...), field.Key),
				Siblings:  append(cloneLevels(siblings), append([]string{}, declared...)),
				Field:     field,
				IsBody:    field.Key == BodyKey,
			})
			declared = append(declared, field.Key)
		}
	}
	return declared
}

func cloneLevels(levels [][]string) [][]string {
	out := make([][]string, len(levels))
	for i, level := range levels {
		out[i] = append([]string{}, level...)
	}
	return out
}
