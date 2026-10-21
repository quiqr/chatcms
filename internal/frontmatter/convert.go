package frontmatter

import (
	"encoding/json"
	"errors"
	"time"
)

// fromGeneric canonicalises a parsed value into the shapes the emitters know.
func fromGeneric(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := NewMap()
		for key, item := range typed {
			out.Set(key, fromGeneric(item))
		}
		return out
	case []any:
		items := make([]any, len(typed))
		for i, item := range typed {
			items[i] = fromGeneric(item)
		}
		return items
	case json.Number:
		if integer, err := typed.Int64(); err == nil {
			return integer
		}
		number, err := typed.Float64()
		if err != nil {
			return typed.String()
		}
		return number
	case int:
		return int64(typed)
	case int32:
		return int64(typed)
	case uint64:
		return int64(typed)
	case float32:
		return float64(typed)
	case time.Time:
		return typed.UTC()
	case string:
		// js-yaml resolves a scalar that reads as a timestamp to a moment in
		// time, so a reader that kept it as a string would write it back
		// quoted and differ from the reference on the very first save.
		if when, ok := asTimestamp(typed); ok {
			return when
		}
		return typed
	default:
		return value
	}
}

func decodeJSONValue(decoder *json.Decoder) (any, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}

	switch typed := token.(type) {
	case json.Delim:
		switch typed {
		case '{':
			out := NewMap()
			for decoder.More() {
				keyToken, err := decoder.Token()
				if err != nil {
					return nil, err
				}
				key, ok := keyToken.(string)
				if !ok {
					return nil, errors.New("an object key is not a string")
				}
				value, err := decodeJSONValue(decoder)
				if err != nil {
					return nil, err
				}
				out.Set(key, value)
			}
			if _, err := decoder.Token(); err != nil {
				return nil, err
			}
			return out, nil

		case '[':
			items := []any{}
			for decoder.More() {
				value, err := decodeJSONValue(decoder)
				if err != nil {
					return nil, err
				}
				items = append(items, value)
			}
			if _, err := decoder.Token(); err != nil {
				return nil, err
			}
			return items, nil
		}
		return nil, errors.New("unexpected delimiter")

	case json.Number:
		return fromGeneric(typed), nil
	case string, bool, nil:
		return typed, nil
	}
	return nil, nil
}
