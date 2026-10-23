package gitsync

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SyncJobKind names the background work that brings a site level with its
// remote. It lives here so the queue never has to know what a sync is.
const SyncJobKind = "site-sync"

// ErrUnusableSiteID says an identifier cannot be turned into a directory
// name. Site identifiers come from the application, not from a person, so
// this is a programming error rather than something to show anyone.
var ErrUnusableSiteID = errors.New("this site identifier cannot be used as a directory name")

// KeyLookup gives the path of the key a site reaches its remote with. Each
// site has its own, so one customer's key never reaches another's host.
type KeyLookup func(ctx context.Context, siteID string) (string, error)

// Sites holds every working copy under one root, so that nothing outside this
// package builds a path to one.
type Sites struct {
	cfg    Config
	root   string
	keyFor KeyLookup
}

// NewSites builds the registry. With no lookup, every site uses whatever key
// the configuration names, which is how the tests run against repositories on
// disk that need none.
func NewSites(cfg Config, root string, keyFor KeyLookup) *Sites {
	return &Sites{cfg: cfg, root: root, keyFor: keyFor}
}

// configFor is the configuration for one site, carrying that site's own key.
func (s *Sites) configFor(ctx context.Context, id string) (Config, error) {
	cfg := s.cfg
	if s.keyFor == nil {
		return cfg, nil
	}

	path, err := s.keyFor(ctx, id)
	if err != nil {
		return Config{}, err
	}
	cfg.SSHKeyPath = path
	return cfg, nil
}

// Dir is where a site's working copy lives.
func (s *Sites) Dir(id string) (string, error) {
	if err := checkSiteID(id); err != nil {
		return "", err
	}
	return filepath.Join(s.root, id, "working-copy"), nil
}

// Open reads an existing working copy.
func (s *Sites) Open(ctx context.Context, id string) (Repo, error) {
	dir, err := s.Dir(id)
	if err != nil {
		return nil, err
	}
	cfg, err := s.configFor(ctx, id)
	if err != nil {
		return nil, err
	}
	return Open(ctx, cfg, dir)
}

// Clone creates a working copy for a site that has none yet.
func (s *Sites) Clone(ctx context.Context, id, remote string) (Repo, error) {
	dir, err := s.Dir(id)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(dir), 0o700); err != nil {
		return nil, fmt.Errorf("preparing the site directory failed: %w", err)
	}
	cfg, err := s.configFor(ctx, id)
	if err != nil {
		return nil, err
	}
	return Clone(ctx, cfg, remote, dir)
}

// Create makes a working copy from a directory of files rather than from a
// remote: an export from the Quiqr desktop app arrives this way. The content
// is committed and the remote is set, so everything afterwards is the same as
// for a site that was cloned.
func (s *Sites) Create(ctx context.Context, id, remote string, content fs.FS, author Identity, message string) (Repo, error) {
	dir, err := s.Dir(id)
	if err != nil {
		return nil, err
	}
	if isWorkingCopy(dir) {
		return nil, ErrAlreadyAWorkingCopy
	}

	cfg, err := s.configFor(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("preparing the site directory failed: %w", err)
	}

	home, err := gitHome(dir)
	if err != nil {
		return nil, err
	}
	runner := newRunner(cfg, dir, home)

	if _, err := runner.run(ctx, "init", "--initial-branch=main"); err != nil {
		return nil, err
	}
	if remote != "" {
		if _, err := runner.run(ctx, "remote", "add", "origin", remote); err != nil {
			return nil, err
		}
	}

	if err := copyInto(dir, content); err != nil {
		return nil, err
	}

	repository, err := open(ctx, cfg, dir)
	if err != nil {
		return nil, err
	}

	if _, err := repository.Save(ctx, Save{
		Message: message,
		Author:  author,
		Writes:  map[string][]byte{},
	}); err != nil && !errors.Is(err, ErrNothingChanged) {
		repository.Close()
		return nil, err
	}

	return repository, nil
}

// copyInto writes a tree of files into a working copy, refusing anything that
// would escape it or that belongs to version control.
func copyInto(dir string, content fs.FS) error {
	return fs.WalkDir(content, ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if name == "." {
			return nil
		}
		if name == versionControlDir || strings.HasPrefix(name, versionControlDir+"/") {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		target := filepath.Join(dir, filepath.FromSlash(name))

		if entry.IsDir() {
			return os.MkdirAll(target, 0o700)
		}

		data, err := fs.ReadFile(content, name)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o600)
	})
}

// Reachable says whether a site's key reaches a remote. It is the question an
// implementer asks after pasting the key into the git host.
func (s *Sites) Reachable(ctx context.Context, id, remote string) error {
	cfg, err := s.configFor(ctx, id)
	if err != nil {
		return err
	}

	home, err := gitHome(filepath.Join(s.root, id, "working-copy"))
	if err != nil {
		return err
	}

	if _, err := newRunner(cfg, home, home).run(ctx, "ls-remote", "--heads", "--", remote); err != nil {
		return classifyRemote(err)
	}
	return nil
}

// ErrKeyRefused says the remote answered and would not take the key, which is
// what every first setup looks like until the key is added to the git host.
var ErrKeyRefused = errors.New("the remote refused this key")

func classifyRemote(err error) error {
	var opErr *OperationError
	if !errors.As(err, &opErr) {
		return err
	}

	lowered := strings.ToLower(opErr.Detail)
	for _, sign := range []string{
		"permission denied", "publickey", "access denied",
		"repository not found", "could not read from remote repository",
		"authentication failed",
	} {
		if strings.Contains(lowered, sign) {
			return fmt.Errorf("%w: %s", ErrKeyRefused, opErr.Detail)
		}
	}
	return fmt.Errorf("%w: %s", ErrRemoteUnreachable, opErr.Detail)
}

// List names every site that has a working copy. Until projects exist this is
// the only list of sites there is.
func (s *Sites) List() ([]string, error) {
	entries, err := os.ReadDir(s.root)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading the sites directory failed: %w", err)
	}

	var ids []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir, err := s.Dir(entry.Name())
		if err != nil {
			continue
		}
		if isWorkingCopy(dir) {
			ids = append(ids, entry.Name())
		}
	}
	sort.Strings(ids)
	return ids, nil
}

func checkSiteID(id string) error {
	if id == "" || id == "." || id == ".." {
		return fmt.Errorf("%w: %q", ErrUnusableSiteID, id)
	}
	if strings.ContainsAny(id, `/\:`+"\x00") {
		return fmt.Errorf("%w: %q", ErrUnusableSiteID, id)
	}
	if strings.HasPrefix(id, ".") {
		return fmt.Errorf("%w: %q", ErrUnusableSiteID, id)
	}
	return nil
}

// InitBareForDevelopment makes an empty remote on disk, so that a developer
// can exercise publishing without owning a git host. It is here rather than
// in the development tool because every git invocation in this codebase
// originates in this package, which docs/ARCHITECTURE.md requires and a test
// over the sources enforces.
func InitBareForDevelopment(ctx context.Context, cfg Config, path string) error {
	if _, err := os.Stat(filepath.Join(path, "HEAD")); err == nil {
		return nil
	}
	home, err := gitHome(filepath.Dir(path))
	if err != nil {
		return err
	}
	_, err = newRunner(cfg, filepath.Dir(path), home).run(ctx,
		"init", "--bare", "--initial-branch=main", path)
	return err
}
