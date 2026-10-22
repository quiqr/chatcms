package gitsync

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ChangeKind says what happened to one file since the last publish.
type ChangeKind string

const (
	Added   ChangeKind = "added"
	Edited  ChangeKind = "edited"
	Removed ChangeKind = "removed"
)

// Change is one file that differs between the site as people see it and the
// site as it is now.
type Change struct {
	Path string     `json:"path"`
	Kind ChangeKind `json:"kind"`
	By   Identity   `json:"by"`
	At   time.Time  `json:"at"`
}

var ErrNotPublished = errors.New("gitsync: the published site does not have that file")

func (r *repo) Changes(ctx context.Context) ([]Change, error) {
	lock := siteLock(r.dir)
	lock.Lock()
	defer lock.Unlock()

	against := r.remoteRef()
	if !r.git.succeeds(ctx, "rev-parse", "--verify", "--quiet", against) {
		against = emptyTree
	}

	fields, err := r.git.nulFields(ctx, "diff", "--name-status", "--no-renames", "-z", against, "HEAD")
	if err != nil {
		return nil, err
	}

	changes := []Change{}
	for index := 0; index+1 < len(fields); index += 2 {
		kind, known := kindOf(fields[index])
		if !known {
			continue
		}
		change := Change{Path: fields[index+1], Kind: kind}
		if err := r.attribute(ctx, &change); err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}
	sort.Slice(changes, func(a, b int) bool { return changes[a].Path < changes[b].Path })
	return changes, nil
}

const emptyTree = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

func kindOf(status string) (ChangeKind, bool) {
	switch {
	case strings.HasPrefix(status, "A"):
		return Added, true
	case strings.HasPrefix(status, "D"):
		return Removed, true
	case strings.HasPrefix(status, "M"), strings.HasPrefix(status, "T"):
		return Edited, true
	default:
		return "", false
	}
}

// attribute asks who last touched a path and when. A removed file is not in
// the tree any more, so the question is asked of the version that removed it.
func (r *repo) attribute(ctx context.Context, change *Change) error {
	out, err := r.git.run(ctx, "log", "-1", "--format=%an%x00%ae%x00%aI", "--", change.Path)
	if err != nil {
		return err
	}
	parts := splitNul(strings.TrimRight(out, "\n"))
	if len(parts) < 3 {
		return nil
	}
	change.By = Identity{Name: parts[0], Email: parts[1]}
	if when, err := time.Parse(time.RFC3339, parts[2]); err == nil {
		change.At = when
	}
	return nil
}

func (r *repo) ReadPublished(ctx context.Context, name string) ([]byte, error) {
	if !r.git.succeeds(ctx, "rev-parse", "--verify", "--quiet", r.remoteRef()) {
		return nil, ErrNotPublished
	}
	out, err := r.git.runWith(ctx, nil, nil, "cat-file", "blob", r.remoteRef()+":"+name)
	if err != nil {
		return nil, ErrNotPublished
	}
	return out, nil
}

// PublishTo sends the site as it is to a named line on the remote. An empty
// name means the line the site is edited on.
func (r *repo) PublishTo(ctx context.Context, line string) (PublishResult, error) {
	if line == "" || line == r.branch {
		return r.Publish(ctx)
	}
	if err := checkLine(line); err != nil {
		return PublishResult{}, err
	}

	lock := siteLock(r.dir)
	lock.Lock()
	defer lock.Unlock()

	if r.exchangeInProgress() {
		return PublishResult{}, ErrExchangeInProgress
	}

	pending, err := r.pendingFor(ctx, "refs/remotes/origin/"+line)
	if err != nil {
		return PublishResult{}, err
	}
	if err := r.push(ctx, "HEAD", line); err != nil {
		return PublishResult{}, err
	}
	return PublishResult{Published: pending}, nil
}

// PublishPages sends built pages to a line of their own. It builds the tree
// without touching the working copy, so somebody can keep editing while it
// happens, and it makes no version of the site's source.
func (r *repo) PublishPages(ctx context.Context, line string, pages fs.FS, author Identity, message string) (PublishResult, error) {
	if err := checkLine(line); err != nil {
		return PublishResult{}, err
	}
	if line == r.branch {
		return PublishResult{}, fmt.Errorf("%w: built pages would replace the site's own files", ErrUnsuitableLine)
	}
	if pages == nil {
		return PublishResult{}, errors.New("gitsync: there are no pages to publish")
	}

	lock := siteLock(r.dir)
	lock.Lock()
	defer lock.Unlock()

	index, err := r.stage(ctx, pages)
	if err != nil {
		return PublishResult{}, err
	}
	defer index.remove()

	tree, err := r.git.runWith(ctx, nil, index.settings(), "write-tree")
	if err != nil {
		return PublishResult{}, err
	}

	arguments := []string{"commit-tree", strings.TrimSpace(string(tree)), "-m", message}
	if parent, found := r.tipOf(ctx, line); found {
		arguments = append(arguments, "-p", parent)
	}
	when := []string{
		"GIT_AUTHOR_NAME=" + author.Name,
		"GIT_AUTHOR_EMAIL=" + author.Email,
	}
	made, err := r.git.runWith(ctx, nil, when, arguments...)
	if err != nil {
		return PublishResult{}, err
	}

	version := strings.TrimSpace(string(made))
	if err := r.push(ctx, version, line); err != nil {
		return PublishResult{}, err
	}
	return PublishResult{Published: []Version{{
		ID:      version,
		Time:    time.Now().UTC(),
		Author:  author,
		Message: message,
	}}}, nil
}

var ErrUnsuitableLine = errors.New("gitsync: that is not somewhere this can be published")

func checkLine(line string) error {
	switch {
	case line == "":
		return fmt.Errorf("%w: it has no name", ErrUnsuitableLine)
	case strings.HasPrefix(line, "-"), strings.Contains(line, ".."), strings.Contains(line, " "):
		return fmt.Errorf("%w: %q is not a usable name", ErrUnsuitableLine, line)
	}
	return nil
}

func (r *repo) pendingFor(ctx context.Context, remote string) ([]Version, error) {
	if !r.git.succeeds(ctx, "rev-parse", "--verify", "--quiet", remote) {
		return r.versionsIn(ctx, "HEAD")
	}
	return r.versionsIn(ctx, remote+"..HEAD")
}

func (r *repo) push(ctx context.Context, what, line string) error {
	if _, err := r.git.run(ctx, "push", "origin", what+":refs/heads/"+line); err != nil {
		var opErr *OperationError
		if errors.As(err, &opErr) && isRejection(opErr.Detail) {
			return fmt.Errorf("%w: %s", ErrRemoteMoved, opErr.Detail)
		}
		return unreachable(err)
	}
	r.git.succeeds(ctx, "fetch", "--prune", "origin")
	return nil
}

func (r *repo) tipOf(ctx context.Context, line string) (string, bool) {
	out, err := r.git.run(ctx, "rev-parse", "--verify", "--quiet", "refs/remotes/origin/"+line)
	if err != nil {
		return "", false
	}
	tip := strings.TrimSpace(out)
	return tip, tip != ""
}

type stagedIndex struct {
	path   string
	remove func()
}

func (s stagedIndex) settings() []string { return []string{"GIT_INDEX_FILE=" + s.path} }

// stage writes every page into the object store and builds an index out of
// them, with no working copy involved at all.
func (r *repo) stage(ctx context.Context, pages fs.FS) (stagedIndex, error) {
	dir, err := os.MkdirTemp("", "chatcms-pages-")
	if err != nil {
		return stagedIndex{}, err
	}
	index := stagedIndex{path: filepath.Join(dir, "index"), remove: func() { os.RemoveAll(dir) }}

	var listing strings.Builder
	err = fs.WalkDir(pages, ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() || !entry.Type().IsRegular() {
			return walkErr
		}
		blob, err := r.store(ctx, pages, name)
		if err != nil {
			return err
		}
		mode := "100644"
		if info, err := entry.Info(); err == nil && info.Mode().Perm()&0o100 != 0 {
			mode = "100755"
		}
		fmt.Fprintf(&listing, "%s %s\t%s\x00", mode, blob, strings.TrimPrefix(name, "./"))
		return nil
	})
	if err != nil {
		index.remove()
		return stagedIndex{}, err
	}
	if listing.Len() == 0 {
		index.remove()
		return stagedIndex{}, errors.New("gitsync: there are no pages to publish")
	}

	if _, err := r.git.runWith(ctx, []byte(listing.String()), index.settings(),
		"update-index", "-z", "--index-info"); err != nil {
		index.remove()
		return stagedIndex{}, err
	}
	return index, nil
}

func (r *repo) store(ctx context.Context, pages fs.FS, name string) (string, error) {
	file, err := pages.Open(name)
	if err != nil {
		return "", err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	out, err := r.git.runWith(ctx, content, nil, "hash-object", "-w", "--stdin")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
