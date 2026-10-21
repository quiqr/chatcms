package model

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// updateGolden is set by CHATCMS_UPDATE_GOLDEN=1, which rewrites the golden
// files instead of comparing against them. It is never set during a normal
// test run; see docs/TESTING.md.
const updateGoldenVar = "CHATCMS_UPDATE_GOLDEN"

const fixturesDir = "../../test/fixtures/models"

type goldenFile struct {
	Model      *Model     `json:"model"`
	Navigation Navigation `json:"navigation"`
}

func fixtureNames(t *testing.T) []string {
	t.Helper()

	entries, err := os.ReadDir(fixturesDir)
	if err != nil {
		t.Fatalf("reading the fixtures failed: %v", err)
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	if len(names) == 0 {
		t.Fatal("no model fixtures found")
	}
	return names
}

func TestEveryFixtureAssembles(t *testing.T) {
	for _, name := range fixtureNames(t) {
		t.Run(name, func(t *testing.T) {
			fsys := os.DirFS(filepath.Join(fixturesDir, name))

			if _, err := Load(fsys); err != nil {
				t.Fatalf("Load() error = %v, want nil", err)
			}
		})
	}
}

func TestModelJSONMatchesGolden(t *testing.T) {
	update := os.Getenv(updateGoldenVar) == "1"

	for _, name := range fixtureNames(t) {
		t.Run(name, func(t *testing.T) {
			fsys := os.DirFS(filepath.Join(fixturesDir, name))

			m, err := Load(fsys)
			if err != nil {
				t.Fatalf("Load() error = %v, want nil", err)
			}

			encoded, err := json.MarshalIndent(goldenFile{Model: m, Navigation: m.Navigate(fsys)}, "", "  ")
			if err != nil {
				t.Fatalf("encoding the model failed: %v", err)
			}
			encoded = append(encoded, '\n')

			golden := filepath.Join(fixturesDir, name, "model.golden.json")

			if update {
				if err := os.WriteFile(golden, encoded, 0o644); err != nil {
					t.Fatalf("writing the golden failed: %v", err)
				}
				return
			}

			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("reading the golden failed: %v; run with %s=1 to create it", err, updateGoldenVar)
			}
			if !bytes.Equal(encoded, want) {
				t.Errorf("the model JSON differs from %s.\n"+
					"If the change is intended, rerun with %s=1 and read the diff before committing it.\n"+
					"got:\n%s", golden, updateGoldenVar, encoded)
			}
		})
	}
}
