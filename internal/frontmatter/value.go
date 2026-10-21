// Package frontmatter writes the front matter of a content file exactly as
// the Quiqr desktop app writes it. The emitters here reproduce js-yaml and
// tomlify-j0.4 at the settings the reference application uses, because no Go
// library produces those bytes and a difference would rewrite every file in a
// customer's repository the first time anyone pressed save.
//
// The emitters are correct for the value shapes a Quiqr model can hold, and
// they refuse anything else rather than guessing. See
// openspec/specs/content/serialisation.
package frontmatter

import (
	"fmt"
	"sort"
	"strconv"
	"time"
)

// Map is an ordered set of named values. Order matters: it is the order the
// file had, and keeping it is what makes an edit change one line.
type Map struct {
	keys   []string
	values map[string]any
}

func NewMap() *Map {
	return &Map{values: map[string]any{}}
}

func (m *Map) Set(key string, value any) {
	if _, exists := m.values[key]; !exists {
		m.keys = append(m.keys, key)
	}
	m.values[key] = value
}

func (m *Map) Get(key string) (any, bool) {
	value, ok := m.values[key]
	return value, ok
}

func (m *Map) Delete(key string) {
	if _, exists := m.values[key]; !exists {
		return
	}
	delete(m.values, key)
	for i, existing := range m.keys {
		if existing == key {
			m.keys = append(m.keys[:i], m.keys[i+1:]...)
			break
		}
	}
}

func (m *Map) Keys() []string {
	out := make([]string, len(m.keys))
	copy(out, m.keys)
	return out
}

func (m *Map) Len() int {
	return len(m.keys)
}

func (m *Map) Clone() *Map {
	out := NewMap()
	for _, key := range m.keys {
		out.Set(key, cloneValue(m.values[key]))
	}
	return out
}

func cloneValue(value any) any {
	switch typed := value.(type) {
	case *Map:
		return typed.Clone()
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = cloneValue(item)
		}
		return out
	default:
		return value
	}
}

// emitOrder is the order the reference implementation writes keys in. It
// stores values in a JavaScript object, which lists keys that look like array
// indices first, in numeric order, and every other key in the order it was
// added. Reproducing it matters only for front matter with numeric keys,
// which is rare and still has to match.
func (m *Map) emitOrder() []string {
	var numeric []string
	var rest []string

	for _, key := range m.keys {
		if isArrayIndexKey(key) {
			numeric = append(numeric, key)
			continue
		}
		rest = append(rest, key)
	}

	sort.Slice(numeric, func(i, j int) bool {
		left, _ := strconv.ParseUint(numeric[i], 10, 64)
		right, _ := strconv.ParseUint(numeric[j], 10, 64)
		return left < right
	})

	return append(numeric, rest...)
}

// isArrayIndexKey matches what JavaScript treats as an array index: a
// canonical non-negative integer below 2^32-1.
func isArrayIndexKey(key string) bool {
	if key == "" || (len(key) > 1 && key[0] == '0') {
		return false
	}
	number, err := strconv.ParseUint(key, 10, 64)
	if err != nil {
		return false
	}
	return number < 4294967295 && strconv.FormatUint(number, 10) == key
}

// UnsupportedError says a value's kind has not been proven to match the
// reference implementation, so writing it would be a guess.
type UnsupportedError struct {
	Path string
	Kind string
}

func (e *UnsupportedError) Error() string {
	return fmt.Sprintf("%s holds a %s, which this writer has not been proven to write the way Quiqr writes it",
		e.Path, e.Kind)
}

func unsupported(path string, value any) error {
	return &UnsupportedError{Path: path, Kind: fmt.Sprintf("%T", value)}
}

// supported reports whether a value is one of the shapes the emitters cover.
func supported(value any) bool {
	switch value.(type) {
	case nil, string, bool, int64, float64, time.Time, []any, *Map:
		return true
	default:
		return false
	}
}

// SetAfter adds a value, placing a key that is not there yet immediately
// after the last of preferAfter that is. It is how a field the model declares
// but the file did not have lands in the place the model implies rather than
// at the end.
func (m *Map) SetAfter(key string, value any, preferAfter []string) {
	if _, exists := m.values[key]; exists {
		m.values[key] = value
		return
	}

	m.values[key] = value

	position := 0
	for _, earlier := range preferAfter {
		for i, existing := range m.keys {
			if existing == earlier && i+1 > position {
				position = i + 1
			}
		}
	}

	m.keys = append(m.keys, "")
	copy(m.keys[position+1:], m.keys[position:])
	m.keys[position] = key
}
