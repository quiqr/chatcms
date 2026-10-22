package gitsync

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

func (r *repo) Save(ctx context.Context, save Save) (Version, error) {
	if save.Message == "" {
		return Version{}, errors.New("a version needs a message saying what changed")
	}
	if !save.Author.valid() {
		return Version{}, errors.New("a version needs the person who made it")
	}

	paths := make([]string, 0, len(save.Writes)+len(save.Deletes))
	for name := range save.Writes {
		paths = append(paths, name)
	}
	paths = append(paths, save.Deletes...)
	sort.Strings(paths)

	for _, name := range paths {
		if err := r.checkPath(name); err != nil {
			return Version{}, err
		}
	}

	lock := siteLock(r.dir)
	lock.Lock()
	defer lock.Unlock()

	for name, contents := range save.Writes {
		full := filepath.Join(r.dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
			return Version{}, fmt.Errorf("preparing the folder for %s failed: %w", name, err)
		}
		if err := os.WriteFile(full, contents, 0o600); err != nil {
			return Version{}, fmt.Errorf("writing %s failed: %w", name, err)
		}
	}

	for _, name := range save.Deletes {
		full := filepath.Join(r.dir, filepath.FromSlash(name))
		if err := os.RemoveAll(full); err != nil {
			return Version{}, fmt.Errorf("removing %s failed: %w", name, err)
		}
	}

	// A save that names no file saves everything the working copy holds,
	// which is what importing a site is: the files are already there and all
	// of them belong in the first version.
	everything := len(paths) == 0

	addArgs := []string{"add", "--all"}
	if !everything {
		addArgs = append(append(addArgs, "--"), paths...)
	}
	if _, err := r.git.run(ctx, addArgs...); err != nil {
		return Version{}, err
	}

	if r.git.succeeds(ctx, "diff", "--cached", "--quiet") {
		return Version{}, ErrNothingChanged
	}

	commit := []string{
		"commit",
		"--author=" + save.Author.String(),
		"--message=" + save.Message,
	}
	if !everything {
		commit = append(commit, "--only", "--")
		commit = append(commit, paths...)
	}
	if _, err := r.git.run(ctx, commit...); err != nil {
		return Version{}, err
	}

	versions, err := r.History(ctx, 1)
	if err != nil {
		return Version{}, err
	}
	if len(versions) == 0 {
		return Version{}, errors.New("the version was written but could not be read back")
	}
	return versions[0], nil
}

// checkPath refuses anything that would write outside the site, and anything
// whose name would make history unparseable.
func (r *repo) checkPath(name string) error {
	if name == "" || name == "." {
		return fmt.Errorf("%w: %q", ErrEscapingPath, name)
	}
	if strings.ContainsAny(name, "\x00\n\r") {
		return fmt.Errorf("%w: a file name cannot contain a line break", ErrEscapingPath)
	}
	if filepath.IsAbs(name) || strings.HasPrefix(name, "/") {
		return fmt.Errorf("%w: %q", ErrEscapingPath, name)
	}
	if name == versionControlDir || strings.HasPrefix(name, versionControlDir+"/") {
		return fmt.Errorf("%w: %q", ErrEscapingPath, name)
	}

	cleaned := path.Clean(name)
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return fmt.Errorf("%w: %q", ErrEscapingPath, name)
	}

	// A symlink anywhere on the way to the file can point out of the site, so
	// the resolved location is checked rather than the spelling.
	root, err := filepath.EvalSymlinks(r.dir)
	if err != nil {
		return fmt.Errorf("reading the site directory failed: %w", err)
	}

	target := filepath.Join(r.dir, filepath.FromSlash(cleaned))
	resolved := target
	for {
		link, err := filepath.EvalSymlinks(resolved)
		if err == nil {
			resolved = link
			break
		}
		parent := filepath.Dir(resolved)
		if parent == resolved {
			break
		}
		resolved = parent
	}

	relative, err := filepath.Rel(root, resolved)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("%w: %q", ErrEscapingPath, name)
	}
	return nil
}
