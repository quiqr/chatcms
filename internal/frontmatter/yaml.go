package frontmatter

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// yamlLineWidth is js-yaml's default. A scalar longer than the width minus
// the indent its block would sit at is folded across lines.
const yamlLineWidth = 80

// yamlIndent is js-yaml's default indentation step.
const yamlIndent = 2

// EmitYAML writes a map the way js-yaml writes it at its defaults.
func EmitYAML(m *Map) ([]byte, error) {
	var out strings.Builder
	if err := emitYAMLMap(&out, m, 0, ""); err != nil {
		return nil, err
	}
	return []byte(out.String()), nil
}

func emitYAMLMap(out *strings.Builder, m *Map, indent int, path string) error {
	for _, key := range m.emitOrder() {
		value, _ := m.Get(key)
		where := join(path, key)

		out.WriteString(strings.Repeat(" ", indent))
		out.WriteString(yamlKey(key))
		out.WriteString(":")

		if err := emitYAMLValue(out, value, indent, where); err != nil {
			return err
		}
	}
	return nil
}

// emitYAMLValue writes what follows a key's colon, or a sequence entry's dash.
func emitYAMLValue(out *strings.Builder, value any, indent int, path string) error {
	switch typed := value.(type) {
	case *Map:
		if typed.Len() == 0 {
			out.WriteString(" {}\n")
			return nil
		}
		out.WriteString("\n")
		return emitYAMLMap(out, typed, indent+yamlIndent, path)

	case []any:
		if len(typed) == 0 {
			out.WriteString(" []\n")
			return nil
		}
		out.WriteString("\n")
		return emitYAMLSequence(out, typed, indent+yamlIndent, path)

	default:
		if !supported(value) {
			return unsupported(path, value)
		}
		scalar, err := yamlScalar(value, indent+yamlIndent)
		if err != nil {
			return &UnsupportedError{Path: path, Kind: err.Error()}
		}
		out.WriteString(" ")
		out.WriteString(scalar)
		out.WriteString("\n")
		return nil
	}
}

func emitYAMLSequence(out *strings.Builder, items []any, indent int, path string) error {
	for index, item := range items {
		where := fmt.Sprintf("%s[%d]", path, index)

		out.WriteString(strings.Repeat(" ", indent))
		out.WriteString("-")

		switch typed := item.(type) {
		case *Map:
			if typed.Len() == 0 {
				out.WriteString(" {}\n")
				continue
			}
			// A mapping inside a sequence starts on the dash's own line.
			out.WriteString(" ")
			if err := emitYAMLMapInline(out, typed, indent+yamlIndent, where); err != nil {
				return err
			}

		case []any:
			if len(typed) == 0 {
				out.WriteString(" []\n")
				continue
			}
			out.WriteString(" ")
			if err := emitYAMLSequenceInline(out, typed, indent+yamlIndent, where); err != nil {
				return err
			}

		default:
			if !supported(item) {
				return unsupported(where, item)
			}
			scalar, err := yamlScalar(item, indent+yamlIndent)
			if err != nil {
				return &UnsupportedError{Path: where, Kind: err.Error()}
			}
			out.WriteString(" ")
			out.WriteString(scalar)
			out.WriteString("\n")
		}
	}
	return nil
}

// emitYAMLMapInline writes a mapping whose first key shares the line with the
// sequence dash that introduces it.
func emitYAMLMapInline(out *strings.Builder, m *Map, indent int, path string) error {
	keys := m.emitOrder()

	for i, key := range keys {
		value, _ := m.Get(key)
		where := join(path, key)

		if i > 0 {
			out.WriteString(strings.Repeat(" ", indent))
		}
		out.WriteString(yamlKey(key))
		out.WriteString(":")

		if err := emitYAMLValue(out, value, indent, where); err != nil {
			return err
		}
	}
	return nil
}

func emitYAMLSequenceInline(out *strings.Builder, items []any, indent int, path string) error {
	for index, item := range items {
		where := fmt.Sprintf("%s[%d]", path, index)

		if index > 0 {
			out.WriteString(strings.Repeat(" ", indent))
		}
		out.WriteString("-")

		switch typed := item.(type) {
		case *Map:
			if typed.Len() == 0 {
				out.WriteString(" {}\n")
				continue
			}
			out.WriteString(" ")
			if err := emitYAMLMapInline(out, typed, indent+yamlIndent, where); err != nil {
				return err
			}
		case []any:
			if len(typed) == 0 {
				out.WriteString(" []\n")
				continue
			}
			out.WriteString(" ")
			if err := emitYAMLSequenceInline(out, typed, indent+yamlIndent, where); err != nil {
				return err
			}
		default:
			if !supported(item) {
				return unsupported(where, item)
			}
			scalar, err := yamlScalar(item, indent+yamlIndent)
			if err != nil {
				return &UnsupportedError{Path: where, Kind: err.Error()}
			}
			out.WriteString(" ")
			out.WriteString(scalar)
			out.WriteString("\n")
		}
	}
	return nil
}

func yamlKey(key string) string {
	scalar, err := yamlString(key, 0, false)
	if err != nil {
		return "'" + strings.ReplaceAll(key, "'", "''") + "'"
	}
	return scalar
}

// yamlScalar writes one value. blockIndent is where a block scalar's body
// would sit, which decides when a long string folds.
func yamlScalar(value any, blockIndent int) (string, error) {
	switch typed := value.(type) {
	case nil:
		return "null", nil
	case bool:
		if typed {
			return "true", nil
		}
		return "false", nil
	case int64:
		return strconv.FormatInt(typed, 10), nil
	case float64:
		return yamlNumber(typed), nil
	case time.Time:
		return typed.UTC().Format("2006-01-02T15:04:05.000Z"), nil
	case string:
		return yamlString(typed, blockIndent, true)
	default:
		return "", fmt.Errorf("%T", value)
	}
}

// yamlNumber writes a float the way JavaScript writes a number: a whole value
// loses its fractional part entirely.
func yamlNumber(value float64) string {
	if value == float64(int64(value)) && value < 1e21 && value > -1e21 {
		return strconv.FormatInt(int64(value), 10)
	}
	return strconv.FormatFloat(value, 'g', -1, 64)
}
