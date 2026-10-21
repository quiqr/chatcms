package gitsync

import (
	"io/fs"
	"os"
	"path"
	"strings"
)

// contentFS serves a site's content and nothing else: version control's own
// directory is invisible, and a path two people changed is refused rather
// than served with markers in it.
type contentFS struct {
	root string
	repo *repo
}

func (r *repo) FS() fs.FS {
	return &contentFS{root: r.dir, repo: r}
}

const versionControlDir = ".git"

func (c *contentFS) conflicted(name string) bool {
	c.repo.mu.Lock()
	defer c.repo.mu.Unlock()
	return c.repo.conflicted[name]
}

func (c *contentFS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	if hidden(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	if c.conflicted(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: ErrPathInConflict}
	}

	file, err := os.Open(path.Join(c.root, name))
	if err != nil {
		return nil, err
	}
	if name == "." {
		return &contentDir{File: file, fs: c, name: name}, nil
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	if info.IsDir() {
		return &contentDir{File: file, fs: c, name: name}, nil
	}
	return file, nil
}

func (c *contentFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if !fs.ValidPath(name) || hidden(name) {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
	}

	entries, err := os.ReadDir(path.Join(c.root, name))
	if err != nil {
		return nil, err
	}

	kept := make([]fs.DirEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.Name() == versionControlDir {
			continue
		}
		kept = append(kept, entry)
	}
	return kept, nil
}

// contentDir hides version control's directory from a listing taken through
// Open rather than through ReadDir.
type contentDir struct {
	*os.File
	fs   *contentFS
	name string
}

func (d *contentDir) ReadDir(n int) ([]fs.DirEntry, error) {
	entries, err := d.File.ReadDir(n)
	if err != nil {
		return entries, err
	}

	kept := make([]fs.DirEntry, 0, len(entries))
	for _, entry := range entries {
		if d.name == "." && entry.Name() == versionControlDir {
			continue
		}
		kept = append(kept, entry)
	}
	return kept, nil
}

func hidden(name string) bool {
	return name == versionControlDir || strings.HasPrefix(name, versionControlDir+"/")
}
