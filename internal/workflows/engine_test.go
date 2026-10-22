package workflows_test

import (
	"context"
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/runner"
	"github.com/quiqr/chatcms/internal/sitebuild"
	"github.com/quiqr/chatcms/internal/workflows"
)

type recordedRunner struct {
	tasks  []runner.Task
	result *runner.Result
	err    error
}

func (r *recordedRunner) Run(ctx context.Context, task runner.Task) (*runner.Result, error) {
	r.tasks = append(r.tasks, task)
	if r.err != nil {
		return nil, r.err
	}
	if r.result == nil {
		return &runner.Result{}, nil
	}
	return r.result, nil
}

type recordedBuilder struct {
	called  int
	result  *runner.Result
	err     error
	options sitebuild.Options
}

func (b *recordedBuilder) Build(ctx context.Context, site fs.FS, options sitebuild.Options) (*runner.Result, error) {
	b.called++
	b.options = options
	if b.err != nil {
		return nil, b.err
	}
	if b.result == nil {
		return &runner.Result{}, nil
	}
	return b.result, nil
}

type recordedPublisher struct {
	called int
	given  fs.FS
	err    error
}

// Publish copies what it is given rather than keeping the filesystem, which
// is what a real publisher has to do: the workspace is released when the
// workflow ends.
func (p *recordedPublisher) Publish(ctx context.Context, projectID uuid.UUID, built fs.FS) error {
	p.called++
	p.given = copyOf(built)
	return p.err
}

func copyOf(given fs.FS) fs.FS {
	if given == nil {
		return nil
	}
	taken := fstest.MapFS{}
	fs.WalkDir(given, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		content, err := fs.ReadFile(given, name)
		if err != nil {
			return err
		}
		taken[name] = &fstest.MapFile{Data: content}
		return nil
	})
	return taken
}

func engine(t *testing.T, r runner.Runner, b workflows.SiteBuilder, p workflows.Publisher) *workflows.Engine {
	t.Helper()

	made, err := workflows.NewEngine(r, b, p, nil, workflows.Limits{})
	if err != nil {
		t.Fatal(err)
	}
	return made
}

func aRun() workflows.Run {
	return workflows.Run{
		ProjectID: uuid.New(),
		Site:      fstest.MapFS{"hugo.toml": {Data: []byte("title = \"a site\"\n")}},
	}
}

func TestStepsRunInOrder(t *testing.T) {
	publisher := &recordedPublisher{}
	builder := &recordedBuilder{}
	made := engine(t, &recordedRunner{}, builder, publisher)

	workflow := workflows.Workflow{Trigger: workflows.OnPublish, Steps: []workflows.Step{
		{Action: workflows.Check, Run: "./check.sh"},
		{Action: workflows.Build},
		{Action: workflows.Publish},
	}}

	outcome, err := made.Run(context.Background(), workflow, aRun())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !outcome.Succeeded() {
		t.Fatalf("Succeeded() = false: %s", outcome.Sentence())
	}
	want := []workflows.Action{workflows.Check, workflows.Build, workflows.Publish}
	if len(outcome.Steps) != len(want) {
		t.Fatalf("ran %d steps, want %d", len(outcome.Steps), len(want))
	}
	for index, action := range want {
		if outcome.Steps[index].Action != action {
			t.Errorf("step %d was %q, want %q", index+1, outcome.Steps[index].Action, action)
		}
		if outcome.Steps[index].State != workflows.Succeeded {
			t.Errorf("step %d state = %q", index+1, outcome.Steps[index].State)
		}
	}
}

func TestAFailedCheckStopsEverythingAfterIt(t *testing.T) {
	publisher := &recordedPublisher{}
	builder := &recordedBuilder{}
	checks := &recordedRunner{result: &runner.Result{ExitCode: 1, Log: "three links are broken"}}
	made := engine(t, checks, builder, publisher)

	workflow := workflows.Workflow{Trigger: workflows.OnPublish, Steps: []workflows.Step{
		{Action: workflows.Check, Run: "./check.sh"},
		{Action: workflows.Build},
		{Action: workflows.Publish},
	}}

	outcome, err := made.Run(context.Background(), workflow, aRun())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if outcome.Succeeded() {
		t.Fatal("Succeeded() = true, want the failed check to have stopped it")
	}
	if builder.called != 0 {
		t.Error("the site was built after the check failed")
	}
	if publisher.called != 0 {
		t.Error("the site was published after the check failed")
	}
	if len(outcome.Steps) != 1 {
		t.Errorf("reported %d steps, want only the one that ran", len(outcome.Steps))
	}
	if outcome.Steps[0].Detail != "three links are broken" {
		t.Errorf("Detail = %q, want what the check said", outcome.Steps[0].Detail)
	}
}

func TestWhatRanBeforeTheFailureIsStillReported(t *testing.T) {
	made := engine(t, &recordedRunner{}, &recordedBuilder{result: &runner.Result{ExitCode: 1, Log: "a template is wrong"}}, &recordedPublisher{})

	workflow := workflows.Workflow{Trigger: workflows.OnPublish, Steps: []workflows.Step{
		{Action: workflows.Check, Run: "./one.sh"},
		{Action: workflows.Check, Run: "./two.sh"},
		{Action: workflows.Build},
		{Action: workflows.Publish},
	}}

	outcome, err := made.Run(context.Background(), workflow, aRun())
	if err != nil {
		t.Fatal(err)
	}

	if len(outcome.Steps) != 3 {
		t.Fatalf("reported %d steps, want the two that worked and the one that failed", len(outcome.Steps))
	}
	if outcome.Steps[2].State != workflows.Failed {
		t.Errorf("the third step state = %q, want failed", outcome.Steps[2].State)
	}
	if outcome.Trouble.Step != "Building the site" {
		t.Errorf("Trouble.Step = %q, want the step that failed", outcome.Trouble.Step)
	}
}

func TestBuildingHandsItsResultToPublishing(t *testing.T) {
	pages := fstest.MapFS{"public/index.html": {Data: []byte("the built page")}}
	publisher := &recordedPublisher{}
	made := engine(t, &recordedRunner{}, &recordedBuilder{result: &runner.Result{Outputs: pages}}, publisher)

	workflow := workflows.Workflow{Trigger: workflows.OnPublish, Steps: []workflows.Step{
		{Action: workflows.Build},
		{Action: workflows.Publish},
	}}

	if _, err := made.Run(context.Background(), workflow, aRun()); err != nil {
		t.Fatal(err)
	}

	if publisher.given == nil {
		t.Fatal("publishing was given nothing after a build")
	}
	page, err := fs.ReadFile(publisher.given, "index.html")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(page) != "the built page" {
		t.Errorf("publishing was given %q", page)
	}
}

func TestPublishingWithoutABuildIsGivenNothing(t *testing.T) {
	publisher := &recordedPublisher{}
	made := engine(t, &recordedRunner{}, &recordedBuilder{}, publisher)

	workflow := workflows.Workflow{Trigger: workflows.OnPublish, Steps: []workflows.Step{
		{Action: workflows.Publish},
	}}

	if _, err := made.Run(context.Background(), workflow, aRun()); err != nil {
		t.Fatal(err)
	}

	if publisher.given != nil {
		t.Error("publishing was given pages even though nothing was built")
	}
}

func TestWhatABuildMadeIsReleasedWhateverHappened(t *testing.T) {
	for name, publishErr := range map[string]error{
		"when publishing worked": nil,
		"when publishing failed": errors.New("the remote refused"),
	} {
		t.Run(name, func(t *testing.T) {
			released := false
			result := &runner.Result{
				Outputs:   fstest.MapFS{"public/index.html": {Data: []byte("page")}},
				OnRelease: func() error { released = true; return nil },
			}
			made := engine(t, &recordedRunner{}, &recordedBuilder{result: result}, &recordedPublisher{err: publishErr})

			workflow := workflows.Workflow{Trigger: workflows.OnPublish, Steps: []workflows.Step{
				{Action: workflows.Build},
				{Action: workflows.Publish},
			}}
			if _, err := made.Run(context.Background(), workflow, aRun()); err != nil {
				t.Fatal(err)
			}

			if !released {
				t.Error("what the build made was not released")
			}
		})
	}
}

func TestAWorkflowWithNoStepsSucceedsAndDoesNothing(t *testing.T) {
	publisher := &recordedPublisher{}
	builder := &recordedBuilder{}
	made := engine(t, &recordedRunner{}, builder, publisher)

	outcome, err := made.Run(context.Background(), workflows.Workflow{Trigger: workflows.OnSave}, aRun())
	if err != nil {
		t.Fatal(err)
	}

	if !outcome.Succeeded() {
		t.Error("a workflow with no steps failed")
	}
	if len(outcome.Steps) != 0 || builder.called != 0 || publisher.called != 0 {
		t.Error("a workflow with no steps did something")
	}
}

func TestEveryWayAStepCanFailSaysSomethingAnEditorCanRead(t *testing.T) {
	cases := map[string]struct {
		engine func(*testing.T) *workflows.Engine
		step   workflows.Step
		want   string
	}{
		"a check that found something": {
			engine: func(t *testing.T) *workflows.Engine {
				return engine(t, &recordedRunner{result: &runner.Result{ExitCode: 1}}, &recordedBuilder{}, &recordedPublisher{})
			},
			step: workflows.Step{Action: workflows.Check, Run: "./check.sh"},
			want: "needs fixing",
		},
		"a check that took too long": {
			engine: func(t *testing.T) *workflows.Engine {
				return engine(t, &recordedRunner{result: &runner.Result{ExitCode: -1, Reached: runner.LimitTime}}, &recordedBuilder{}, &recordedPublisher{})
			},
			step: workflows.Step{Action: workflows.Check, Run: "./check.sh"},
			want: "took too long",
		},
		"a check that wanted too much memory": {
			engine: func(t *testing.T) *workflows.Engine {
				return engine(t, &recordedRunner{result: &runner.Result{ExitCode: 2, Reached: runner.LimitMemory}}, &recordedBuilder{}, &recordedPublisher{})
			},
			step: workflows.Step{Action: workflows.Check, Run: "./check.sh"},
			want: "memory",
		},
		"a check that could not be contained": {
			engine: func(t *testing.T) *workflows.Engine {
				return engine(t, &recordedRunner{err: runner.ErrNoIsolation}, &recordedBuilder{}, &recordedPublisher{})
			},
			step: workflows.Step{Action: workflows.Check, Run: "./check.sh"},
			want: "safely",
		},
		"a site that would not build": {
			engine: func(t *testing.T) *workflows.Engine {
				return engine(t, &recordedRunner{}, &recordedBuilder{result: &runner.Result{ExitCode: 1}}, &recordedPublisher{})
			},
			step: workflows.Step{Action: workflows.Build},
			want: "could not be built",
		},
		"a build that could not be started": {
			engine: func(t *testing.T) *workflows.Engine {
				return engine(t, &recordedRunner{}, &recordedBuilder{err: runner.ErrNoIsolation}, &recordedPublisher{})
			},
			step: workflows.Step{Action: workflows.Build},
			want: "could not be built",
		},
		"a remote that could not be reached": {
			engine: func(t *testing.T) *workflows.Engine {
				return engine(t, &recordedRunner{}, &recordedBuilder{}, &recordedPublisher{err: errors.New("ssh: connect to host example.org port 22: Connection refused")})
			},
			step: workflows.Step{Action: workflows.Publish},
			want: "could not reach",
		},
	}

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			workflow := workflows.Workflow{Trigger: workflows.OnPublish, Steps: []workflows.Step{test.step}}

			outcome, err := test.engine(t).Run(context.Background(), workflow, aRun())
			if err != nil {
				t.Fatalf("Run() error = %v, want the failure in the outcome", err)
			}
			if outcome.Succeeded() {
				t.Fatal("Succeeded() = true, want a failure")
			}

			sentence := outcome.Sentence()
			if !strings.Contains(sentence, test.want) {
				t.Errorf("Sentence() = %q, want it to mention %q", sentence, test.want)
			}
			if !strings.Contains(sentence, "Nothing changed on your live site") {
				t.Errorf("Sentence() = %q, want the reassurance", sentence)
			}
		})
	}
}

// gitWords are the words docs/design/DESIGN.md section 7 forbids in anything
// an editor reads. A failure sentence is the most likely place for one of
// them to slip in, because the thing that failed usually is git.
var gitWords = []string{
	"commit", "branch", "push", "pull", "merge", "repository", "repo",
	"deploy", "build log", "HEAD", "SHA", "conflict", "remote", "ssh", "git",
}

func TestNoFailureSentenceCarriesAGitWord(t *testing.T) {
	cases := []struct {
		trigger workflows.Trigger
		engine  func(*testing.T) *workflows.Engine
		step    workflows.Step
	}{
		{workflows.OnPublish, func(t *testing.T) *workflows.Engine {
			return engine(t, &recordedRunner{result: &runner.Result{ExitCode: 1, Log: "fatal: not a git repository"}}, &recordedBuilder{}, &recordedPublisher{})
		}, workflows.Step{Action: workflows.Check, Run: "./check.sh"}},
		{workflows.OnSave, func(t *testing.T) *workflows.Engine {
			return engine(t, &recordedRunner{}, &recordedBuilder{result: &runner.Result{ExitCode: 1, Log: "error: failed to push some refs"}}, &recordedPublisher{})
		}, workflows.Step{Action: workflows.Build}},
		{workflows.OnPreview, func(t *testing.T) *workflows.Engine {
			return engine(t, &recordedRunner{}, &recordedBuilder{}, &recordedPublisher{err: errors.New("git push failed: remote rejected the branch")})
		}, workflows.Step{Action: workflows.Publish}},
	}

	for _, test := range cases {
		t.Run(string(test.trigger), func(t *testing.T) {
			workflow := workflows.Workflow{Trigger: test.trigger, Steps: []workflows.Step{test.step}}

			outcome, err := test.engine(t).Run(context.Background(), workflow, aRun())
			if err != nil {
				t.Fatal(err)
			}

			sentence := strings.ToLower(outcome.Sentence())
			for _, word := range gitWords {
				if strings.Contains(sentence, strings.ToLower(word)) {
					t.Errorf("Sentence() = %q, which carries the word %q", outcome.Sentence(), word)
				}
			}
			for _, step := range outcome.Steps {
				if strings.Contains(strings.ToLower(step.Title), "git") {
					t.Errorf("a step is called %q", step.Title)
				}
			}
		})
	}
}

func TestTheRawOutputIsKeptApartFromTheSentence(t *testing.T) {
	made := engine(t, &recordedRunner{result: &runner.Result{ExitCode: 1, Log: "fatal: not a git repository"}}, &recordedBuilder{}, &recordedPublisher{})
	workflow := workflows.Workflow{Trigger: workflows.OnPublish, Steps: []workflows.Step{
		{Action: workflows.Check, Run: "./check.sh"},
	}}

	outcome, err := made.Run(context.Background(), workflow, aRun())
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(outcome.Sentence(), "fatal") {
		t.Error("the raw output ended up in the sentence an editor reads")
	}
	if !strings.Contains(outcome.Steps[0].Detail, "fatal") {
		t.Error("the raw output was not kept on the step")
	}
}

func TestAVeryLongDetailIsShortened(t *testing.T) {
	long := strings.Repeat("a line of noise\n", 100000) + "THE REAL ERROR"
	made := engine(t, &recordedRunner{result: &runner.Result{ExitCode: 1, Log: long}}, &recordedBuilder{}, &recordedPublisher{})
	workflow := workflows.Workflow{Trigger: workflows.OnPublish, Steps: []workflows.Step{
		{Action: workflows.Check, Run: "./check.sh"},
	}}

	outcome, err := made.Run(context.Background(), workflow, aRun())
	if err != nil {
		t.Fatal(err)
	}

	detail := outcome.Steps[0].Detail
	if len(detail) > workflows.MaxDetailBytes+200 {
		t.Errorf("the detail is %d bytes, want it bounded", len(detail))
	}
	if !strings.HasSuffix(detail, "THE REAL ERROR") {
		t.Error("the detail does not end with the error")
	}
	if !strings.Contains(detail, "dropped") {
		t.Error("the detail does not say output was dropped")
	}
}

func TestACheckAsksForNoNetworkAndGetsItsToolbox(t *testing.T) {
	checks := &recordedRunner{}
	made, err := workflows.NewEngine(checks, &recordedBuilder{}, &recordedPublisher{},
		[]string{"/nix/store/bash/bin/sh", "/nix/store/coreutils/bin/cat"},
		workflows.Limits{CheckTimeout: 30 * time.Second, CheckMemory: 64 << 20})
	if err != nil {
		t.Fatal(err)
	}

	workflow := workflows.Workflow{Trigger: workflows.OnPublish, Steps: []workflows.Step{
		{Action: workflows.Check, Run: "./check.sh"},
	}}
	if _, err := made.Run(context.Background(), workflow, aRun()); err != nil {
		t.Fatal(err)
	}

	if len(checks.tasks) != 1 {
		t.Fatalf("the runner was asked %d times, want once", len(checks.tasks))
	}
	task := checks.tasks[0]
	if task.Network {
		t.Error("a check asked for the network")
	}
	if len(task.Tools) != 2 {
		t.Errorf("Tools = %v, want the toolbox", task.Tools)
	}
	if task.Timeout != 30*time.Second || task.MemoryBytes != 64<<20 {
		t.Errorf("limits = %v and %d, want the ones the engine was given", task.Timeout, task.MemoryBytes)
	}
	if task.Inputs == nil {
		t.Error("a check was not given the site")
	}
	if len(task.Outputs) != 0 {
		t.Error("a check was asked to produce output; nothing it writes should escape")
	}
}

func TestAStepCanAskForLessTimeThanTheDefault(t *testing.T) {
	checks := &recordedRunner{}
	made := engine(t, checks, &recordedBuilder{}, &recordedPublisher{})

	workflow := workflows.Workflow{Trigger: workflows.OnPublish, Steps: []workflows.Step{
		{Action: workflows.Check, Run: "./check.sh", Timeout: 15 * time.Second},
	}}
	if _, err := made.Run(context.Background(), workflow, aRun()); err != nil {
		t.Fatal(err)
	}

	if checks.tasks[0].Timeout != 15*time.Second {
		t.Errorf("Timeout = %v, want what the step asked for", checks.tasks[0].Timeout)
	}
}

func TestTheLimitsHaveDefaults(t *testing.T) {
	checks := &recordedRunner{}
	made := engine(t, checks, &recordedBuilder{}, &recordedPublisher{})

	workflow := workflows.Workflow{Trigger: workflows.OnPublish, Steps: []workflows.Step{
		{Action: workflows.Check, Run: "./check.sh"},
	}}
	if _, err := made.Run(context.Background(), workflow, aRun()); err != nil {
		t.Fatal(err)
	}

	if checks.tasks[0].Timeout != workflows.DefaultCheckTimeout {
		t.Errorf("Timeout = %v, want the default", checks.tasks[0].Timeout)
	}
	if checks.tasks[0].MemoryBytes != workflows.DefaultCheckMemory {
		t.Errorf("MemoryBytes = %d, want the default", checks.tasks[0].MemoryBytes)
	}
}

func TestNewEngineRefusesAMissingPart(t *testing.T) {
	cases := map[string]struct {
		runner    runner.Runner
		builder   workflows.SiteBuilder
		publisher workflows.Publisher
	}{
		"no runner":    {nil, &recordedBuilder{}, &recordedPublisher{}},
		"no builder":   {&recordedRunner{}, nil, &recordedPublisher{}},
		"no publisher": {&recordedRunner{}, &recordedBuilder{}, nil},
	}

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := workflows.NewEngine(test.runner, test.builder, test.publisher, nil, workflows.Limits{}); !errors.Is(err, workflows.ErrIncomplete) {
				t.Errorf("NewEngine() error = %v, want ErrIncomplete", err)
			}
		})
	}
}

func TestRunRefusesWithNoSite(t *testing.T) {
	made := engine(t, &recordedRunner{}, &recordedBuilder{}, &recordedPublisher{})

	run := aRun()
	run.Site = nil
	if _, err := made.Run(context.Background(), workflows.Workflow{Trigger: workflows.OnPublish}, run); err == nil {
		t.Error("Run() with no site error = nil, want an error")
	}
}

func TestTheBaseURLReachesTheBuild(t *testing.T) {
	builder := &recordedBuilder{}
	made := engine(t, &recordedRunner{}, builder, &recordedPublisher{})

	run := aRun()
	run.BaseURL = "https://the-customer.example/"
	workflow := workflows.Workflow{Trigger: workflows.OnPublish, Steps: []workflows.Step{{Action: workflows.Build}}}
	if _, err := made.Run(context.Background(), workflow, run); err != nil {
		t.Fatal(err)
	}

	if builder.options.BaseURL != "https://the-customer.example/" {
		t.Errorf("BaseURL = %q, want the one the run carried", builder.options.BaseURL)
	}
}

func TestASucceededOutcomeSaysNothing(t *testing.T) {
	made := engine(t, &recordedRunner{}, &recordedBuilder{}, &recordedPublisher{})

	outcome, err := made.Run(context.Background(), workflows.Workflow{Trigger: workflows.OnPublish}, aRun())
	if err != nil {
		t.Fatal(err)
	}

	if outcome.Sentence() != "" {
		t.Errorf("Sentence() = %q, want nothing when nothing went wrong", outcome.Sentence())
	}
	var absent *workflows.Outcome
	if absent.Succeeded() || absent.Sentence() != "" {
		t.Error("an absent outcome claimed something")
	}
}
