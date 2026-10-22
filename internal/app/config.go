package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	Listen          string
	BaseURL         string
	StateDir        string
	DatabaseURL     string
	ShutdownTimeout time.Duration
	LogFormat       string
	GitBin          string
	HugoBin         string
	SSHKeyPath      string
	SecretKeyFile   string
	SyncInterval    time.Duration
}

// SitesDir is where every site's working copy lives.
func (c Config) SitesDir() string {
	return filepath.Join(c.StateDir, "sites")
}

type Lookup func(string) (string, bool)

type ConfigError struct {
	Variable string
	Reason   string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("%s %s", e.Variable, e.Reason)
}

const (
	reasonMissing    = "is not set"
	reasonConflict   = "is set both directly and as a _FILE variable"
	reasonUnreadable = "names a file that cannot be read"
	reasonInvalid    = "is not valid"
)

func LoadConfig(lookup Lookup) (Config, error) {
	var cfg Config

	fields := []struct {
		name     string
		target   *string
		fallback string
		required bool
	}{
		{"CHATCMS_LISTEN", &cfg.Listen, "127.0.0.1:8080", false},
		{"CHATCMS_BASE_URL", &cfg.BaseURL, "http://localhost:8080", false},
		{"CHATCMS_STATE_DIR", &cfg.StateDir, "./.chatcms-state", false},
		{"CHATCMS_DATABASE_URL", &cfg.DatabaseURL, "", true},
		{"CHATCMS_LOG_FORMAT", &cfg.LogFormat, "auto", false},
		{"CHATCMS_GIT_BIN", &cfg.GitBin, "git", false},
		{"CHATCMS_HUGO_BIN", &cfg.HugoBin, "hugo", false},
		{"CHATCMS_SSH_KEY_PATH", &cfg.SSHKeyPath, "", false},
		{"CHATCMS_SECRET_KEY_FILE", &cfg.SecretKeyFile, "", true},
	}

	for _, field := range fields {
		value, err := resolve(lookup, field.name)
		if err != nil {
			return Config{}, err
		}
		if value == "" {
			if field.required {
				return Config{}, &ConfigError{Variable: field.name, Reason: reasonMissing}
			}
			value = field.fallback
		}
		*field.target = value
	}

	durations := []struct {
		name     string
		target   *time.Duration
		fallback string
	}{
		{"CHATCMS_SHUTDOWN_TIMEOUT", &cfg.ShutdownTimeout, "20s"},
		{"CHATCMS_SYNC_INTERVAL", &cfg.SyncInterval, "2m"},
	}

	for _, field := range durations {
		raw, err := resolve(lookup, field.name)
		if err != nil {
			return Config{}, err
		}
		if raw == "" {
			raw = field.fallback
		}
		value, err := time.ParseDuration(raw)
		if err != nil || value <= 0 {
			return Config{}, &ConfigError{Variable: field.name, Reason: reasonInvalid}
		}
		*field.target = value
	}

	switch cfg.LogFormat {
	case "auto", "json", "text":
	default:
		return Config{}, &ConfigError{Variable: "CHATCMS_LOG_FORMAT", Reason: reasonInvalid}
	}

	return cfg, nil
}

func resolve(lookup Lookup, name string) (string, error) {
	direct, hasDirect := lookup(name)
	path, hasFile := lookup(name + "_FILE")

	if hasDirect && direct != "" && hasFile && path != "" {
		return "", &ConfigError{Variable: name, Reason: reasonConflict}
	}

	if hasFile && path != "" {
		contents, err := os.ReadFile(path)
		if err != nil {
			return "", &ConfigError{Variable: name + "_FILE", Reason: reasonUnreadable}
		}
		return strings.TrimSuffix(string(contents), "\n"), nil
	}

	return direct, nil
}

func EnvLookup() Lookup {
	return os.LookupEnv
}
