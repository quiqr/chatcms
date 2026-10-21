package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/quiqr/chatcms/internal/version"
)

type Config struct {
	Listen      string
	BaseURL     string
	StateDir    string
	DatabaseURL string
	GitBin      string
	HugoBin     string
}

var ErrMissingDatabaseURL = errors.New("CHATCMS_DATABASE_URL is not set")

func LoadConfig(lookup func(string) string) (Config, error) {
	cfg := Config{
		Listen:      withDefault(lookup("CHATCMS_LISTEN"), "127.0.0.1:8080"),
		BaseURL:     withDefault(lookup("CHATCMS_BASE_URL"), "http://localhost:8080"),
		StateDir:    withDefault(lookup("CHATCMS_STATE_DIR"), "./.chatcms-state"),
		DatabaseURL: lookup("CHATCMS_DATABASE_URL"),
		GitBin:      withDefault(lookup("CHATCMS_GIT_BIN"), "git"),
		HugoBin:     withDefault(lookup("CHATCMS_HUGO_BIN"), "hugo"),
	}

	if cfg.DatabaseURL == "" {
		return cfg, ErrMissingDatabaseURL
	}

	return cfg, nil
}

func withDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func Run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("chatcmsd", flag.ContinueOnError)
	fs.SetOutput(stderr)
	showVersion := fs.Bool("version", false, "print the version and exit")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *showVersion {
		fmt.Fprintln(stdout, version.String())
		return 0
	}

	cfg, err := LoadConfig(os.Getenv)
	if err != nil {
		fmt.Fprintf(stderr, "chatcmsd: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "%s\nlisten=%s base-url=%s state-dir=%s\n",
		version.String(), cfg.Listen, cfg.BaseURL, cfg.StateDir)
	fmt.Fprintln(stderr, "chatcmsd: no HTTP server yet; see openspec/changes for the active work")
	return 0
}
