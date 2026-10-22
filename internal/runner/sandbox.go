package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	workInside = "/work"
	tempInside = "/tmp"
)

type Config struct {
	BwrapBin     string
	PrlimitBin   string
	WorkRoot     string
	StoreDir     string
	ReadOnlyDirs []string
	AllowNetwork bool
	MaxLogBytes  int
}

func (c Config) withDefaults() Config {
	if c.BwrapBin == "" {
		c.BwrapBin = "bwrap"
	}
	if c.PrlimitBin == "" {
		c.PrlimitBin = "prlimit"
	}
	if c.WorkRoot == "" {
		c.WorkRoot = os.TempDir()
	}
	if c.StoreDir == "" {
		c.StoreDir = "/nix/store"
	}
	if c.ReadOnlyDirs == nil {
		c.ReadOnlyDirs = []string{"/usr", "/bin", "/sbin", "/lib", "/lib64"}
	}
	if c.MaxLogBytes == 0 {
		c.MaxLogBytes = DefaultLogBytes
	}
	return c
}

type Sandbox struct {
	cfg      Config
	bwrap    string
	prlimit  string
	readOnly []string
}

func NewSandbox(cfg Config) (*Sandbox, error) {
	cfg = cfg.withDefaults()

	bwrap, err := realPath(cfg.BwrapBin)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNoIsolation, err)
	}
	prlimit, err := realPath(cfg.PrlimitBin)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNoIsolation, err)
	}
	if err := os.MkdirAll(cfg.WorkRoot, 0o700); err != nil {
		return nil, fmt.Errorf("runner: %w", err)
	}

	s := &Sandbox{cfg: cfg, bwrap: bwrap, prlimit: prlimit}
	for _, dir := range append([]string{cfg.StoreDir}, cfg.ReadOnlyDirs...) {
		if _, err := os.Stat(dir); err == nil {
			s.readOnly = append(s.readOnly, dir)
		}
	}
	if err := s.probe(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Sandbox) probe() error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	args := s.isolation(false)
	if !s.alreadyReadable(s.prlimit) {
		args = append(args, "--ro-bind", s.prlimit, s.prlimit)
	}
	args = append(args, "--", s.prlimit, "--version")

	out, err := exec.CommandContext(ctx, s.bwrap, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %v: %s", ErrNoIsolation, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (s *Sandbox) isolation(network bool) []string {
	args := []string{"--unshare-all", "--die-with-parent", "--new-session"}
	if network {
		args = append(args, "--share-net")
	}
	for _, dir := range s.readOnly {
		args = append(args, "--ro-bind", dir, dir)
	}
	return append(args, "--proc", "/proc", "--dev", "/dev", "--tmpfs", tempInside)
}

func (s *Sandbox) Run(ctx context.Context, task Task) (*Result, error) {
	if err := Validate(task); err != nil {
		return nil, err
	}
	if task.Network && !s.cfg.AllowNetwork {
		return nil, ErrNetworkDenied
	}

	workspace, err := os.MkdirTemp(s.cfg.WorkRoot, "build-")
	if err != nil {
		return nil, fmt.Errorf("runner: %w", err)
	}
	keep := false
	defer func() {
		if !keep {
			os.RemoveAll(workspace)
		}
	}()

	work := filepath.Join(workspace, "work")
	if err := os.MkdirAll(work, 0o700); err != nil {
		return nil, fmt.Errorf("runner: %w", err)
	}
	if task.Inputs != nil {
		if err := copyTree(task.Inputs, work); err != nil {
			return nil, fmt.Errorf("runner: laying out the inputs: %w", err)
		}
	}

	command, binds, toolDirs, err := s.resolve(task)
	if err != nil {
		return nil, err
	}

	args := s.isolation(task.Network)
	for _, bind := range binds {
		args = append(args, "--ro-bind", bind, bind)
	}
	args = append(args, "--bind", work, workInside, "--chdir", workInside, "--clearenv")
	for name, value := range environment(task, toolDirs) {
		args = append(args, "--setenv", name, value)
	}
	args = append(args, "--",
		s.prlimit, "--data="+strconv.FormatInt(task.MemoryBytes, 10), "--",
		command)
	args = append(args, task.Arguments...)

	ctx, cancel := context.WithTimeout(ctx, task.Timeout)
	defer cancel()

	log := &tailWriter{max: s.cfg.MaxLogBytes}
	exitCode, reached, duration, err := s.execute(ctx, args, log)
	if err != nil {
		return nil, err
	}

	text := log.String()
	if reached == LimitNone && exitCode != 0 && outOfMemory(text) {
		reached = LimitMemory
	}

	result := &Result{
		ExitCode: exitCode,
		Log:      text,
		Duration: duration,
		Reached:  reached,
	}
	if exitCode == 0 && reached == LimitNone {
		collected := filepath.Join(workspace, "outputs")
		if err := collect(work, collected, task.Outputs); err != nil {
			return nil, err
		}
		result.Outputs = os.DirFS(collected)
	}
	result.OnRelease = func() error { return os.RemoveAll(workspace) }
	keep = true
	return result, nil
}

func (s *Sandbox) execute(ctx context.Context, args []string, log io.Writer) (int, Limit, time.Duration, error) {
	cmd := exec.CommandContext(ctx, s.bwrap, args...)
	cmd.Stdout = log
	cmd.Stderr = log
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = 5 * time.Second

	started := time.Now()
	err := cmd.Run()
	duration := time.Since(started)

	switch {
	case err == nil:
		return 0, LimitNone, duration, nil
	case ctx.Err() != nil && errors.Is(ctx.Err(), context.DeadlineExceeded):
		return exitCodeOf(err), LimitTime, duration, nil
	case ctx.Err() != nil:
		return 0, LimitNone, duration, ctx.Err()
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), LimitNone, duration, nil
	}
	return 0, LimitNone, duration, fmt.Errorf("runner: %w", err)
}

func outOfMemory(log string) bool {
	lowered := strings.ToLower(log)
	for _, said := range []string{
		"cannot allocate",
		"out of memory",
		"memory exhausted",
		"bad_alloc",
	} {
		if strings.Contains(lowered, said) {
			return true
		}
	}
	return false
}

func exitCodeOf(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

func (s *Sandbox) resolve(task Task) (string, []string, []string, error) {
	command, err := realPath(task.Command)
	if err != nil {
		return "", nil, nil, &TaskError{Field: "command", Reason: err.Error()}
	}

	needed := map[string]bool{}
	dirs := map[string]bool{}
	for _, named := range append([]string{command, s.prlimit}, task.Tools...) {
		real, err := realPath(named)
		if err != nil {
			return "", nil, nil, &TaskError{Field: "tools", Reason: err.Error()}
		}
		if !s.alreadyReadable(real) {
			needed[real] = true
		}
		if info, err := os.Stat(real); err == nil && info.IsDir() {
			dirs[real] = true
		} else {
			dirs[filepath.Dir(real)] = true
		}
	}
	return command, sorted(needed), sorted(dirs), nil
}

func realPath(named string) (string, error) {
	found, err := exec.LookPath(named)
	if err != nil {
		if _, statErr := os.Stat(named); statErr != nil {
			return "", err
		}
		found = named
	}
	return filepath.EvalSymlinks(found)
}

func sorted(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for key := range set {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func (s *Sandbox) alreadyReadable(path string) bool {
	for _, dir := range s.readOnly {
		if path == dir || strings.HasPrefix(path, dir+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

func environment(task Task, toolDirs []string) map[string]string {
	env := map[string]string{
		"HOME":   tempInside,
		"TMPDIR": tempInside,
		"PWD":    workInside,
		"PATH":   strings.Join(append(append([]string{}, toolDirs...), "/usr/bin", "/bin"), ":"),
	}
	for name, value := range task.Environment {
		env[name] = value
	}
	return env
}

func copyTree(from fs.FS, to string) error {
	return fs.WalkDir(from, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(to, filepath.FromSlash(name))
		if entry.IsDir() {
			return os.MkdirAll(target, 0o700)
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		return copyFile(from, name, target, info.Mode().Perm()|0o600)
	})
}

func copyFile(from fs.FS, name, target string, mode fs.FileMode) error {
	source, err := from.Open(name)
	if err != nil {
		return err
	}
	defer source.Close()

	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return err
	}
	sink, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(sink, source); err != nil {
		sink.Close()
		return err
	}
	return sink.Close()
}

func collect(work, into string, wanted []string) error {
	if err := os.MkdirAll(into, 0o700); err != nil {
		return fmt.Errorf("runner: %w", err)
	}
	for _, want := range wanted {
		clean := filepath.FromSlash(filepath.Clean(want))
		source := filepath.Join(work, clean)
		if _, err := os.Lstat(source); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return &MissingOutputError{Path: want}
			}
			return fmt.Errorf("runner: %w", err)
		}
		target := filepath.Join(into, clean)
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return fmt.Errorf("runner: %w", err)
		}
		if err := os.Rename(source, target); err != nil {
			return fmt.Errorf("runner: collecting %q: %w", want, err)
		}
	}
	return os.RemoveAll(work)
}

var _ Runner = (*Sandbox)(nil)
