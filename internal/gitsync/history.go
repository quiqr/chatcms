package gitsync

import (
	"context"
	"strconv"
	"strings"
	"time"
)

const (
	// The separators are written as git's own escapes in the format argument,
	// because an argument passed to a process cannot contain a NUL byte. git
	// expands them in its output, which is what the parser splits on.
	recordSeparator = "\x00"
	fieldSeparator  = "\x01"
	logFormat       = "%x00%H%x01%aI%x01%an%x01%ae%x01%s"
)

func (r *repo) History(ctx context.Context, limit int) ([]Version, error) {
	if limit <= 0 {
		limit = 50
	}

	out, err := r.git.run(ctx,
		"log", "--max-count="+strconv.Itoa(limit), "--name-only", "--format="+logFormat)
	if err != nil {
		// A repository with no versions has nothing to log, which is not a
		// failure a caller should have to handle.
		if r.unborn(ctx) {
			return nil, nil
		}
		return nil, err
	}
	return parseVersions(out), nil
}

func (r *repo) unborn(ctx context.Context) bool {
	return !r.git.succeeds(ctx, "rev-parse", "--verify", "--quiet", "HEAD")
}

func parseVersions(out string) []Version {
	var versions []Version

	for _, record := range strings.Split(out, recordSeparator) {
		if strings.TrimSpace(record) == "" {
			continue
		}

		header, rest, _ := strings.Cut(record, "\n")
		fields := strings.Split(header, fieldSeparator)
		if len(fields) < 5 {
			continue
		}

		when, _ := time.Parse(time.RFC3339, fields[1])
		version := Version{
			ID:      fields[0],
			Time:    when,
			Author:  Identity{Name: fields[2], Email: fields[3]},
			Message: fields[4],
		}

		for _, line := range strings.Split(rest, "\n") {
			path := strings.TrimSpace(line)
			if path != "" {
				version.Paths = append(version.Paths, path)
			}
		}

		versions = append(versions, version)
	}
	return versions
}

// versionsIn lists the versions in a range, such as the ones that arrived
// from the remote.
func (r *repo) versionsIn(ctx context.Context, revisionRange string) ([]Version, error) {
	out, err := r.git.run(ctx, "log", "--name-only", "--format="+logFormat, revisionRange)
	if err != nil {
		return nil, err
	}
	return parseVersions(out), nil
}

func (r *repo) countIn(ctx context.Context, revisionRange string) (int, error) {
	out, err := r.git.run(ctx, "rev-list", "--count", revisionRange)
	if err != nil {
		return 0, err
	}
	count, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		return 0, nil
	}
	return count, nil
}
