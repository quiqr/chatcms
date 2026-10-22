package sitebuild

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/quiqr/chatcms/internal/runner"
)

const (
	Output             = "public"
	DefaultTimeout     = 10 * time.Minute
	DefaultMemoryBytes = 2 << 30
	cacheInside        = "/tmp/hugo-cache"
	defaultEnvironment = "production"
)

type Limits struct {
	Timeout     time.Duration
	MemoryBytes int64
}

func (l Limits) withDefaults() Limits {
	if l.Timeout <= 0 {
		l.Timeout = DefaultTimeout
	}
	if l.MemoryBytes <= 0 {
		l.MemoryBytes = DefaultMemoryBytes
	}
	return l
}

type Options struct {
	BaseURL     string
	Environment string
	Drafts      bool
}

type Builder struct {
	runner runner.Runner
	hugo   string
	limits Limits
}

var ErrNoHugo = errors.New("sitebuild: no Hugo was named")

func New(r runner.Runner, hugoBin string, limits Limits) (*Builder, error) {
	if r == nil {
		return nil, errors.New("sitebuild: no runner")
	}
	if hugoBin == "" {
		return nil, ErrNoHugo
	}
	return &Builder{runner: r, hugo: hugoBin, limits: limits.withDefaults()}, nil
}

func (b *Builder) Build(ctx context.Context, site fs.FS, opts Options) (*runner.Result, error) {
	if site == nil {
		return nil, errors.New("sitebuild: no site to build")
	}

	environment := opts.Environment
	if environment == "" {
		environment = defaultEnvironment
	}

	arguments := []string{
		"--destination", Output,
		"--cleanDestinationDir",
		"--cacheDir", cacheInside,
		"--environment", environment,
		"--logLevel", "info",
		"--noBuildLock",
	}
	if opts.BaseURL != "" {
		arguments = append(arguments, "--baseURL", opts.BaseURL)
	}
	if opts.Drafts {
		arguments = append(arguments, "--buildDrafts")
	}

	result, err := b.runner.Run(ctx, runner.Task{
		Inputs:    site,
		Command:   b.hugo,
		Arguments: arguments,
		Environment: map[string]string{
			"HUGO_ENV":         environment,
			"HUGO_ENVIRONMENT": environment,
		},
		Tools:       []string{b.hugo},
		Outputs:     []string{Output},
		Timeout:     b.limits.Timeout,
		MemoryBytes: b.limits.MemoryBytes,
	})
	var missing *runner.MissingOutputError
	if errors.As(err, &missing) {
		return nil, fmt.Errorf("sitebuild: Hugo reported success but wrote no %s", Output)
	}
	return result, err
}

// Pages is the built site itself rather than the folder holding it. A caller
// publishing what was built wants the pages at the root, which is where they
// land on a server.
func Pages(result *runner.Result) fs.FS {
	if result == nil || result.Outputs == nil {
		return nil
	}
	pages, err := fs.Sub(result.Outputs, Output)
	if err != nil {
		return result.Outputs
	}
	return pages
}
