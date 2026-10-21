package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func lookupFrom(pairs map[string]string) Lookup {
	return func(key string) (string, bool) {
		value, ok := pairs[key]
		return value, ok
	}
}

func TestLoadConfigAppliesDefaults(t *testing.T) {
	cfg, err := LoadConfig(lookupFrom(map[string]string{
		"CHATCMS_DATABASE_URL": "postgres:///chatcms",
	}))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}

	want := map[string]string{
		"Listen":    "127.0.0.1:8080",
		"BaseURL":   "http://localhost:8080",
		"StateDir":  "./.chatcms-state",
		"LogFormat": "auto",
		"GitBin":    "git",
		"HugoBin":   "hugo",
	}
	got := map[string]string{
		"Listen":    cfg.Listen,
		"BaseURL":   cfg.BaseURL,
		"StateDir":  cfg.StateDir,
		"LogFormat": cfg.LogFormat,
		"GitBin":    cfg.GitBin,
		"HugoBin":   cfg.HugoBin,
	}
	for field, value := range want {
		if got[field] != value {
			t.Errorf("%s = %q, want %q", field, got[field], value)
		}
	}

	if cfg.ShutdownTimeout != 20*time.Second {
		t.Errorf("ShutdownTimeout = %v, want 20s", cfg.ShutdownTimeout)
	}
}

func TestLoadConfigPrefersEnvironment(t *testing.T) {
	cfg, err := LoadConfig(lookupFrom(map[string]string{
		"CHATCMS_DATABASE_URL":     "postgres:///other",
		"CHATCMS_LISTEN":           "0.0.0.0:9000",
		"CHATCMS_BASE_URL":         "https://chatcms.eu",
		"CHATCMS_STATE_DIR":        "/var/lib/chatcms",
		"CHATCMS_GIT_BIN":          "/nix/store/git",
		"CHATCMS_HUGO_BIN":         "/nix/store/hugo",
		"CHATCMS_LOG_FORMAT":       "json",
		"CHATCMS_SHUTDOWN_TIMEOUT": "45s",
	}))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}

	if cfg.Listen != "0.0.0.0:9000" || cfg.BaseURL != "https://chatcms.eu" {
		t.Errorf("environment not applied: %+v", cfg)
	}
	if cfg.StateDir != "/var/lib/chatcms" || cfg.GitBin != "/nix/store/git" {
		t.Errorf("environment not applied: %+v", cfg)
	}
	if cfg.HugoBin != "/nix/store/hugo" || cfg.LogFormat != "json" {
		t.Errorf("environment not applied: %+v", cfg)
	}
	if cfg.ShutdownTimeout != 45*time.Second {
		t.Errorf("ShutdownTimeout = %v, want 45s", cfg.ShutdownTimeout)
	}
}

func TestLoadConfigRequiresDatabaseURL(t *testing.T) {
	_, err := LoadConfig(lookupFrom(nil))

	var configErr *ConfigError
	if !errors.As(err, &configErr) {
		t.Fatalf("LoadConfig() error = %v, want a *ConfigError", err)
	}
	if configErr.Variable != "CHATCMS_DATABASE_URL" {
		t.Errorf("Variable = %q, want CHATCMS_DATABASE_URL", configErr.Variable)
	}
}

func TestLoadConfigErrorNamesVariableNotValue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "url")
	if err := os.WriteFile(path, []byte("postgres://secret@host/db"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig(lookupFrom(map[string]string{
		"CHATCMS_DATABASE_URL_FILE": path,
		"CHATCMS_LOG_FORMAT":        "yelling",
	}))
	if err == nil {
		t.Fatal("LoadConfig() error = nil, want an error for the log format")
	}
	if !strings.Contains(err.Error(), "CHATCMS_LOG_FORMAT") {
		t.Errorf("error = %q, want it to name the variable", err)
	}
	if strings.Contains(err.Error(), "secret") {
		t.Errorf("error = %q, want it to carry no value", err)
	}
}

func TestLoadConfigReadsSecretFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "database-url")
	if err := os.WriteFile(path, []byte("postgres:///from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(lookupFrom(map[string]string{
		"CHATCMS_DATABASE_URL_FILE": path,
	}))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}
	if cfg.DatabaseURL != "postgres:///from-file" {
		t.Errorf("DatabaseURL = %q, want the file contents without the trailing newline", cfg.DatabaseURL)
	}
}

func TestLoadConfigRejectsBothValueAndFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "database-url")
	if err := os.WriteFile(path, []byte("postgres:///from-file"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig(lookupFrom(map[string]string{
		"CHATCMS_DATABASE_URL":      "postgres:///direct",
		"CHATCMS_DATABASE_URL_FILE": path,
	}))

	var configErr *ConfigError
	if !errors.As(err, &configErr) {
		t.Fatalf("LoadConfig() error = %v, want a *ConfigError", err)
	}
	if !strings.Contains(configErr.Error(), "both") {
		t.Errorf("error = %q, want it to say both forms are set", configErr)
	}
}

func TestLoadConfigRejectsUnreadableSecretFile(t *testing.T) {
	_, err := LoadConfig(lookupFrom(map[string]string{
		"CHATCMS_DATABASE_URL_FILE": filepath.Join(t.TempDir(), "absent"),
	}))

	var configErr *ConfigError
	if !errors.As(err, &configErr) {
		t.Fatalf("LoadConfig() error = %v, want a *ConfigError", err)
	}
	if configErr.Variable != "CHATCMS_DATABASE_URL_FILE" {
		t.Errorf("Variable = %q, want the _FILE variable", configErr.Variable)
	}
}

func TestLoadConfigRejectsInvalidShutdownTimeout(t *testing.T) {
	for name, value := range map[string]string{
		"unparseable": "soon",
		"zero":        "0s",
		"negative":    "-5s",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := LoadConfig(lookupFrom(map[string]string{
				"CHATCMS_DATABASE_URL":     "postgres:///chatcms",
				"CHATCMS_SHUTDOWN_TIMEOUT": value,
			}))
			if err == nil {
				t.Fatalf("LoadConfig() with %q error = nil, want an error", value)
			}
		})
	}
}

func TestLoadConfigReadsSyncSettings(t *testing.T) {
	cfg, err := LoadConfig(lookupFrom(map[string]string{
		"CHATCMS_DATABASE_URL":  "postgres:///chatcms",
		"CHATCMS_SYNC_INTERVAL": "45s",
		"CHATCMS_SSH_KEY_PATH":  "/run/secrets/site-key",
	}))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}
	if cfg.SyncInterval != 45*time.Second {
		t.Errorf("SyncInterval = %v, want 45s", cfg.SyncInterval)
	}
	if cfg.SSHKeyPath != "/run/secrets/site-key" {
		t.Errorf("SSHKeyPath = %q", cfg.SSHKeyPath)
	}
}

func TestSyncIntervalHasADefault(t *testing.T) {
	cfg, err := LoadConfig(lookupFrom(map[string]string{
		"CHATCMS_DATABASE_URL": "postgres:///chatcms",
	}))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}
	if cfg.SyncInterval <= 0 {
		t.Errorf("SyncInterval = %v, want a usable default", cfg.SyncInterval)
	}
}

func TestLoadConfigRejectsAnInvalidSyncInterval(t *testing.T) {
	for _, value := range []string{"never", "0s", "-1m"} {
		_, err := LoadConfig(lookupFrom(map[string]string{
			"CHATCMS_DATABASE_URL":  "postgres:///chatcms",
			"CHATCMS_SYNC_INTERVAL": value,
		}))
		if err == nil {
			t.Errorf("LoadConfig() with %q error = nil, want a rejection", value)
		}
	}
}

func TestSitesDirIsUnderTheStateDirectory(t *testing.T) {
	cfg := Config{StateDir: "/var/lib/chatcms"}

	if got := cfg.SitesDir(); got != "/var/lib/chatcms/sites" {
		t.Errorf("SitesDir() = %q, want it under the state directory", got)
	}
}

func TestHostOf(t *testing.T) {
	cases := map[string]string{
		"https://chatcms.eu":      "chatcms.eu",
		"http://localhost:8080":   "localhost",
		"https://CHATCMS.EU/path": "chatcms.eu",
		"not a url":               "chatcms.local",
		"":                        "chatcms.local",
	}

	for in, want := range cases {
		if got := hostOf(in); got != want {
			t.Errorf("hostOf(%q) = %q, want %q", in, got, want)
		}
	}
}
