package forms

import (
	"errors"
	"fmt"

	"github.com/quiqr/chatcms/internal/frontmatter"
	"github.com/quiqr/chatcms/internal/model"
)

// ErrNoSuchField says a path does not name anything the model declares.
var ErrNoSuchField = errors.New("this model declares no such field")

// Form is one content file, seen through the fields a model declares.
type Form struct {
	document *frontmatter.Document
	bindings []Binding
	byPath   map[string]Binding
}

// Load reads a content file against a field tree.
func Load(fields []model.Field, name string, data []byte) (*Form, error) {
	document, err := frontmatter.Parse(name, data)
	if err != nil {
		return nil, err
	}
	return New(fields, document), nil
}

// New wraps a document that has already been read.
func New(fields []model.Field, document *frontmatter.Document) *Form {
	bindings := Bind(fields)

	byPath := make(map[string]Binding, len(bindings))
	for _, binding := range bindings {
		byPath[binding.FieldPath] = binding
	}

	return &Form{document: document, bindings: bindings, byPath: byPath}
}

// Bindings lists everything this form can edit.
func (f *Form) Bindings() []Binding {
	out := make([]Binding, len(f.bindings))
	copy(out, f.bindings)
	return out
}

// Document is the file behind the form.
func (f *Form) Document() *frontmatter.Document {
	return f.document
}

// Body is the Markdown after the front matter.
func (f *Form) Body() string {
	return f.document.Body
}

// Get reads the value of one field.
func (f *Form) Get(fieldPath string) (any, bool) {
	binding, ok := f.byPath[fieldPath]
	if !ok {
		return nil, false
	}
	if binding.IsBody {
		return f.document.Body, true
	}
	return read(f.document.Fields, binding.DataPath)
}

// Set changes the value of one field, creating whatever it has to live in.
func (f *Form) Set(fieldPath string, value any) error {
	binding, ok := f.byPath[fieldPath]
	if !ok {
		return fmt.Errorf("%w: %s", ErrNoSuchField, fieldPath)
	}
	if binding.IsBody {
		text, isString := value.(string)
		if !isString {
			return fmt.Errorf("the body of %s has to be text", fieldPath)
		}
		f.document.Body = text
		return nil
	}
	write(f.document.Fields, binding.DataPath, binding.Siblings, value)
	return nil
}

// Clear writes the value the reference implementation writes for a field that
// has been emptied, rather than removing the key.
func (f *Form) Clear(fieldPath string) error {
	binding, ok := f.byPath[fieldPath]
	if !ok {
		return fmt.Errorf("%w: %s", ErrNoSuchField, fieldPath)
	}
	if binding.IsBody {
		f.document.Body = ""
		return nil
	}
	write(f.document.Fields, binding.DataPath, binding.Siblings, nil)
	return nil
}

// Bytes writes the file back.
func (f *Form) Bytes() ([]byte, error) {
	return f.document.Bytes()
}

func read(root *frontmatter.Map, path []string) (any, bool) {
	current := root
	for i, key := range path {
		value, ok := current.Get(key)
		if !ok {
			return nil, false
		}
		if i == len(path)-1 {
			return value, true
		}
		next, isMap := value.(*frontmatter.Map)
		if !isMap {
			return nil, false
		}
		current = next
	}
	return nil, false
}

func write(root *frontmatter.Map, path []string, siblings [][]string, value any) {
	current := root

	for i, key := range path {
		after := []string(nil)
		if i < len(siblings) {
			after = siblings[i]
		}

		if i == len(path)-1 {
			current.SetAfter(key, value, after)
			return
		}

		existing, ok := current.Get(key)
		next, isMap := existing.(*frontmatter.Map)
		if !ok || !isMap {
			next = frontmatter.NewMap()
			current.SetAfter(key, next, after)
		}
		current = next
	}
}

// FieldPaths lists every editable path, which is what a test or an interface
// iterates over.
func (f *Form) FieldPaths() []string {
	paths := make([]string, 0, len(f.bindings))
	for _, binding := range f.bindings {
		paths = append(paths, binding.FieldPath)
	}
	return paths
}
