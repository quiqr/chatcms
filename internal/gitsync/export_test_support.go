package gitsync

import "context"

// RunForTest lets another package's tests build a repository on disk without
// learning how to run git. It exists so that the rule in
// docs/ARCHITECTURE.md holds for test code as well: every git invocation
// originates here.
func RunForTest(ctx context.Context, cfg Config, dir, home string, args ...string) error {
	if dir == "" {
		dir = home
	}
	runner := newRunner(cfg, dir, home)
	_, err := runner.run(ctx, args...)
	return err
}
