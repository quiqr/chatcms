package frontmatter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

const goldenDir = "../../test/fixtures/content/golden"

// datePrefix marks a value in the case table that is a date; see
// scripts/goldens/generate.mjs, which revives it the same way.
const datePrefix = "@@DATE@@"

type goldenIndex struct {
	JSYaml  string `json:"jsYaml"`
	Tomlify string `json:"tomlify"`
	Cases   []struct {
		Name     string `json:"name"`
		YAML     bool   `json:"yaml"`
		TOML     bool   `json:"toml"`
		JSON     bool   `json:"json"`
		Document bool   `json:"document"`
	} `json:"cases"`
}

type testCase struct {
	Name  string          `json:"name"`
	Value json.RawMessage `json:"value"`
}

func loadIndex(t *testing.T) goldenIndex {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join(goldenDir, "index.json"))
	if err != nil {
		t.Fatalf("reading the golden index failed: %v; run scripts/generate-goldens.sh", err)
	}

	var index goldenIndex
	if err := json.Unmarshal(raw, &index); err != nil {
		t.Fatalf("the golden index is unreadable: %v", err)
	}
	return index
}

func loadCases(t *testing.T) map[string]*Map {
	t.Helper()

	raw, err := os.ReadFile("../../scripts/goldens/cases.json")
	if err != nil {
		t.Fatalf("reading the case table failed: %v", err)
	}

	var cases []testCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatalf("the case table is unreadable: %v", err)
	}

	out := make(map[string]*Map, len(cases))
	for _, item := range cases {
		value, err := decodeOrdered(item.Value)
		if err != nil {
			t.Fatalf("case %s is unreadable: %v", item.Name, err)
		}
		asMap, ok := value.(*Map)
		if !ok {
			t.Fatalf("case %s is not a mapping", item.Name)
		}
		out[item.Name] = asMap
	}
	return out
}

// decodeOrdered reads the case table keeping key order, which is the whole
// point: the emitter has to reproduce it.
func decodeOrdered(raw json.RawMessage) (any, error) {
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.UseNumber()
	return decodeValue(decoder)
}

func decodeValue(decoder *json.Decoder) (any, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}

	switch typed := token.(type) {
	case json.Delim:
		switch typed {
		case '{':
			m := NewMap()
			for decoder.More() {
				keyToken, err := decoder.Token()
				if err != nil {
					return nil, err
				}
				value, err := decodeValue(decoder)
				if err != nil {
					return nil, err
				}
				m.Set(keyToken.(string), value)
			}
			if _, err := decoder.Token(); err != nil {
				return nil, err
			}
			return m, nil
		case '[':
			items := []any{}
			for decoder.More() {
				value, err := decodeValue(decoder)
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
		return nil, nil

	case string:
		if after, ok := strings.CutPrefix(typed, datePrefix); ok {
			when, err := time.Parse(time.RFC3339, after)
			if err != nil {
				return nil, err
			}
			return when, nil
		}
		return typed, nil

	case json.Number:
		if integer, err := typed.Int64(); err == nil {
			return integer, nil
		}
		number, err := typed.Float64()
		return number, err

	case bool:
		return typed, nil

	case nil:
		return nil, nil
	}
	return nil, nil
}

func TestYAMLMatchesTheReferenceImplementation(t *testing.T) {
	index := loadIndex(t)
	cases := loadCases(t)

	if index.JSYaml == "" {
		t.Fatal("the goldens do not record which js-yaml produced them")
	}

	var mismatches []string

	for _, entry := range index.Cases {
		if !entry.YAML {
			continue
		}

		t.Run(entry.Name, func(t *testing.T) {
			value, ok := cases[entry.Name]
			if !ok {
				t.Fatalf("the golden index names %s but the case table does not", entry.Name)
			}

			want, err := os.ReadFile(filepath.Join(goldenDir, entry.Name+".yaml.golden"))
			if err != nil {
				t.Fatalf("reading the golden failed: %v", err)
			}

			got, err := EmitYAML(value)
			if err != nil {
				t.Fatalf("EmitYAML() error = %v, want nil", err)
			}

			if string(got) != string(want) {
				mismatches = append(mismatches, entry.Name)
				t.Errorf("js-yaml %s wrote:\n%s\nthis writer wrote:\n%s",
					index.JSYaml, want, got)
			}
		})
	}

	if len(mismatches) > 0 {
		sort.Strings(mismatches)
		t.Logf("%d of %d cases differ: %s", len(mismatches), len(index.Cases), strings.Join(mismatches, " "))
	}
}
