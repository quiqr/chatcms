package gitsync

import (
	"context"
	"errors"
	"fmt"
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

// Sites holds every working copy under one root, so that nothing outside this
// package builds a path to one.
type Sites struct {
	cfg  Config
	root string
}

func NewSites(cfg Config, root string) *Sites {
	return &Sites{cfg: cfg, root: root}
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
	return Open(ctx, s.cfg, dir)
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
	return Clone(ctx, s.cfg, remote, dir)
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
