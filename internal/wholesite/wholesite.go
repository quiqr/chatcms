// Package wholesite is what can be done to a site as one thing: copied,
// moved, downloaded, archived. Everything here is about a whole site rather
// than anything inside one, which is why it is not in content or in publish.
package wholesite

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/gitsync"
)

// Sites is what this package needs from the working copies: one to read, and
// a way to make another.
type Sites interface {
	Open(ctx context.Context, id string) (gitsync.Repo, error)
	Create(ctx context.Context, id, remote string, content fs.FS, author gitsync.Identity, message string) (gitsync.Repo, error)
}

// Copier makes one site from another.
type Copier struct {
	sites Sites
}

func NewCopier(sites Sites) *Copier { return &Copier{sites: sites} }

// CopyTo makes a new working copy holding what the original holds, and
// nothing else. The remote is usually empty: a copy that published where the
// original publishes would overwrite a live website the first time anybody
// pressed Publish, so where the copy publishes is set afterwards, by
// somebody who meant it.
func (c *Copier) CopyTo(ctx context.Context, from, to uuid.UUID, remote string, author gitsync.Identity, message string) error {
	original, err := c.sites.Open(ctx, from.String())
	if err != nil {
		return fmt.Errorf("opening the site to copy failed: %w", err)
	}
	defer func() { _ = original.Close() }()

	copied, err := c.sites.Create(ctx, to.String(), remote, withoutGit{original.FS()}, author, message)
	if err != nil {
		return fmt.Errorf("making the copy failed: %w", err)
	}
	return copied.Close()
}

// Zip writes the site's files as a zip. It is the working copy without the
// directory git keeps its own state in: what a site is made of is its pages,
// its media and its model, and naming three folders would be wrong about the
// fourth the first time a template used one.
func (c *Copier) Zip(ctx context.Context, projectID uuid.UUID, into io.Writer) error {
	repository, err := c.sites.Open(ctx, projectID.String())
	if err != nil {
		return fmt.Errorf("opening the site failed: %w", err)
	}
	defer func() { _ = repository.Close() }()

	writer := zip.NewWriter(into)
	err = fs.WalkDir(repository.FS(), ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if isGit(name) {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}

		content, err := fs.ReadFile(repository.FS(), name)
		if err != nil {
			return err
		}
		file, err := writer.Create(name)
		if err != nil {
			return err
		}
		_, err = file.Write(content)
		return err
	})
	if err != nil {
		_ = writer.Close()
		return fmt.Errorf("reading the site failed: %w", err)
	}
	return writer.Close()
}

func isGit(name string) bool {
	for _, part := range strings.Split(path.Clean(name), "/") {
		if part == ".git" {
			return true
		}
	}
	return false
}

// withoutGit hides the directory git keeps its own state in, so that a copy
// made from a working copy carries neither the original's history nor its
// idea of where it publishes.
type withoutGit struct{ fs.FS }

func (w withoutGit) Open(name string) (fs.File, error) {
	if isGit(name) {
		return nil, fs.ErrNotExist
	}
	return w.FS.Open(name)
}
