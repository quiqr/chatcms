package gitsync

import (
	"context"
	"os"
	"path/filepath"
)

func (r *repo) Status(ctx context.Context) (Status, error) {
	status := Status{ExchangeInProgress: r.exchangeInProgress()}

	dirty, err := r.git.run(ctx, "status", "--porcelain", "--untracked-files=all")
	if err != nil {
		return Status{}, err
	}
	status.UnsavedWork = len(splitLines(dirty)) > 0

	if r.unborn(ctx) {
		return status, nil
	}

	remote := r.remoteRef()
	if !r.git.succeeds(ctx, "rev-parse", "--verify", "--quiet", remote) {
		// Nothing has been fetched yet, so every version is unpublished.
		count, err := r.countIn(ctx, "HEAD")
		if err != nil {
			return Status{}, err
		}
		status.Unpublished = count
		return status, nil
	}

	ahead, err := r.countIn(ctx, remote+"..HEAD")
	if err != nil {
		return Status{}, err
	}
	behind, err := r.countIn(ctx, "HEAD.."+remote)
	if err != nil {
		return Status{}, err
	}

	status.Unpublished = ahead
	status.RemoteAhead = behind
	return status, nil
}

func (r *repo) remoteRef() string {
	return "refs/remotes/origin/" + r.branch
}

// exchangeInProgress asks the repository itself rather than any state of our
// own, so there is no second source of truth to get out of step.
func (r *repo) exchangeInProgress() bool {
	for _, marker := range []string{"rebase-merge", "rebase-apply"} {
		if _, err := os.Stat(filepath.Join(r.dir, ".git", marker)); err == nil {
			return true
		}
	}
	return false
}
