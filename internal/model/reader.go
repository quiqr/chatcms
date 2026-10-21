package model

import "fmt"

// reader pulls typed values out of one generic map, remembers which keys it
// used so the rest can be preserved, and keeps the first error it hits.
type reader struct {
	file string
	path string
	tree map[string]any
	used map[string]bool
	err  *Error
}

func newReader(file, path string, tree map[string]any) *reader {
	return &reader{file: file, path: path, tree: tree, used: map[string]bool{}}
}

func (r *reader) fail(key, want string) {
	if r.err != nil {
		return
	}
	r.err = at(r.file, join(r.path, key), want)
}

func (r *reader) take(key string) (any, bool) {
	value, ok := r.tree[key]
	if ok {
		r.used[key] = true
	}
	return value, ok
}

func (r *reader) str(key string) string {
	value, ok := r.take(key)
	if !ok || value == nil {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		r.fail(key, "a string")
		return ""
	}
	return text
}

func (r *reader) requiredStr(key string) string {
	text := r.str(key)
	if text == "" && r.err == nil {
		r.fail(key, "a non-empty string")
	}
	return text
}

func (r *reader) boolean(key string) bool {
	value, ok := r.take(key)
	if !ok || value == nil {
		return false
	}
	flag, ok := value.(bool)
	if !ok {
		r.fail(key, "true or false")
		return false
	}
	return flag
}

func (r *reader) booleanPtr(key string) *bool {
	if _, ok := r.tree[key]; !ok {
		return nil
	}
	flag := r.boolean(key)
	return &flag
}

func (r *reader) number(key string) *float64 {
	value, ok := r.take(key)
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case int64:
		number := float64(typed)
		return &number
	case float64:
		return &typed
	default:
		r.fail(key, "a number")
		return nil
	}
}

func (r *reader) requiredNumber(key string) float64 {
	number := r.number(key)
	if number == nil {
		if r.err == nil {
			r.fail(key, "a number")
		}
		return 0
	}
	return *number
}

func (r *reader) integer(key string) *int64 {
	value, ok := r.take(key)
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case int64:
		return &typed
	case float64:
		rounded := int64(typed)
		return &rounded
	default:
		r.fail(key, "a whole number")
		return nil
	}
}

func (r *reader) strList(key string) []string {
	value, ok := r.take(key)
	if !ok || value == nil {
		return nil
	}
	list, ok := value.([]any)
	if !ok {
		r.fail(key, "a list of strings")
		return nil
	}

	out := make([]string, 0, len(list))
	for index, item := range list {
		text, ok := item.(string)
		if !ok {
			r.fail(fmt.Sprintf("%s[%d]", key, index), "a string")
			return nil
		}
		out = append(out, text)
	}
	return out
}

func (r *reader) list(key string) []any {
	value, ok := r.take(key)
	if !ok || value == nil {
		return nil
	}
	list, ok := value.([]any)
	if !ok {
		r.fail(key, "a list")
		return nil
	}
	return list
}

func (r *reader) mapList(key string) []map[string]any {
	list := r.list(key)
	if list == nil {
		return nil
	}

	out := make([]map[string]any, 0, len(list))
	for index, item := range list {
		entry, ok := item.(map[string]any)
		if !ok {
			r.fail(fmt.Sprintf("%s[%d]", key, index), "a mapping")
			return nil
		}
		out = append(out, entry)
	}
	return out
}

func (r *reader) mapping(key string) map[string]any {
	value, ok := r.take(key)
	if !ok || value == nil {
		return nil
	}
	entry, ok := value.(map[string]any)
	if !ok {
		r.fail(key, "a mapping")
		return nil
	}
	return entry
}

func (r *reader) raw(key string) any {
	value, _ := r.take(key)
	return value
}

func (r *reader) extra() map[string]any {
	var leftovers map[string]any
	for key, value := range r.tree {
		if r.used[key] {
			continue
		}
		if leftovers == nil {
			leftovers = map[string]any{}
		}
		leftovers[key] = value
	}
	return leftovers
}

func join(parent, key string) string {
	if parent == "" {
		return key
	}
	if key == "" {
		return parent
	}
	return parent + "." + key
}
