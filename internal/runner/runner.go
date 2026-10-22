package runner

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"time"
)

type Task struct {
	Inputs      fs.FS
	Command     string
	Arguments   []string
	Environment map[string]string
	Tools       []string
	Outputs     []string
	Timeout     time.Duration
	MemoryBytes int64
	Network     bool
}

type Limit string

const (
	LimitNone   Limit = ""
	LimitTime   Limit = "time"
	LimitMemory Limit = "memory"
)

type Result struct {
	ExitCode int
	Log      string
	Outputs  fs.FS
	Duration time.Duration
	Reached  Limit

	OnRelease func() error
}

func (r *Result) Succeeded() bool {
	return r.ExitCode == 0 && r.Reached == LimitNone
}

func (r *Result) Release() error {
	if r == nil || r.OnRelease == nil {
		return nil
	}
	release := r.OnRelease
	r.OnRelease = nil
	r.Outputs = nil
	return release()
}

type Runner interface {
	Run(ctx context.Context, task Task) (*Result, error)
}

var (
	ErrNoIsolation   = errors.New("runner: builds cannot be isolated on this machine")
	ErrNetworkDenied = errors.New("runner: this runner will not grant network access")
)

type MissingOutputError struct {
	Path string
}

func (e *MissingOutputError) Error() string {
	return fmt.Sprintf("runner: the build did not produce %q", e.Path)
}

type TaskError struct {
	Field  string
	Reason string
}

func (e *TaskError) Error() string {
	return fmt.Sprintf("runner: %s: %s", e.Field, e.Reason)
}

func Validate(task Task) error {
	if task.Command == "" {
		return &TaskError{Field: "command", Reason: "no command to run"}
	}
	if task.Timeout <= 0 {
		return &TaskError{Field: "timeout", Reason: "a build must have a time limit"}
	}
	if task.MemoryBytes <= 0 {
		return &TaskError{Field: "memory", Reason: "a build must have a memory limit"}
	}
	for _, want := range task.Outputs {
		if err := validOutput(want); err != nil {
			return err
		}
	}
	for name := range task.Environment {
		if name == "" || strings.ContainsAny(name, "=\x00") {
			return &TaskError{Field: "environment", Reason: fmt.Sprintf("%q is not a usable name", name)}
		}
	}
	return nil
}

func validOutput(want string) error {
	reason := ""
	switch {
	case want == "":
		reason = "an output needs a path"
	case path.IsAbs(want) || strings.HasPrefix(want, "/"):
		reason = fmt.Sprintf("%q is absolute; outputs are relative to the working directory", want)
	case strings.Contains(want, "\x00"):
		reason = fmt.Sprintf("%q contains a null byte", want)
	case !fs.ValidPath(path.Clean(want)):
		reason = fmt.Sprintf("%q climbs out of the working directory", want)
	}
	if reason != "" {
		return &TaskError{Field: "outputs", Reason: reason}
	}
	return nil
}
