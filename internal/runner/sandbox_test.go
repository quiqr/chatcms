package runner_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/quiqr/chatcms/internal/runner"
)

func newSandbox(t *testing.T, change ...func(*runner.Config)) *runner.Sandbox {
	t.Helper()

	cfg := runner.Config{WorkRoot: t.TempDir()}
	for _, apply := range change {
		apply(&cfg)
	}

	sandbox, err := runner.NewSandbox(cfg)
	if err != nil {
		if errors.Is(err, runner.ErrNoIsolation) {
			t.Skipf("no isolation available here: %v", err)
		}
		t.Fatalf("NewSandbox() error = %v", err)
	}
	return sandbox
}

func TestTheSandboxSatisfiesTheInterface(t *testing.T) {
	runBehaviours(t, func(t *testing.T) runner.Runner {
		return newSandbox(t)
	})
}

func run(t *testing.T, sandbox *runner.Sandbox, script string) *runner.Result {
	t.Helper()

	result, err := sandbox.Run(context.Background(), shellTask(script))
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	t.Cleanup(func() { result.Release() })
	return result
}

func TestABuildCannotReadAnotherProject(t *testing.T) {
	elsewhere := t.TempDir()
	secret := filepath.Join(elsewhere, "someone-elses-draft.md")
	if err := os.WriteFile(secret, []byte("the other customer's unpublished page"), 0o600); err != nil {
		t.Fatal(err)
	}

	sandbox := newSandbox(t)
	result := run(t, sandbox, "cat "+secret+" || echo could-not-read")

	if strings.Contains(result.Log, "unpublished page") {
		t.Fatalf("Log = %q, want the other project to be unreadable", result.Log)
	}
	if !strings.Contains(result.Log, "could-not-read") {
		t.Errorf("Log = %q, want the read to have failed", result.Log)
	}
}

func TestABuildCannotReadTheDeployKeys(t *testing.T) {
	keys := t.TempDir()
	if err := os.WriteFile(filepath.Join(keys, "site"), []byte("PRIVATE KEY MATERIAL"), 0o600); err != nil {
		t.Fatal(err)
	}

	sandbox := newSandbox(t)
	result := run(t, sandbox, "ls "+keys+" 2>&1; cat "+filepath.Join(keys, "site")+" 2>&1; echo done")

	if strings.Contains(result.Log, "PRIVATE KEY MATERIAL") {
		t.Fatalf("Log = %q, want the key to be unreadable", result.Log)
	}
}

func TestABuildCannotReachTheNetwork(t *testing.T) {
	sandbox := newSandbox(t)

	result := run(t, sandbox, "cat /proc/net/dev; echo ---; cat /proc/net/route")

	if got := interfaces(result.Log); len(got) != 1 || got[0] != "lo" {
		t.Fatalf("interfaces inside the build = %v, want loopback alone", got)
	}
	routes := strings.SplitN(result.Log, "---", 2)[1]
	if len(strings.Fields(routes)) > 11 {
		t.Errorf("the build has routes off the machine:\n%s", routes)
	}
}

// interfaces reads the names out of /proc/net/dev, which is the only honest
// place to ask: a network namespace of its own is what the build does not
// have, and /proc/net/dev is that namespace's own answer.
func interfaces(log string) []string {
	names := []string{}
	for _, line := range strings.Split(log, "\n") {
		name, _, found := strings.Cut(strings.TrimSpace(line), ":")
		if !found || strings.Contains(name, " ") {
			continue
		}
		names = append(names, name)
	}
	return names
}

func TestAskingForTheNetworkIsRefusedUnlessAllowed(t *testing.T) {
	sandbox := newSandbox(t)

	task := shellTask("true")
	task.Network = true
	_, err := sandbox.Run(context.Background(), task)

	if !errors.Is(err, runner.ErrNetworkDenied) {
		t.Fatalf("Run() error = %v, want ErrNetworkDenied", err)
	}
}

func TestTheNetworkIsThereWhenItIsBothAskedForAndAllowed(t *testing.T) {
	sandbox := newSandbox(t, func(cfg *runner.Config) { cfg.AllowNetwork = true })

	task := shellTask("cat /proc/net/dev")
	task.Network = true
	result, err := sandbox.Run(context.Background(), task)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	defer result.Release()

	onTheHost, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		t.Skipf("cannot read the host's interfaces: %v", err)
	}
	if len(interfaces(string(onTheHost))) < 2 {
		t.Skip("this machine has only loopback, so sharing its network proves nothing")
	}

	shared := interfaces(result.Log)
	alone := interfaces(run(t, sandbox, "cat /proc/net/dev").Log)

	if len(shared) <= len(alone) {
		t.Errorf("interfaces with the network = %v, without = %v; want more when it is shared", shared, alone)
	}
}

func TestABuildCannotWriteOutsideItsWorkspace(t *testing.T) {
	state := t.TempDir()

	sandbox := newSandbox(t)
	result := run(t, sandbox, "echo mine > inside.txt && echo ok-inside; echo theirs > "+filepath.Join(state, "stolen.txt")+" 2>/dev/null || echo blocked-outside")

	if !strings.Contains(result.Log, "ok-inside") {
		t.Errorf("Log = %q, want writing inside the workspace to work", result.Log)
	}
	if !strings.Contains(result.Log, "blocked-outside") {
		t.Errorf("Log = %q, want writing outside to fail", result.Log)
	}
	if _, err := os.Stat(filepath.Join(state, "stolen.txt")); err == nil {
		t.Error("the build wrote into the state directory")
	}
}

func TestTheStoreIsReadableButNotWritable(t *testing.T) {
	if _, err := os.Stat("/nix/store"); err != nil {
		t.Skip("no store on this machine")
	}

	sandbox := newSandbox(t)
	result := run(t, sandbox, "touch /nix/store/chatcms-should-not-exist 2>/dev/null && echo WROTE || echo read-only")

	if strings.Contains(result.Log, "WROTE") {
		t.Errorf("Log = %q, want the store to be read-only", result.Log)
	}
}

func TestABuildThatAsksForTooMuchMemoryDoesNotSucceed(t *testing.T) {
	sandbox := newSandbox(t)

	task := shellTask(`i=0; block=""; while [ $i -lt 200 ]; do block="$block$(head -c 1000000 /dev/zero | tr '\0' 'x')"; i=$((i+1)); done; echo ALLOCATED`)
	task.MemoryBytes = 32 << 20
	task.Timeout = 30 * time.Second

	result, err := sandbox.Run(context.Background(), task)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	defer result.Release()

	if strings.Contains(result.Log, "ALLOCATED") {
		t.Fatalf("the build allocated past its limit: %q", result.Log)
	}
	if result.Succeeded() {
		t.Errorf("Succeeded() = true, want false; log %q", result.Log)
	}
}

func TestCancellingLeavesNothingBehind(t *testing.T) {
	sandbox := newSandbox(t, func(cfg *runner.Config) { cfg.WorkRoot = t.TempDir() })

	ctx, cancel := context.WithCancel(context.Background())
	marker := "chatcms-cancel-marker-" + t.Name()

	done := make(chan error, 1)
	go func() {
		task := shellTask("echo started; sleep 120 # " + marker)
		_, err := sandbox.Run(ctx, task)
		done <- err
	}()

	time.Sleep(2 * time.Second)
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Run() error = %v, want context.Canceled", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("Run() did not return after the context was cancelled")
	}

	if stillRunning(t, marker) {
		t.Error("a process from the cancelled build is still running")
	}
}

func stillRunning(t *testing.T, marker string) bool {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)
	for {
		out, err := exec.Command("ps", "-eo", "args").Output()
		if err != nil {
			t.Skipf("cannot list processes here: %v", err)
		}
		if !strings.Contains(string(out), marker) {
			return false
		}
		if time.Now().After(deadline) {
			return true
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func TestTheWorkspaceIsGoneAfterRelease(t *testing.T) {
	root := t.TempDir()
	sandbox := newSandbox(t, func(cfg *runner.Config) { cfg.WorkRoot = root })

	task := shellTask("mkdir public && echo page > public/index.html")
	task.Outputs = []string{"public"}
	result, err := sandbox.Run(context.Background(), task)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if err := result.Release(); err != nil {
		t.Fatalf("Release() error = %v", err)
	}

	left, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) != 0 {
		t.Errorf("the work root still holds %d entries after release", len(left))
	}
}

func TestAFailedBuildLeavesNoWorkspaceBehind(t *testing.T) {
	root := t.TempDir()
	sandbox := newSandbox(t, func(cfg *runner.Config) { cfg.WorkRoot = root })

	task := shellTask("exit 0")
	task.Outputs = []string{"never-written"}
	if _, err := sandbox.Run(context.Background(), task); err == nil {
		t.Fatal("Run() error = nil, want a missing output")
	}

	left, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) != 0 {
		t.Errorf("the work root still holds %d entries after a failure", len(left))
	}
}

func TestExecutableInputsStayExecutable(t *testing.T) {
	sandbox := newSandbox(t)

	task := shellTask("./script.sh")
	task.Inputs = fstest.MapFS{
		"script.sh": {Data: []byte("#!/bin/sh\necho ran the script\n"), Mode: 0o755},
	}
	result, err := sandbox.Run(context.Background(), task)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	defer result.Release()

	if !strings.Contains(result.Log, "ran the script") {
		t.Errorf("Log = %q, want the script to have run", result.Log)
	}
}

func TestARunnerThatCannotIsolateRefusesToBeMade(t *testing.T) {
	_, err := runner.NewSandbox(runner.Config{
		BwrapBin: "chatcms-there-is-no-such-sandbox",
		WorkRoot: t.TempDir(),
	})

	if !errors.Is(err, runner.ErrNoIsolation) {
		t.Fatalf("NewSandbox() error = %v, want ErrNoIsolation", err)
	}
}

func TestAnUnknownCommandIsRefusedRatherThanRun(t *testing.T) {
	sandbox := newSandbox(t)

	task := shellTask("true")
	task.Command = "chatcms-there-is-no-such-command"
	_, err := sandbox.Run(context.Background(), task)

	var taskErr *runner.TaskError
	if !errors.As(err, &taskErr) {
		t.Fatalf("Run() error = %v, want a *TaskError", err)
	}
}

func TestAVeryLongLogIsShortenedRatherThanKeptWhole(t *testing.T) {
	sandbox := newSandbox(t, func(cfg *runner.Config) { cfg.MaxLogBytes = 4096 })

	result := run(t, sandbox, "i=0; while [ $i -lt 2000 ]; do echo 'a line of output that takes up room'; i=$((i+1)); done; echo THE-END")

	if len(result.Log) > 8192 {
		t.Errorf("Log is %d bytes, want it bounded", len(result.Log))
	}
	if !strings.Contains(result.Log, "THE-END") {
		t.Error("Log does not end with what the build last said")
	}
	if !strings.Contains(result.Log, "dropped") {
		t.Error("Log does not say that output was dropped")
	}
}

func TestAMemoryFailureIsReportedAsTheMemoryLimit(t *testing.T) {
	sandbox := newSandbox(t)

	task := shellTask(`i=0; block=""; while [ $i -lt 200 ]; do block="$block$(head -c 1000000 /dev/zero | tr '\0' 'x')"; i=$((i+1)); done; echo ALLOCATED`)
	task.MemoryBytes = 32 << 20

	result, err := sandbox.Run(context.Background(), task)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	defer result.Release()

	if result.Reached != runner.LimitMemory {
		t.Errorf("Reached = %q, want %q; log %q", result.Reached, runner.LimitMemory, result.Log)
	}
}

func TestAnOrdinaryFailureIsNotGuessedToBeMemory(t *testing.T) {
	sandbox := newSandbox(t)

	result := run(t, sandbox, "echo the template is wrong >&2; exit 1")

	if result.Reached != runner.LimitNone {
		t.Errorf("Reached = %q, want no limit to have been reached", result.Reached)
	}
}
