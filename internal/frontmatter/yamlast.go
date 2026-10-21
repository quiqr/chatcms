package frontmatter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/goccy/go-yaml/token"
)

// parseYAMLHeader reads front matter through the syntax tree rather than
// through a decoder, for one reason: whether a scalar was quoted. The
// reference implementation resolves an unquoted scalar that reads as a
// timestamp to a moment in time, and leaves a quoted one a string. A decoder
// that hands back values has already thrown that distinction away, and the
// two are written back differently.
func parseYAMLHeader(header string) (*Map, error) {
	if strings.TrimSpace(header) == "" {
		return NewMap(), nil
	}

	file, err := parser.ParseBytes([]byte(header), 0)
	if err != nil {
		return nil, err
	}
	if len(file.Docs) == 0 || file.Docs[0].Body == nil {
		return NewMap(), nil
	}

	value, err := fromNode(file.Docs[0].Body)
	if err != nil {
		return nil, err
	}

	asMap, ok := value.(*Map)
	if !ok {
		return nil, fmt.Errorf("the front matter is not a mapping")
	}
	return asMap, nil
}

func fromNode(node ast.Node) (any, error) {
	switch typed := node.(type) {
	case *ast.MappingNode:
		out := NewMap()
		for _, pair := range typed.Values {
			key, value, err := fromPair(pair)
			if err != nil {
				return nil, err
			}
			out.Set(key, value)
		}
		return out, nil

	case *ast.MappingValueNode:
		out := NewMap()
		key, value, err := fromPair(typed)
		if err != nil {
			return nil, err
		}
		out.Set(key, value)
		return out, nil

	case *ast.SequenceNode:
		items := make([]any, 0, len(typed.Values))
		for _, item := range typed.Values {
			value, err := fromNode(item)
			if err != nil {
				return nil, err
			}
			items = append(items, value)
		}
		return items, nil

	case *ast.NullNode:
		return nil, nil

	case *ast.BoolNode:
		return typed.Value, nil

	case *ast.IntegerNode:
		return toInt64(typed.Value), nil

	case *ast.FloatNode:
		return typed.Value, nil

	case *ast.LiteralNode:
		// A block scalar is always a string; it cannot be read as anything
		// else however it looks.
		return typed.Value.Value, nil

	case *ast.StringNode:
		return fromStringNode(typed), nil

	case *ast.TagNode:
		return fromNode(typed.Value)

	case *ast.AnchorNode:
		return fromNode(typed.Value)

	default:
		return nil, fmt.Errorf("a value of kind %T is not something a content file can hold", node)
	}
}

func fromPair(pair *ast.MappingValueNode) (string, any, error) {
	key := keyName(pair.Key)

	value, err := fromNode(pair.Value)
	if err != nil {
		return "", nil, err
	}
	return key, value, nil
}

func keyName(node ast.Node) string {
	if stringNode, ok := node.(*ast.StringNode); ok {
		return stringNode.Value
	}
	return strings.TrimSpace(node.String())
}

// fromStringNode applies the reference implementation's timestamp resolver,
// but only to a scalar that was not quoted.
func fromStringNode(node *ast.StringNode) any {
	if node.Token != nil {
		switch node.Token.Type {
		case token.SingleQuoteType, token.DoubleQuoteType:
			return node.Value
		}
	}
	if when, ok := asTimestamp(node.Value); ok {
		return when
	}
	return node.Value
}

func toInt64(value any) any {
	switch typed := value.(type) {
	case int64:
		return typed
	case int:
		return int64(typed)
	case uint64:
		return int64(typed)
	case string:
		if parsed, err := strconv.ParseInt(typed, 0, 64); err == nil {
			return parsed
		}
		return typed
	default:
		return value
	}
}
