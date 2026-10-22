package runner_test

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/quiqr/chatcms/internal/runner"
)

// bare runs a task with nothing between it and the machine. It exists to be
// held against the same table as the sandbox: if the interface ever grows a
// member only bubblewrap could serve, this stops compiling, and if a
// behaviour only the sandbox can have ever moves into the table, this fails.
// It is a test fixture and is never built into the server.
type bare struct {
	root string
}

func (b *bare) Run(ctx context.Context, task runner.Task) (*runner.Result, error) {
	if err := runner.Validate(task); err != nil {
		return nil, err
	}

	workspace, err := os.MkdirTemp(b.root, "bare-")
	if err != nil {
		return nil, err
	}
	work := filepath.Join(workspace, "work")
	if err := os.MkdirAll(work, 0o700); err != nil {
		return nil, err
	}
	if task.Inputs != nil {
		if err := copyInto(task.Inputs, work); err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithTimeout(ctx, task.Timeout)
	defer cancel()

	log := &writeCollector{}
	cmd := exec.CommandContext(ctx, task.Command, task.Arguments...)
	cmd.Dir = work
	cmd.Stdout = log
	cmd.Stderr = log
	cmd.Env = []string{"PATH=" + os.Getenv("PATH")}
	for name, value := range task.Environment {
		cmd.Env = append(cmd.Env, name+"="+value)
	}

	started := time.Now()
	runErr := cmd.Run()
	result := &runner.Result{
		Log:      log.String(),
		Duration: time.Since(started),
	}

	var exitErr *exec.ExitError
	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		result.Reached = runner.LimitTime
		result.ExitCode = -1
	case errors.As(runErr, &exitErr):
		result.ExitCode = exitErr.ExitCode()
	case runErr != nil:
		return nil, runErr
	}

	if result.Succeeded() {
		collected := filepath.Join(workspace, "outputs")
		for _, want := range task.Outputs {
			source := filepath.Join(work, filepath.FromSlash(want))
			if _, err := os.Lstat(source); err != nil {
				return nil, &runner.MissingOutputError{Path: want}
			}
			target := filepath.Join(collected, filepath.FromSlash(want))
			if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
				return nil, err
			}
			if err := os.Rename(source, target); err != nil {
				return nil, err
			}
		}
		result.Outputs = os.DirFS(collected)
	}
	return result, nil
}

type writeCollector struct {
	seen []byte
}

func (w *writeCollector) Write(p []byte) (int, error) {
	w.seen = append(w.seen, p...)
	return len(p), nil
}

func (w *writeCollector) String() string { return string(w.seen) }

func copyInto(from fs.FS, to string) error {
	return fs.WalkDir(from, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(to, filepath.FromSlash(name))
		if entry.IsDir() {
			return os.MkdirAll(target, 0o700)
		}
		source, err := from.Open(name)
		if err != nil {
			return err
		}
		defer source.Close()
		sink, err := os.Create(target)
		if err != nil {
			return err
		}
		defer sink.Close()
		_, err = io.Copy(sink, source)
		return err
	})
}

var _ runner.Runner = (*bare)(nil)

func TestAnUnisolatedRunnerSatisfiesTheSameInterface(t *testing.T) {
	runBehaviours(t, func(t *testing.T) runner.Runner {
		return &bare{root: t.TempDir()}
	})
}

func TestTheTableSaysNothingAboutIsolation(t *testing.T) {
	// The point of the table is what every runner owes a caller. Isolation is
	// not in it, because a caller cannot observe it through this interface;
	// it is proven separately against the sandbox.
	for _, test := range behaviours() {
		if test.task().Network {
			t.Errorf("%q asks for the network; that belongs with the sandbox tests", test.name)
		}
	}
}
