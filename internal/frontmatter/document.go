package frontmatter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

// Format is which syntax a content file's front matter is written in.
type Format string

const (
	FormatYAML Format = "yaml"
	FormatTOML Format = "toml"
	FormatJSON Format = "json"
	FormatNone Format = "none"
)

const (
	yamlDelimiter = "---"
	tomlDelimiter = "+++"
)

// Document is a content file: named values, and the Markdown body after them.
type Document struct {
	Format Format
	Fields *Map
	Body   string
}

// ErrUnwritableFormat says the front matter is in a format this writer will
// not produce. See openspec/specs/content/serialisation.
var ErrUnwritableFormat = errors.New("front matter in this format cannot be written")

// ParseError says where a content file could not be read.
type ParseError struct {
	File   string
	Format Format
	Err    error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s: the %s front matter could not be read: %v", e.File, e.Format, e.Err)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// Parse reads a content file into its values and its body.
func Parse(name string, data []byte) (*Document, error) {
	text := string(data)

	format, header, body := split(text)
	document := &Document{Format: format, Fields: NewMap(), Body: body}

	if format == FormatNone {
		return document, nil
	}

	fields, err := parseHeader(format, header)
	if err != nil {
		return nil, &ParseError{File: name, Format: format, Err: err}
	}
	document.Fields = fields
	return document, nil
}

// split separates the front matter from the body the way the reference
// implementation does: it looks for the opening delimiter at the very start
// and the closing one at the start of a later line.
func split(text string) (Format, string, string) {
	for _, candidate := range []struct {
		format    Format
		delimiter string
	}{
		{FormatYAML, yamlDelimiter},
		{FormatTOML, tomlDelimiter},
	} {
		header, body, ok := splitDelimited(text, candidate.delimiter)
		if ok {
			return candidate.format, header, body
		}
	}

	if header, body, ok := splitJSON(text); ok {
		return FormatJSON, header, body
	}

	return FormatNone, "", text
}

func splitDelimited(text, delimiter string) (string, string, bool) {
	rest, ok := cutOpening(text, delimiter)
	if !ok {
		return "", "", false
	}

	index := indexOfDelimiterLine(rest, delimiter)
	if index < 0 {
		return "", "", false
	}

	header := rest[:index]
	body := rest[index+len(delimiter):]

	// The reference drops the two line breaks that follow the closing
	// delimiter when there is a body after them.
	if strings.TrimSpace(body) != "" {
		body = strings.TrimPrefix(body, "\r")
		body = strings.TrimPrefix(body, "\n")
		body = strings.TrimPrefix(body, "\r")
		body = strings.TrimPrefix(body, "\n")
	}
	return header, body, true
}

func cutOpening(text, delimiter string) (string, bool) {
	if !strings.HasPrefix(text, delimiter) {
		return "", false
	}
	rest := text[len(delimiter):]
	rest = strings.TrimLeft(rest, " \t")
	if !strings.HasPrefix(rest, "\n") && !strings.HasPrefix(rest, "\r\n") {
		return "", false
	}
	return rest, true
}

// indexOfDelimiterLine finds a delimiter that begins a line, which is what
// the reference's own expression requires.
func indexOfDelimiterLine(text, delimiter string) int {
	offset := 0
	for {
		index := strings.Index(text[offset:], delimiter)
		if index < 0 {
			return -1
		}
		absolute := offset + index

		atLineStart := absolute == 0 || text[absolute-1] == '\n'
		after := text[absolute+len(delimiter):]
		trimmed := strings.TrimLeft(after, " \t")
		endsLine := trimmed == "" || strings.HasPrefix(trimmed, "\n") || strings.HasPrefix(trimmed, "\r")

		if atLineStart && endsLine {
			return absolute
		}
		offset = absolute + len(delimiter)
	}
}

func splitJSON(text string) (string, string, bool) {
	trimmed := strings.TrimLeft(text, " \t\r\n")
	if !strings.HasPrefix(trimmed, "{") {
		return "", "", false
	}

	depth := 0
	inString := false
	escaped := false

	for i, r := range trimmed {
		switch {
		case escaped:
			escaped = false
		case r == '\\' && inString:
			escaped = true
		case r == '"':
			inString = !inString
		case inString:
		case r == '{':
			depth++
		case r == '}':
			depth--
			if depth == 0 {
				header := trimmed[:i+1]
				body := strings.TrimLeft(trimmed[i+1:], "\r\n")
				return header, body, true
			}
		}
	}
	return "", "", false
}

func parseHeader(format Format, header string) (*Map, error) {
	switch format {
	case FormatYAML:
		return parseYAMLHeader(header)
	case FormatTOML:
		return parseTOMLHeader(header)
	case FormatJSON:
		return parseJSONHeader(header)
	}
	return NewMap(), nil
}

func parseTOMLHeader(header string) (*Map, error) {
	var tree map[string]any
	meta, err := toml.Decode(header, &tree)
	if err != nil {
		return nil, err
	}

	out := NewMap()
	for _, key := range meta.Keys() {
		if len(key) != 1 {
			continue
		}
		name := key[0]
		if value, ok := tree[name]; ok {
			out.Set(name, fromGeneric(value))
		}
	}
	// Anything the key list did not reach, such as a table's own name, is
	// added afterwards so nothing is dropped.
	for name, value := range tree {
		if _, seen := out.Get(name); !seen {
			out.Set(name, fromGeneric(value))
		}
	}
	return out, nil
}

func parseJSONHeader(header string) (*Map, error) {
	decoder := json.NewDecoder(bytes.NewReader([]byte(header)))
	decoder.UseNumber()

	value, err := decodeJSONValue(decoder)
	if err != nil {
		return nil, err
	}
	asMap, ok := value.(*Map)
	if !ok {
		return nil, errors.New("the front matter is not an object")
	}
	return asMap, nil
}
