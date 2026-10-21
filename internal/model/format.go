package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/goccy/go-yaml"
)

var Extensions = []string{"json", "toml", "yaml", "yml"}

func extensionOf(name string) string {
	return strings.TrimPrefix(strings.ToLower(path.Ext(name)), ".")
}

func supported(ext string) bool {
	for _, candidate := range Extensions {
		if candidate == ext {
			return true
		}
	}
	return false
}

// Parse turns the bytes of one model file into a generic tree. The format is
// chosen by the file's extension, and an unknown extension is read as YAML,
// which is what the reference application falls back to.
func Parse(name string, data []byte) (map[string]any, error) {
	var tree any
	var err error

	switch extensionOf(name) {
	case "json":
		tree, err = parseJSON(data)
	case "toml":
		tree, err = parseTOML(data)
	default:
		tree, err = parseYAML(data)
	}
	if err != nil {
		return nil, &Error{File: name, Want: "a readable model file", Err: err}
	}

	normalised, ok := Normalise(tree).(map[string]any)
	if !ok {
		return nil, &Error{File: name, Want: "a mapping at the top level"}
	}
	return normalised, nil
}

func parseJSON(data []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	var tree any
	if err := decoder.Decode(&tree); err != nil {
		return nil, err
	}
	if decoder.More() {
		return nil, fmt.Errorf("trailing content after the first JSON document")
	}
	return tree, nil
}

func parseTOML(data []byte) (any, error) {
	tree := map[string]any{}
	if _, err := toml.Decode(string(data), &tree); err != nil {
		return nil, err
	}
	return tree, nil
}

func parseYAML(data []byte) (any, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return map[string]any{}, nil
	}

	var tree any
	if err := yaml.Unmarshal(data, &tree); err != nil {
		return nil, err
	}
	return tree, nil
}

// Normalise canonicalises a parsed tree so that the same document in any
// supported format yields an equal tree: map keys are strings, integers are
// int64 and other numbers are float64.
func Normalise(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			out[key] = Normalise(item)
		}
		return out
	case map[any]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			out[fmt.Sprintf("%v", key)] = Normalise(item)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = Normalise(item)
		}
		return out
	case json.Number:
		if integer, err := typed.Int64(); err == nil {
			return integer
		}
		if number, err := typed.Float64(); err == nil {
			return number
		}
		return typed.String()
	case int:
		return int64(typed)
	case int8:
		return int64(typed)
	case int16:
		return int64(typed)
	case int32:
		return int64(typed)
	case int64:
		return typed
	case uint:
		return int64(typed)
	case uint8:
		return int64(typed)
	case uint16:
		return int64(typed)
	case uint32:
		return int64(typed)
	case uint64:
		return int64(typed)
	case float32:
		return float64(typed)
	default:
		return value
	}
}
