package runner_test

import (
	"context"
	"errors"
	"io/fs"
	"os/exec"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/quiqr/chatcms/internal/runner"
)

func TestValidateRefusesAnUnusableTask(t *testing.T) {
	usable := func() runner.Task {
		return runner.Task{
			Command:     "sh",
			Timeout:     time.Minute,
			MemoryBytes: 1 << 28,
		}
	}

	cases := map[string]func(*runner.Task){
		"no command":        func(task *runner.Task) { task.Command = "" },
		"no time limit":     func(task *runner.Task) { task.Timeout = 0 },
		"negative time":     func(task *runner.Task) { task.Timeout = -time.Second },
		"no memory limit":   func(task *runner.Task) { task.MemoryBytes = 0 },
		"absolute output":   func(task *runner.Task) { task.Outputs = []string{"/etc/passwd"} },
		"climbing output":   func(task *runner.Task) { task.Outputs = []string{"../../elsewhere"} },
		"sneaky output":     func(task *runner.Task) { task.Outputs = []string{"public/../../elsewhere"} },
		"empty output":      func(task *runner.Task) { task.Outputs = []string{""} },
		"environment name":  func(task *runner.Task) { task.Environment = map[string]string{"A=B": "c"} },
		"environment empty": func(task *runner.Task) { task.Environment = map[string]string{"": "c"} },
	}

	for name, break_ := range cases {
		t.Run(name, func(t *testing.T) {
			task := usable()
			break_(&task)

			err := runner.Validate(task)

			var taskErr *runner.TaskError
			if !errors.As(err, &taskErr) {
				t.Fatalf("Validate() error = %v, want a *TaskError", err)
			}
		})
	}
}

func TestValidateAcceptsAWorkableTask(t *testing.T) {
	err := runner.Validate(runner.Task{
		Command:     "hugo",
		Arguments:   []string{"--minify"},
		Outputs:     []string{"public", "nested/thing.json"},
		Environment: map[string]string{"HUGO_ENV": "production"},
		Timeout:     time.Minute,
		MemoryBytes: 1 << 28,
	})
	if err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

func TestResultSucceededNeedsBothAZeroExitAndNoLimit(t *testing.T) {
	cases := map[string]struct {
		result runner.Result
		want   bool
	}{
		"clean":         {runner.Result{ExitCode: 0, Reached: runner.LimitNone}, true},
		"non-zero exit": {runner.Result{ExitCode: 1, Reached: runner.LimitNone}, false},
		"out of time":   {runner.Result{ExitCode: 0, Reached: runner.LimitTime}, false},
		"out of memory": {runner.Result{ExitCode: 0, Reached: runner.LimitMemory}, false},
		"both at once":  {runner.Result{ExitCode: 137, Reached: runner.LimitTime}, false},
	}

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			if got := test.result.Succeeded(); got != test.want {
				t.Errorf("Succeeded() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestReleaseIsSafeToCallTwiceAndOnNothing(t *testing.T) {
	var absent *runner.Result
	if err := absent.Release(); err != nil {
		t.Errorf("Release() on a nil result = %v, want nil", err)
	}

	result := &runner.Result{}
	if err := result.Release(); err != nil {
		t.Errorf("first Release() = %v, want nil", err)
	}
	if err := result.Release(); err != nil {
		t.Errorf("second Release() = %v, want nil", err)
	}
}

func TestTailWriterKeepsTheEndAndSaysSo(t *testing.T) {
	long := strings.Repeat("a", 100) + "the error is here"

	got := runner.TrimLog(long, 40)

	if !strings.HasSuffix(got, "the error is here") {
		t.Errorf("TrimLog() = %q, want it to end with the error", got)
	}
	if !strings.Contains(got, "dropped") {
		t.Errorf("TrimLog() = %q, want it to say output was dropped", got)
	}
}

func TestTrimLogLeavesAShortLogAlone(t *testing.T) {
	if got := runner.TrimLog("short", 100); got != "short" {
		t.Errorf("TrimLog() = %q, want it untouched", got)
	}
	if got := runner.TrimLog("short", 0); got != "short" {
		t.Errorf("TrimLog() with no limit = %q, want it untouched", got)
	}
}

func TestTrimLogCutsAtALineBoundary(t *testing.T) {
	log := "first line\nsecond line\nthird line\n"

	got := runner.TrimLog(log, 20)

	body := got[strings.Index(got, "\n")+1:]
	if strings.HasPrefix(body, "ine") || strings.HasPrefix(body, "d line") {
		t.Errorf("TrimLog() = %q, want it to start at a line rather than mid-word", got)
	}
}

// behaviour is the table every Runner must satisfy, whatever it does
// underneath. It is run against the sandbox and against a runner that
// isolates nothing, so that a property only one of them can have cannot
// quietly become part of the interface.
type behaviour struct {
	name    string
	task    func() runner.Task
	check   func(*testing.T, *runner.Result)
	wantErr func(error) bool
}

func shellTask(script string) runner.Task {
	return runner.Task{
		Command:     "sh",
		Arguments:   []string{"-c", script},
		Tools:       shellTools(),
		Timeout:     30 * time.Second,
		MemoryBytes: 1 << 29,
	}
}

// shellTools names the utilities the scripts in these tests reach for. A real
// task names the tools it needs the same way; nothing is on the path inside a
// build that the task did not ask for.
func shellTools() []string {
	wanted := []string{"cat", "echo", "ls", "mkdir", "touch", "head", "tr", "wc", "sleep", "env"}
	found := []string{}
	for _, name := range wanted {
		if path, err := exec.LookPath(name); err == nil {
			found = append(found, path)
		}
	}
	return found
}

func behaviours() []behaviour {
	return []behaviour{
		{
			name: "a command that works reports a zero exit",
			task: func() runner.Task { return shellTask("exit 0") },
			check: func(t *testing.T, result *runner.Result) {
				if !result.Succeeded() {
					t.Errorf("Succeeded() = false, exit %d, log %q", result.ExitCode, result.Log)
				}
			},
		},
		{
			name: "a command that fails reports its exit code",
			task: func() runner.Task { return shellTask("exit 3") },
			check: func(t *testing.T, result *runner.Result) {
				if result.ExitCode != 3 {
					t.Errorf("ExitCode = %d, want 3", result.ExitCode)
				}
				if result.Succeeded() {
					t.Error("Succeeded() = true, want false")
				}
			},
		},
		{
			name: "both streams end up in the log",
			task: func() runner.Task { return shellTask("echo out; echo err >&2") },
			check: func(t *testing.T, result *runner.Result) {
				for _, want := range []string{"out", "err"} {
					if !strings.Contains(result.Log, want) {
						t.Errorf("Log = %q, want it to contain %q", result.Log, want)
					}
				}
			},
		},
		{
			name: "the inputs are there to read",
			task: func() runner.Task {
				task := shellTask("cat greeting.txt && cat nested/deep.txt")
				task.Inputs = fstest.MapFS{
					"greeting.txt":    {Data: []byte("hello\n")},
					"nested/deep.txt": {Data: []byte("deeper\n")},
				}
				return task
			},
			check: func(t *testing.T, result *runner.Result) {
				for _, want := range []string{"hello", "deeper"} {
					if !strings.Contains(result.Log, want) {
						t.Errorf("Log = %q, want it to contain %q", result.Log, want)
					}
				}
			},
		},
		{
			name: "a named output comes back whole",
			task: func() runner.Task {
				task := shellTask("mkdir -p public/css && echo page > public/index.html && echo body > public/css/site.css")
				task.Outputs = []string{"public"}
				return task
			},
			check: func(t *testing.T, result *runner.Result) {
				want := map[string]string{
					"public/index.html":   "page\n",
					"public/css/site.css": "body\n",
				}
				for name, content := range want {
					got, err := fs.ReadFile(result.Outputs, name)
					if err != nil {
						t.Fatalf("ReadFile(%q) error = %v", name, err)
					}
					if string(got) != content {
						t.Errorf("%s = %q, want %q", name, got, content)
					}
				}
			},
		},
		{
			name: "an unnamed file does not come back",
			task: func() runner.Task {
				task := shellTask("mkdir public && echo page > public/index.html && echo secret > leftover.txt")
				task.Outputs = []string{"public"}
				return task
			},
			check: func(t *testing.T, result *runner.Result) {
				if _, err := fs.Stat(result.Outputs, "leftover.txt"); !errors.Is(err, fs.ErrNotExist) {
					t.Errorf("Stat(leftover.txt) error = %v, want it absent", err)
				}
			},
		},
		{
			name: "a missing output is an error rather than an empty result",
			task: func() runner.Task {
				task := shellTask("exit 0")
				task.Outputs = []string{"public"}
				return task
			},
			wantErr: func(err error) bool {
				var missing *runner.MissingOutputError
				return errors.As(err, &missing) && missing.Path == "public"
			},
		},
		{
			name: "only the named environment is there",
			task: func() runner.Task {
				task := shellTask("echo wanted=$CHATCMS_WANTED; echo leaked=$CHATCMS_LEAKED")
				task.Environment = map[string]string{"CHATCMS_WANTED": "yes"}
				return task
			},
			check: func(t *testing.T, result *runner.Result) {
				if !strings.Contains(result.Log, "wanted=yes") {
					t.Errorf("Log = %q, want the named variable", result.Log)
				}
				if !strings.Contains(result.Log, "leaked=\n") {
					t.Errorf("Log = %q, want the server's variable to be absent", result.Log)
				}
			},
		},
		{
			name: "a build that runs too long is ended",
			task: func() runner.Task {
				task := shellTask("sleep 60")
				task.Timeout = 500 * time.Millisecond
				return task
			},
			check: func(t *testing.T, result *runner.Result) {
				if result.Reached != runner.LimitTime {
					t.Errorf("Reached = %q, want %q", result.Reached, runner.LimitTime)
				}
				if result.Succeeded() {
					t.Error("Succeeded() = true, want false")
				}
				if result.Duration > 30*time.Second {
					t.Errorf("Duration = %v, want it to end at the limit", result.Duration)
				}
			},
		},
		{
			name: "the output before the limit is kept",
			task: func() runner.Task {
				task := shellTask("echo got this far; sleep 60")
				task.Timeout = time.Second
				return task
			},
			check: func(t *testing.T, result *runner.Result) {
				if !strings.Contains(result.Log, "got this far") {
					t.Errorf("Log = %q, want what was written before the limit", result.Log)
				}
			},
		},
	}
}

func runBehaviours(t *testing.T, make func(*testing.T) runner.Runner) {
	t.Helper()

	for _, test := range behaviours() {
		t.Run(test.name, func(t *testing.T) {
			r := make(t)

			result, err := r.Run(context.Background(), test.task())

			if test.wantErr != nil {
				if !test.wantErr(err) {
					t.Fatalf("Run() error = %v, want the stated failure", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Run() error = %v, want nil", err)
			}
			defer result.Release()
			test.check(t, result)
		})
	}
}
