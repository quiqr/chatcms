package gitsync

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Repo is a site's working copy. It speaks in sites and versions, never in
// commits and branches, so that a different version control system could
// implement it. See docs/ARCHITECTURE.md.
type Repo interface {
	// FS returns the site's content as a read-only filesystem.
	FS() fs.FS
	// Save writes files and makes one version of them.
	Save(ctx context.Context, save Save) (Version, error)
	// History lists versions, newest first.
	History(ctx context.Context, limit int) ([]Version, error)
	// Status answers what is saved, what is unpublished and what arrived.
	Status(ctx context.Context) (Status, error)
	// Sync brings the remote's versions in and replays local ones on top.
	Sync(ctx context.Context) (SyncResult, error)
	// Publish sends local versions to the remote.
	Publish(ctx context.Context) (PublishResult, error)
	// PublishTo sends them to a named line on the remote instead. An empty
	// name means the line the site is edited on.
	PublishTo(ctx context.Context, line string) (PublishResult, error)
	// PublishPages sends built pages to a line of their own, without making
	// a version of the site's own files and without touching the working
	// copy, so somebody can keep editing while it happens.
	PublishPages(ctx context.Context, line string, pages fs.FS, author Identity, message string) (PublishResult, error)
	// Changes lists what differs between the site as people see it and the
	// site as it is now.
	Changes(ctx context.Context) ([]Change, error)
	// ReadPublished reads a file as the published site has it.
	ReadPublished(ctx context.Context, name string) ([]byte, error)
	// GoBackTo makes a new version whose content is an earlier version's,
	// throwing nothing away, so that going back can itself be undone.
	GoBackTo(ctx context.Context, version string, author Identity, message string) (Version, error)
	// ChangesIn says what one version changed.
	ChangesIn(ctx context.Context, version string) ([]Change, error)
	// At is the site as one version left it, read only.
	At(ctx context.Context, version string) (fs.FS, error)
	// Before is the site as it was just before one version.
	Before(ctx context.Context, version string) (fs.FS, error)
	// Resolve answers an unresolved difference and continues the exchange.
	Resolve(ctx context.Context, resolutions map[string][]byte) (SyncResult, error)
	// Abandon gives up an exchange and restores the site as it was.
	Abandon(ctx context.Context) error
	// Close releases the working copy.
	Close() error
}

type repo struct {
	cfg    Config
	dir    string
	git    *runner
	branch string

	mu sync.Mutex

	// conflicted is the set of paths an unfinished exchange is stuck on, and
	// resolved remembers answers within one exchange so the same question is
	// not asked twice during a replay.
	conflicted map[string]bool
	resolved   map[string][]byte
}

var (
	locksMu sync.Mutex
	locks   = map[string]*sync.Mutex{}
)

// siteLock serialises writes to one working copy. Two sites never wait for
// each other. See design.md for why this is per process.
func siteLock(dir string) *sync.Mutex {
	locksMu.Lock()
	defer locksMu.Unlock()

	if lock, ok := locks[dir]; ok {
		return lock
	}
	lock := &sync.Mutex{}
	locks[dir] = lock
	return lock
}

// Clone creates a working copy of a site from its remote. The directory must
// not already hold one.
func Clone(ctx context.Context, cfg Config, remote, dir string) (Repo, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	if isWorkingCopy(dir) {
		return nil, ErrAlreadyAWorkingCopy
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("preparing the site directory failed: %w", err)
	}

	home, err := gitHome(dir)
	if err != nil {
		return nil, err
	}

	if _, err := newRunner(cfg, filepath.Dir(dir), home).run(ctx, "clone", "--", remote, dir); err != nil {
		return nil, unreachable(err)
	}
	return open(ctx, cfg, dir)
}

// Open reads an existing working copy. It never touches the remote.
func Open(ctx context.Context, cfg Config, dir string) (Repo, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	if !isWorkingCopy(dir) {
		return nil, ErrNotAWorkingCopy
	}
	return open(ctx, cfg, dir)
}

func open(ctx context.Context, cfg Config, dir string) (Repo, error) {
	home, err := gitHome(dir)
	if err != nil {
		return nil, err
	}

	r := &repo{
		cfg:        cfg,
		dir:        dir,
		git:        newRunner(cfg, dir, home),
		conflicted: map[string]bool{},
		resolved:   map[string][]byte{},
	}

	branch, err := r.currentBranch(ctx)
	if err != nil {
		return nil, err
	}
	r.branch = branch

	if err := r.recoverConflicts(ctx); err != nil {
		return nil, err
	}
	return r, nil
}

func (c Config) validate() error {
	if c.GitBin == "" {
		return errors.New("gitsync: no git binary is configured")
	}
	if !c.System.valid() {
		return errors.New("gitsync: no system identity is configured")
	}
	return nil
}

// gitHome keeps git away from the server's own configuration by giving it a
// home of its own next to the working copy.
func gitHome(dir string) (string, error) {
	home := filepath.Join(filepath.Dir(dir), ".githome")
	if err := os.MkdirAll(home, 0o700); err != nil {
		return "", fmt.Errorf("preparing the git home failed: %w", err)
	}
	return home, nil
}

func isWorkingCopy(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil && (info.IsDir() || info.Mode().IsRegular())
}

func (r *repo) currentBranch(ctx context.Context) (string, error) {
	out, err := r.git.run(ctx, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err == nil {
		return strings.TrimSpace(out), nil
	}
	// A repository with no versions yet has an unborn HEAD, which symbolic-ref
	// reports with --quiet as a failure rather than a message.
	out, err = r.git.run(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func (r *repo) Close() error {
	return nil
}

func unreachable(err error) error {
	var opErr *OperationError
	if errors.As(err, &opErr) {
		lowered := strings.ToLower(opErr.Detail)
		for _, sign := range []string{
			"could not resolve host", "connection refused", "connection timed out",
			"network is unreachable", "no route to host", "could not read from remote",
			"repository not found", "permission denied", "does not appear to be a git repository",
		} {
			if strings.Contains(lowered, sign) {
				return fmt.Errorf("%w: %s", ErrRemoteUnreachable, opErr.Detail)
			}
		}
	}
	return err
}
