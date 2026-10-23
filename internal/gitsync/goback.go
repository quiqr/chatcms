package gitsync

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrNoSuchVersion says the site has no version by that name.
	ErrNoSuchVersion = errors.New("gitsync: the site has no such version")
	// ErrAlreadyThere says the site already holds exactly that content, so
	// going back to it would change nothing.
	ErrAlreadyThere = errors.New("gitsync: the site is already as that version left it")
)

// GoBackTo makes a new version whose content is the content of an earlier
// one. It does not throw versions away: everything that happened since is
// still in the history, so going back is itself something that can be undone.
func (r *repo) GoBackTo(ctx context.Context, version string, author Identity, message string) (Version, error) {
	lock := siteLock(r.dir)
	lock.Lock()
	defer lock.Unlock()

	if r.exchangeInProgress() {
		return Version{}, ErrExchangeInProgress
	}

	tree, err := r.treeOf(ctx, version)
	if err != nil {
		return Version{}, err
	}
	if current, err := r.tree(ctx, "HEAD"); err == nil && current == tree {
		return Version{}, ErrAlreadyThere
	}

	if _, err := r.git.run(ctx, "read-tree", "-u", "--reset", tree); err != nil {
		return Version{}, err
	}

	when := []string{
		"GIT_AUTHOR_NAME=" + author.Name,
		"GIT_AUTHOR_EMAIL=" + author.Email,
	}
	if _, err := r.git.runWith(ctx, nil, when,
		"commit", "--allow-empty", "--message="+message); err != nil {
		r.git.succeeds(ctx, "read-tree", "-u", "--reset", "HEAD")
		return Version{}, err
	}

	made, err := r.versionsIn(ctx, "-1")
	if err != nil || len(made) == 0 {
		return Version{}, err
	}
	return made[0], nil
}

// treeOf is what a version identifier from outside resolves to. tree is the
// same question asked of something this package wrote itself, such as HEAD,
// which checkVersion would otherwise refuse.
func (r *repo) treeOf(ctx context.Context, version string) (string, error) {
	if err := checkVersion(version); err != nil {
		return "", err
	}
	return r.tree(ctx, version)
}

func (r *repo) tree(ctx context.Context, revision string) (string, error) {
	out, err := r.git.run(ctx, "rev-parse", "--verify", "--quiet", revision+"^{tree}")
	if err != nil || strings.TrimSpace(out) == "" {
		return "", fmt.Errorf("%w: %s", ErrNoSuchVersion, revision)
	}
	return strings.TrimSpace(out), nil
}

// checkVersion refuses anything that is not plainly a version identifier, so
// that a name from outside cannot become an argument to git.
func checkVersion(version string) error {
	if version == "" {
		return fmt.Errorf("%w: it has no name", ErrNoSuchVersion)
	}
	for _, letter := range version {
		switch {
		case letter >= '0' && letter <= '9',
			letter >= 'a' && letter <= 'f',
			letter >= 'A' && letter <= 'F':
		default:
			return fmt.Errorf("%w: %q is not a version identifier", ErrNoSuchVersion, version)
		}
	}
	return nil
}

// ChangesIn says what one version changed, compared with the one before it.
func (r *repo) ChangesIn(ctx context.Context, version string) ([]Change, error) {
	if err := checkVersion(version); err != nil {
		return nil, err
	}
	if _, err := r.treeOf(ctx, version); err != nil {
		return nil, err
	}

	against := version + "^"
	if !r.git.succeeds(ctx, "rev-parse", "--verify", "--quiet", against+"^{commit}") {
		against = emptyTree
	}

	fields, err := r.git.nulFields(ctx, "diff", "--name-status", "--no-renames", "-z", against, version)
	if err != nil {
		return nil, err
	}

	who, when := r.madeBy(ctx, version)
	changes := []Change{}
	for index := 0; index+1 < len(fields); index += 2 {
		kind, known := kindOf(fields[index])
		if !known {
			continue
		}
		changes = append(changes, Change{Path: fields[index+1], Kind: kind, By: who, At: when})
	}
	sort.Slice(changes, func(a, b int) bool { return changes[a].Path < changes[b].Path })
	return changes, nil
}

func (r *repo) madeBy(ctx context.Context, version string) (Identity, time.Time) {
	out, err := r.git.run(ctx, "log", "-1", "--format=%an%x00%ae%x00%aI", version)
	if err != nil {
		return Identity{}, time.Time{}
	}
	parts := splitNul(strings.TrimRight(out, "\n"))
	if len(parts) < 3 {
		return Identity{}, time.Time{}
	}
	at, _ := time.Parse(time.RFC3339, parts[2])
	return Identity{Name: parts[0], Email: parts[1]}, at
}

// At is the site as one version left it, read only. It exists so that a
// person can be shown what a version holds, and so that the site as it was
// can be previewed, without the working copy being disturbed.
func (r *repo) At(ctx context.Context, version string) (fs.FS, error) {
	if _, err := r.treeOf(ctx, version); err != nil {
		return nil, err
	}
	return r.at(ctx, version)
}

// Before is the site as it was just before one version, which is what a
// comparison of that version needs on its other side. The first version of a
// site has an empty site before it rather than nothing at all.
func (r *repo) Before(ctx context.Context, version string) (fs.FS, error) {
	if _, err := r.treeOf(ctx, version); err != nil {
		return nil, err
	}

	parent := version + "^"
	if _, err := r.tree(ctx, parent); err != nil {
		parent = emptyTree
	}
	return r.at(ctx, parent)
}

func (r *repo) at(ctx context.Context, revision string) (fs.FS, error) {
	listed, err := r.git.nulFields(ctx, "ls-tree", "-r", "-z", "--format=%(objectmode) %(objectname) %(objectsize) %(path)", revision)
	if err != nil {
		return nil, err
	}

	held := &atVersion{files: map[string]atFile{}, dirs: map[string]bool{".": true}}
	for _, entry := range listed {
		mode, rest, found := strings.Cut(entry, " ")
		if !found {
			continue
		}
		object, rest, found := strings.Cut(rest, " ")
		if !found {
			continue
		}
		size, name, found := strings.Cut(rest, " ")
		if !found || !fs.ValidPath(name) {
			continue
		}
		bytes, err := strconv.ParseInt(strings.TrimSpace(size), 10, 64)
		if err != nil {
			continue
		}
		held.files[name] = atFile{object: object, mode: mode, size: bytes}
		for dir := path.Dir(name); dir != "." && dir != "/"; dir = path.Dir(dir) {
			held.dirs[dir] = true
		}
	}

	held.read = func(object string) ([]byte, error) {
		return r.git.runWith(ctx, nil, nil, "cat-file", "blob", object)
	}
	return held, nil
}

type atFile struct {
	object string
	mode   string
	size   int64
}

// atVersion is a read-only filesystem over what one version holds. It reads a
// file's content only when it is asked for, so showing what a version
// contains does not read the whole site.
type atVersion struct {
	files map[string]atFile
	dirs  map[string]bool
	read  func(object string) ([]byte, error)
}

func (a *atVersion) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	if a.dirs[name] {
		return &atDir{fsys: a, name: name}, nil
	}

	held, found := a.files[name]
	if !found {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	content, err := a.read(held.object)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}
	return &atOpenFile{
		name: path.Base(name),
		mode: held.mode,
		body: strings.NewReader(string(content)),
		size: int64(len(content)),
	}, nil
}

func (a *atVersion) ReadDir(name string) ([]fs.DirEntry, error) {
	if !fs.ValidPath(name) || !a.dirs[name] {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
	}

	seen := map[string]bool{}
	entries := []fs.DirEntry{}
	for file := range a.files {
		if path.Dir(file) == name {
			entries = append(entries, atEntry{name: path.Base(file), file: a.files[file]})
			continue
		}
		if name == "." || strings.HasPrefix(file, name+"/") {
			rest := strings.TrimPrefix(strings.TrimPrefix(file, name), "/")
			if first, _, nested := strings.Cut(rest, "/"); nested && !seen[first] {
				seen[first] = true
				entries = append(entries, atEntry{name: first, dir: true})
			}
		}
	}
	sort.Slice(entries, func(x, y int) bool { return entries[x].Name() < entries[y].Name() })
	return entries, nil
}

type atEntry struct {
	name string
	dir  bool
	file atFile
}

func (e atEntry) Name() string { return e.name }
func (e atEntry) IsDir() bool  { return e.dir }
func (e atEntry) Type() fs.FileMode {
	if e.dir {
		return fs.ModeDir
	}
	return 0
}
func (e atEntry) Info() (fs.FileInfo, error) {
	return atInfo{
		name:       e.name,
		dir:        e.dir,
		size:       e.file.size,
		executable: e.file.mode == "100755",
	}, nil
}

type atDir struct {
	fsys *atVersion
	name string
	read int
}

func (d *atDir) Stat() (fs.FileInfo, error) { return atInfo{name: path.Base(d.name), dir: true}, nil }
func (d *atDir) Read([]byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: d.name, Err: errors.New("is a directory")}
}
func (d *atDir) Close() error { return nil }

func (d *atDir) ReadDir(count int) ([]fs.DirEntry, error) {
	entries, err := d.fsys.ReadDir(d.name)
	if err != nil {
		return nil, err
	}
	if count <= 0 {
		remaining := entries[min(d.read, len(entries)):]
		d.read = len(entries)
		return remaining, nil
	}
	if d.read >= len(entries) {
		return nil, io.EOF
	}
	end := min(d.read+count, len(entries))
	taken := entries[d.read:end]
	d.read = end
	return taken, nil
}

type atOpenFile struct {
	name string
	mode string
	body *strings.Reader
	size int64
}

func (f *atOpenFile) Stat() (fs.FileInfo, error) {
	return atInfo{name: f.name, size: f.size, executable: f.mode == "100755"}, nil
}
func (f *atOpenFile) Read(into []byte) (int, error) { return f.body.Read(into) }
func (f *atOpenFile) Close() error                  { return nil }

type atInfo struct {
	name       string
	size       int64
	dir        bool
	executable bool
}

func (i atInfo) Name() string       { return i.name }
func (i atInfo) Size() int64        { return i.size }
func (i atInfo) ModTime() time.Time { return time.Time{} }
func (i atInfo) IsDir() bool        { return i.dir }
func (i atInfo) Sys() any           { return nil }
func (i atInfo) Mode() fs.FileMode {
	if i.dir {
		return fs.ModeDir | 0o555
	}
	if i.executable {
		return 0o555
	}
	return 0o444
}
