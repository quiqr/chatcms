package db

import (
	"io/fs"
	"strings"
	"testing"
)

func TestMigrationsAreEmbedded(t *testing.T) {
	entries, err := fs.ReadDir(Migrations(), ".")
	if err != nil {
		t.Fatalf("reading the embedded migrations failed: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("no migrations are embedded")
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".sql") {
			t.Errorf("%s is embedded but is not a .sql migration", entry.Name())
		}
	}
}

func TestEmbeddedMigrationsCarryBothDirections(t *testing.T) {
	entries, _ := fs.ReadDir(Migrations(), ".")

	for _, entry := range entries {
		contents, err := fs.ReadFile(Migrations(), entry.Name())
		if err != nil {
			t.Fatalf("reading %s failed: %v", entry.Name(), err)
		}
		for _, marker := range []string{"-- +goose Up", "-- +goose Down"} {
			if !strings.Contains(string(contents), marker) {
				t.Errorf("%s has no %q section", entry.Name(), marker)
			}
		}
	}
}
