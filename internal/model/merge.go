package model

// Merge reproduces the defaults of the JavaScript deepmerge package, which the
// Quiqr desktop app uses to assemble a model. Maps merge recursively, arrays
// concatenate with the target's entries first, and anything else is replaced
// by the source. See docs/QUIQR-MODEL.md for the two call sites where the
// direction reads backwards.
func Merge(target, source any) any {
	targetMap, targetIsMap := target.(map[string]any)
	sourceMap, sourceIsMap := source.(map[string]any)
	if targetIsMap && sourceIsMap {
		merged := make(map[string]any, len(targetMap)+len(sourceMap))
		for key, value := range targetMap {
			merged[key] = clone(value)
		}
		for key, value := range sourceMap {
			if existing, ok := merged[key]; ok {
				merged[key] = Merge(existing, value)
				continue
			}
			merged[key] = clone(value)
		}
		return merged
	}

	targetList, targetIsList := target.([]any)
	sourceList, sourceIsList := source.([]any)
	if targetIsList && sourceIsList {
		merged := make([]any, 0, len(targetList)+len(sourceList))
		for _, value := range targetList {
			merged = append(merged, clone(value))
		}
		for _, value := range sourceList {
			merged = append(merged, clone(value))
		}
		return merged
	}

	return clone(source)
}

func clone(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			out[key] = clone(item)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = clone(item)
		}
		return out
	default:
		return value
	}
}
