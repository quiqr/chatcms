package app

import (
	"fmt"
	"os"
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

	timeout, err := resolve(lookup, "CHATCMS_SHUTDOWN_TIMEOUT")
	if err != nil {
		return Config{}, err
	}
	if timeout == "" {
		timeout = "20s"
	}
	cfg.ShutdownTimeout, err = time.ParseDuration(timeout)
	if err != nil {
		return Config{}, &ConfigError{Variable: "CHATCMS_SHUTDOWN_TIMEOUT", Reason: reasonInvalid}
	}
	if cfg.ShutdownTimeout <= 0 {
		return Config{}, &ConfigError{Variable: "CHATCMS_SHUTDOWN_TIMEOUT", Reason: reasonInvalid}
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
