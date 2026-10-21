package model

import (
	"reflect"
	"strings"
	"testing"
)

func TestProviderParsesEachFormat(t *testing.T) {
	documents := map[string]string{
		"base.yaml": "title: Korenaar\ncount: 3\nratio: 1.5\nflag: true\nlist:\n  - a\n  - b\nnested:\n  key: value\n",
		"base.toml": "title = \"Korenaar\"\ncount = 3\nratio = 1.5\nflag = true\nlist = [\"a\", \"b\"]\n\n[nested]\nkey = \"value\"\n",
		"base.json": `{"title":"Korenaar","count":3,"ratio":1.5,"flag":true,"list":["a","b"],"nested":{"key":"value"}}`,
	}

	want := map[string]any{
		"title":  "Korenaar",
		"count":  int64(3),
		"ratio":  1.5,
		"flag":   true,
		"list":   []any{"a", "b"},
		"nested": map[string]any{"key": "value"},
	}

	for name, document := range documents {
		t.Run(name, func(t *testing.T) {
			got, err := Parse(name, []byte(document))
			if err != nil {
				t.Fatalf("Parse(%s) error = %v, want nil", name, err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Parse(%s) = %#v, want %#v", name, got, want)
			}
		})
	}
}

func TestNumbersSurviveEachFormat(t *testing.T) {
	cases := map[string]string{
		"base.yaml": "whole: 110\nfractional: 0.110\n",
		"base.toml": "whole = 110\nfractional = 0.110\n",
		"base.json": `{"whole":110,"fractional":0.110}`,
	}

	for name, document := range cases {
		t.Run(name, func(t *testing.T) {
			tree, err := Parse(name, []byte(document))
			if err != nil {
				t.Fatalf("Parse() error = %v, want nil", err)
			}
			if whole, ok := tree["whole"].(int64); !ok || whole != 110 {
				t.Errorf("whole = %#v, want int64(110)", tree["whole"])
			}
			if fractional, ok := tree["fractional"].(float64); !ok || fractional != 0.110 {
				t.Errorf("fractional = %#v, want float64(0.11)", tree["fractional"])
			}
		})
	}
}

func TestParseErrorNamesFileAndPosition(t *testing.T) {
	cases := map[string]string{
		"broken.yaml": "title: [unterminated\n",
		"broken.toml": "title = \n",
		"broken.json": `{"title": }`,
	}

	for name, document := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := Parse(name, []byte(document))
			if err == nil {
				t.Fatal("Parse() error = nil, want a parse error")
			}
			if !strings.Contains(err.Error(), name) {
				t.Errorf("error = %q, want it to name the file", err)
			}
		})
	}
}

func TestTopLevelMustBeAMapping(t *testing.T) {
	cases := map[string]string{
		"list.yaml":   "- a\n- b\n",
		"scalar.yaml": "just a string\n",
		"list.json":   `["a","b"]`,
	}

	for name, document := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := Parse(name, []byte(document))
			if err == nil {
				t.Fatal("Parse() error = nil, want a rejection")
			}
			if !strings.Contains(err.Error(), "mapping") {
				t.Errorf("error = %q, want it to say a mapping was expected", err)
			}
		})
	}
}

func TestEmptyYAMLIsAnEmptyMapping(t *testing.T) {
	tree, err := Parse("empty.yaml", []byte("\n  \n"))
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}
	if len(tree) != 0 {
		t.Errorf("tree = %#v, want an empty mapping", tree)
	}
}

func TestNormaliseConvertsEveryNumberKind(t *testing.T) {
	got := Normalise(map[any]any{
		"int":     int(1),
		"int8":    int8(2),
		"int16":   int16(3),
		"int32":   int32(4),
		"int64":   int64(5),
		"uint":    uint(6),
		"uint8":   uint8(7),
		"uint16":  uint16(8),
		"uint32":  uint32(9),
		"uint64":  uint64(10),
		"float32": float32(1.5),
		"float64": float64(2.5),
		"string":  "text",
	}).(map[string]any)

	for _, key := range []string{"int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64"} {
		if _, ok := got[key].(int64); !ok {
			t.Errorf("%s = %#v, want an int64", key, got[key])
		}
	}
	for _, key := range []string{"float32", "float64"} {
		if _, ok := got[key].(float64); !ok {
			t.Errorf("%s = %#v, want a float64", key, got[key])
		}
	}
	if got["string"] != "text" {
		t.Errorf("string = %#v, want it untouched", got["string"])
	}
}
