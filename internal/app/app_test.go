package app

import (
	"bytes"
	"strings"
	"testing"
)

func envFrom(pairs map[string]string) func(string) string {
	return func(key string) string { return pairs[key] }
}

func TestLoadConfigAppliesDefaults(t *testing.T) {
	cfg, err := LoadConfig(envFrom(map[string]string{
		"CHATCMS_DATABASE_URL": "postgres:///chatcms",
	}))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}

	cases := map[string]string{
		"Listen":   "127.0.0.1:8080",
		"BaseURL":  "http://localhost:8080",
		"StateDir": "./.chatcms-state",
		"GitBin":   "git",
		"HugoBin":  "hugo",
	}
	got := map[string]string{
		"Listen":   cfg.Listen,
		"BaseURL":  cfg.BaseURL,
		"StateDir": cfg.StateDir,
		"GitBin":   cfg.GitBin,
		"HugoBin":  cfg.HugoBin,
	}
	for field, want := range cases {
		if got[field] != want {
			t.Errorf("%s = %q, want %q", field, got[field], want)
		}
	}
}

func TestLoadConfigPrefersEnvironment(t *testing.T) {
	cfg, err := LoadConfig(envFrom(map[string]string{
		"CHATCMS_DATABASE_URL": "postgres:///other",
		"CHATCMS_LISTEN":       "0.0.0.0:9000",
		"CHATCMS_BASE_URL":     "https://chatcms.eu",
		"CHATCMS_STATE_DIR":    "/var/lib/chatcms",
		"CHATCMS_GIT_BIN":      "/nix/store/git",
		"CHATCMS_HUGO_BIN":     "/nix/store/hugo",
	}))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}
	if cfg.Listen != "0.0.0.0:9000" || cfg.BaseURL != "https://chatcms.eu" {
		t.Errorf("environment not applied: %+v", cfg)
	}
	if cfg.StateDir != "/var/lib/chatcms" || cfg.GitBin != "/nix/store/git" || cfg.HugoBin != "/nix/store/hugo" {
		t.Errorf("environment not applied: %+v", cfg)
	}
}

func TestLoadConfigRequiresDatabaseURL(t *testing.T) {
	_, err := LoadConfig(envFrom(nil))
	if err != ErrMissingDatabaseURL {
		t.Errorf("LoadConfig() error = %v, want %v", err, ErrMissingDatabaseURL)
	}
}

func TestRunVersionFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer

	if code := Run([]string{"-version"}, &stdout, &stderr); code != 0 {
		t.Fatalf("Run() = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "chatcmsd") {
		t.Errorf("stdout = %q, want it to name the binary", stdout.String())
	}
}

func TestRunRejectsUnknownFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer

	if code := Run([]string{"-nope"}, &stdout, &stderr); code != 2 {
		t.Errorf("Run() = %d, want 2", code)
	}
}

func TestRunFailsWithoutDatabaseURL(t *testing.T) {
	t.Setenv("CHATCMS_DATABASE_URL", "")
	var stdout, stderr bytes.Buffer

	if code := Run(nil, &stdout, &stderr); code != 1 {
		t.Errorf("Run() = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "CHATCMS_DATABASE_URL") {
		t.Errorf("stderr = %q, want it to name the missing variable", stderr.String())
	}
}

func TestRunReportsConfiguration(t *testing.T) {
	t.Setenv("CHATCMS_DATABASE_URL", "postgres:///chatcms")
	t.Setenv("CHATCMS_LISTEN", "127.0.0.1:7000")
	var stdout, stderr bytes.Buffer

	if code := Run(nil, &stdout, &stderr); code != 0 {
		t.Fatalf("Run() = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "127.0.0.1:7000") {
		t.Errorf("stdout = %q, want it to report the listen address", stdout.String())
	}
}
