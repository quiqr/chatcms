package model

import (
	"reflect"
	"testing"
)

func TestMergeSemantics(t *testing.T) {
	cases := map[string]struct {
		target any
		source any
		want   any
	}{
		"two maps merge key by key": {
			target: map[string]any{"a": int64(1)},
			source: map[string]any{"b": int64(2)},
			want:   map[string]any{"a": int64(1), "b": int64(2)},
		},
		"the source wins a conflicting key": {
			target: map[string]any{"a": "target"},
			source: map[string]any{"a": "source"},
			want:   map[string]any{"a": "source"},
		},
		"nested maps merge recursively": {
			target: map[string]any{"a": map[string]any{"x": int64(1), "y": int64(2)}},
			source: map[string]any{"a": map[string]any{"y": int64(3)}},
			want:   map[string]any{"a": map[string]any{"x": int64(1), "y": int64(3)}},
		},
		"two arrays concatenate with the target first": {
			target: []any{"a", "b"},
			source: []any{"c"},
			want:   []any{"a", "b", "c"},
		},
		"an array in a map concatenates too": {
			target: map[string]any{"list": []any{"a"}},
			source: map[string]any{"list": []any{"b"}},
			want:   map[string]any{"list": []any{"a", "b"}},
		},
		"a scalar replaces a map": {
			target: map[string]any{"a": map[string]any{"x": int64(1)}},
			source: map[string]any{"a": "flat"},
			want:   map[string]any{"a": "flat"},
		},
		"a map replaces a scalar": {
			target: map[string]any{"a": "flat"},
			source: map[string]any{"a": map[string]any{"x": int64(1)}},
			want:   map[string]any{"a": map[string]any{"x": int64(1)}},
		},
		"an array replaces a scalar": {
			target: map[string]any{"a": "flat"},
			source: map[string]any{"a": []any{"x"}},
			want:   map[string]any{"a": []any{"x"}},
		},
		"nil as a source replaces": {
			target: map[string]any{"a": "flat"},
			source: map[string]any{"a": nil},
			want:   map[string]any{"a": nil},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := Merge(tc.target, tc.source)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Merge() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestMergeDoesNotAliasItsInputs(t *testing.T) {
	target := map[string]any{"list": []any{"a"}, "nested": map[string]any{"x": int64(1)}}
	source := map[string]any{"nested": map[string]any{"y": int64(2)}}

	merged := Merge(target, source).(map[string]any)

	merged["list"].([]any)[0] = "changed"
	merged["nested"].(map[string]any)["x"] = int64(99)

	if target["list"].([]any)[0] != "a" {
		t.Error("mutating the merged array changed the target")
	}
	if target["nested"].(map[string]any)["x"] != int64(1) {
		t.Error("mutating the merged map changed the target")
	}
}
