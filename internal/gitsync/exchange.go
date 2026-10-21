package gitsync

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (r *repo) Sync(ctx context.Context) (SyncResult, error) {
	lock := siteLock(r.dir)
	lock.Lock()
	defer lock.Unlock()

	if r.exchangeInProgress() {
		return SyncResult{}, ErrExchangeInProgress
	}

	if _, err := r.git.run(ctx, "fetch", "--prune", "origin"); err != nil {
		return SyncResult{}, unreachable(err)
	}

	remote := r.remoteRef()
	if !r.git.succeeds(ctx, "rev-parse", "--verify", "--quiet", remote) {
		return SyncResult{}, nil
	}

	if r.unborn(ctx) {
		if _, err := r.git.run(ctx, "reset", "--hard", remote); err != nil {
			return SyncResult{}, err
		}
		arrived, err := r.versionsIn(ctx, remote)
		return SyncResult{Arrived: arrived}, err
	}

	arrived, err := r.versionsIn(ctx, "HEAD.."+remote)
	if err != nil {
		return SyncResult{}, err
	}

	local, err := r.countIn(ctx, remote+"..HEAD")
	if err != nil {
		return SyncResult{}, err
	}

	if len(arrived) == 0 {
		return SyncResult{Replayed: 0}, nil
	}

	if local == 0 {
		if _, err := r.git.run(ctx, "merge", "--ff-only", remote); err != nil {
			return SyncResult{}, err
		}
		return SyncResult{Arrived: arrived}, nil
	}

	r.clearResolutions()

	if _, err := r.git.run(ctx, "rebase", remote); err != nil {
		result, conflictErr := r.collectConflict(ctx)
		if conflictErr != nil {
			return SyncResult{}, conflictErr
		}
		if result == nil {
			return SyncResult{}, err
		}
		return SyncResult{Arrived: arrived, Conflict: result}, nil
	}

	return SyncResult{Arrived: arrived, Replayed: local}, nil
}

func (r *repo) Publish(ctx context.Context) (PublishResult, error) {
	lock := siteLock(r.dir)
	lock.Lock()
	defer lock.Unlock()

	if r.exchangeInProgress() {
		return PublishResult{}, ErrExchangeInProgress
	}

	remote := r.remoteRef()
	var pending []Version
	if r.git.succeeds(ctx, "rev-parse", "--verify", "--quiet", remote) {
		versions, err := r.versionsIn(ctx, remote+"..HEAD")
		if err != nil {
			return PublishResult{}, err
		}
		pending = versions
	}

	if _, err := r.git.run(ctx, "push", "origin", "HEAD:refs/heads/"+r.branch); err != nil {
		var opErr *OperationError
		if errors.As(err, &opErr) && isRejection(opErr.Detail) {
			return PublishResult{}, fmt.Errorf("%w: %s", ErrRemoteMoved, opErr.Detail)
		}
		return PublishResult{}, unreachable(err)
	}

	if _, err := r.git.run(ctx, "fetch", "--prune", "origin"); err != nil {
		return PublishResult{Published: pending}, nil
	}
	return PublishResult{Published: pending}, nil
}

func isRejection(detail string) bool {
	lowered := strings.ToLower(detail)
	for _, sign := range []string{"non-fast-forward", "fetch first", "rejected", "stale info"} {
		if strings.Contains(lowered, sign) {
			return true
		}
	}
	return false
}

func (r *repo) Resolve(ctx context.Context, resolutions map[string][]byte) (SyncResult, error) {
	lock := siteLock(r.dir)
	lock.Lock()
	defer lock.Unlock()

	if !r.exchangeInProgress() {
		return SyncResult{}, ErrNoExchange
	}

	for name := range resolutions {
		if !r.isConflicted(name) {
			return SyncResult{}, fmt.Errorf("%w: %q", ErrNotConflicted, name)
		}
	}

	for name, contents := range resolutions {
		if err := r.applyResolution(ctx, name, contents); err != nil {
			return SyncResult{}, err
		}
		r.rememberResolution(name, contents)
	}

	return r.continueExchange(ctx)
}

func (r *repo) applyResolution(ctx context.Context, name string, contents []byte) error {
	if err := r.checkPath(name); err != nil {
		return err
	}

	full := filepath.Join(r.dir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
		return fmt.Errorf("preparing the folder for %s failed: %w", name, err)
	}
	if err := os.WriteFile(full, contents, 0o600); err != nil {
		return fmt.Errorf("writing %s failed: %w", name, err)
	}

	_, err := r.git.run(ctx, "add", "--", name)
	return err
}

// continueExchange moves the replay forward until it finishes or hits a
// difference it has not been given an answer for.
func (r *repo) continueExchange(ctx context.Context) (SyncResult, error) {
	for {
		args := []string{"rebase", "--continue"}
		if r.git.succeeds(ctx, "diff", "--cached", "--quiet", "--exit-code") &&
			r.git.succeeds(ctx, "diff", "--quiet", "--exit-code") {
			args = []string{"rebase", "--skip"}
		}

		if _, err := r.git.run(ctx, args...); err != nil {
			if !r.exchangeInProgress() {
				return SyncResult{}, err
			}

			conflict, collectErr := r.collectConflict(ctx)
			if collectErr != nil {
				return SyncResult{}, collectErr
			}
			if conflict == nil {
				return SyncResult{}, err
			}

			applied, applyErr := r.applyRemembered(ctx, conflict)
			if applyErr != nil {
				return SyncResult{}, applyErr
			}
			if applied {
				continue
			}
			return SyncResult{Conflict: conflict}, nil
		}

		if r.exchangeInProgress() {
			continue
		}

		r.clearConflicts()
		return SyncResult{}, nil
	}
}

func (r *repo) Abandon(ctx context.Context) error {
	lock := siteLock(r.dir)
	lock.Lock()
	defer lock.Unlock()

	if !r.exchangeInProgress() {
		return ErrNoExchange
	}

	if _, err := r.git.run(ctx, "rebase", "--abort"); err != nil {
		return err
	}
	r.clearConflicts()
	r.clearResolutions()
	return nil
}

// collectConflict reads every version of each unresolved path out of the
// index and then puts valid content back on disk, so no file an editor can
// open ever holds a marker. See design.md.
func (r *repo) collectConflict(ctx context.Context) (*Conflict, error) {
	entries, err := r.git.nulFields(ctx, "ls-files", "--unmerged", "-z")
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, nil
	}

	type stages struct {
		base, ours, theirs string
	}
	byPath := map[string]*stages{}
	var order []string

	for _, entry := range entries {
		// "<mode> <object> <stage>\t<path>"
		meta, path, found := strings.Cut(entry, "\t")
		if !found {
			continue
		}
		fields := strings.Fields(meta)
		if len(fields) != 3 {
			continue
		}

		if _, seen := byPath[path]; !seen {
			byPath[path] = &stages{}
			order = append(order, path)
		}
		switch fields[2] {
		case "1":
			byPath[path].base = fields[1]
		case "2":
			byPath[path].ours = fields[1]
		case "3":
			byPath[path].theirs = fields[1]
		}
	}

	conflict := &Conflict{}
	for _, path := range order {
		stage := byPath[path]

		base, err := r.blob(ctx, stage.base)
		if err != nil {
			return nil, err
		}
		// During a replay git calls the side already on the remote "ours" and
		// the version being replayed "theirs", which is the other way round
		// from how a person sees it.
		remote, err := r.blob(ctx, stage.ours)
		if err != nil {
			return nil, err
		}
		local, err := r.blob(ctx, stage.theirs)
		if err != nil {
			return nil, err
		}

		conflict.Paths = append(conflict.Paths, ConflictedPath{
			Path: path, Base: base, Local: local, Remote: remote,
		})

		if err := r.writeCleanContent(path, remote, local); err != nil {
			return nil, err
		}
	}

	r.setConflicts(conflict.pathNames())
	return conflict, nil
}

// writeCleanContent replaces the marked-up file git left on disk with content
// that parses. The side already on the remote is used when there is one,
// because it is the version the rest of the working copy is consistent with.
func (r *repo) writeCleanContent(path string, remote, local []byte) error {
	contents := remote
	if contents == nil {
		contents = local
	}

	full := filepath.Join(r.dir, filepath.FromSlash(path))
	if contents == nil {
		return os.RemoveAll(full)
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
		return err
	}
	return os.WriteFile(full, contents, 0o600)
}

func (r *repo) blob(ctx context.Context, object string) ([]byte, error) {
	if object == "" {
		return nil, nil
	}
	out, err := r.git.run(ctx, "cat-file", "blob", object)
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}

// recoverConflicts rebuilds what an unfinished exchange was stuck on, so that
// opening a working copy after an interruption reports the same thing the
// process that was interrupted would have.
func (r *repo) recoverConflicts(ctx context.Context) error {
	if !r.exchangeInProgress() {
		return nil
	}
	_, err := r.collectConflict(ctx)
	return err
}

func (r *repo) setConflicts(paths []string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.conflicted = make(map[string]bool, len(paths))
	for _, path := range paths {
		r.conflicted[path] = true
	}
}

func (r *repo) clearConflicts() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.conflicted = map[string]bool{}
}

func (r *repo) isConflicted(path string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.conflicted[path]
}

// rememberResolution keeps the answer given for one page for the rest of this
// exchange. A replay stops once per version it carries, so without this an
// editor would be asked about the same page two or three times in a row for
// what is, to them, one question. The memory lasts for one exchange only: the
// next sync asks again, because by then the page may have changed.
func (r *repo) rememberResolution(name string, contents []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resolved[name] = contents
}

func (r *repo) clearResolutions() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resolved = map[string][]byte{}
}

// applyRemembered answers any part of a conflict that was already answered
// during this exchange, and reports whether everything was covered.
func (r *repo) applyRemembered(ctx context.Context, conflict *Conflict) (bool, error) {
	r.mu.Lock()
	answers := make(map[string][]byte, len(conflict.Paths))
	for _, path := range conflict.Paths {
		if contents, ok := r.resolved[path.Path]; ok {
			answers[path.Path] = contents
		}
	}
	r.mu.Unlock()

	if len(answers) != len(conflict.Paths) {
		return false, nil
	}

	for name, contents := range answers {
		if err := r.applyResolution(ctx, name, contents); err != nil {
			return false, err
		}
	}
	return true, nil
}
