package workflows_test

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/quiqr/chatcms/internal/runner"
	"github.com/quiqr/chatcms/internal/sitebuild"
	"github.com/quiqr/chatcms/internal/workflows"
)

const checkedSite = "../../test/fixtures/sites/with-workflow"

func toolbox(t *testing.T) []string {
	t.Helper()

	found := []string{}
	for _, name := range []string{"sh", "grep", "cat", "ls", "echo", "test", "find", "sed"} {
		if path, err := exec.LookPath(name); err == nil {
			found = append(found, path)
		}
	}
	return found
}

func realEngine(t *testing.T, publisher workflows.Publisher) *workflows.Engine {
	t.Helper()

	sandbox, err := runner.NewSandbox(runner.Config{WorkRoot: t.TempDir()})
	if err != nil {
		if errors.Is(err, runner.ErrNoIsolation) {
			t.Skipf("no isolation available here: %v", err)
		}
		t.Fatal(err)
	}

	hugo := os.Getenv("CHATCMS_HUGO_BIN")
	if hugo == "" {
		hugo, err = exec.LookPath("hugo")
		if err != nil {
			t.Skipf("no Hugo on this machine: %v", err)
		}
	}
	builder, err := sitebuild.New(sandbox, hugo, sitebuild.Limits{Timeout: time.Minute})
	if err != nil {
		t.Fatal(err)
	}

	made, err := workflows.NewEngine(sandbox, builder, publisher, toolbox(t),
		workflows.Limits{CheckTimeout: 30 * time.Second, CheckMemory: 512 << 20})
	if err != nil {
		t.Fatal(err)
	}
	return made
}

func publishWorkflow(t *testing.T, site fs.FS) workflows.Workflow {
	t.Helper()

	set, err := workflows.Load(site)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	return set.For(workflows.OnPublish, workflows.Settings{ChatCMSBuilds: true})
}

func TestASiteChecksItselfThenBuildsThenPublishes(t *testing.T) {
	publisher := &recordedPublisher{}
	made := realEngine(t, publisher)
	site := os.DirFS(checkedSite)

	outcome, err := made.Run(context.Background(), publishWorkflow(t, site),
		workflows.Run{Site: site, BaseURL: "https://the-customer.example/"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !outcome.Succeeded() {
		t.Fatalf("Succeeded() = false: %s\n%v", outcome.Sentence(), outcome.Steps)
	}
	if len(outcome.Steps) != 3 {
		t.Fatalf("ran %d steps, want three", len(outcome.Steps))
	}
	if !strings.Contains(outcome.Steps[0].Detail, "looks finished") {
		t.Errorf("the check's output was not kept: %q", outcome.Steps[0].Detail)
	}
	if publisher.called != 1 {
		t.Fatalf("publishing was asked %d times, want once", publisher.called)
	}

	page, err := fs.ReadFile(publisher.given, "public/index.html")
	if err != nil {
		t.Fatalf("the built page did not reach publishing: %v", err)
	}
	if !strings.Contains(string(page), "A site that checks itself") {
		t.Errorf("the page handed to publishing is %q", page)
	}
}

func TestAFailingCheckPublishesNothing(t *testing.T) {
	publisher := &recordedPublisher{}
	made := realEngine(t, publisher)

	site := overlay(t, checkedSite, map[string]string{
		"content/unfinished.md": "---\ntitle: Unfinished\n---\n\nTODO: write this page.\n",
	})

	outcome, err := made.Run(context.Background(), publishWorkflow(t, site), workflows.Run{Site: site})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if outcome.Succeeded() {
		t.Fatal("Succeeded() = true, want the check to have stopped it")
	}
	if publisher.called != 0 {
		t.Error("the site was published even though its own check failed")
	}
	if len(outcome.Steps) != 1 {
		t.Fatalf("ran %d steps, want only the check", len(outcome.Steps))
	}
	if !strings.Contains(outcome.Steps[0].Detail, "still says TODO") {
		t.Errorf("the check's own words were not kept: %q", outcome.Steps[0].Detail)
	}
	if !strings.Contains(outcome.Sentence(), "Nothing changed on your live site") {
		t.Errorf("Sentence() = %q", outcome.Sentence())
	}
}

func TestACheckCannotReachTheNetwork(t *testing.T) {
	made := realEngine(t, &recordedPublisher{})

	workflow := workflows.Workflow{Trigger: workflows.OnPublish, Steps: []workflows.Step{
		{Action: workflows.Check, Run: "cat /proc/net/dev"},
	}}
	outcome, err := made.Run(context.Background(), workflow, workflows.Run{Site: os.DirFS(checkedSite)})
	if err != nil {
		t.Fatal(err)
	}

	if got := interfacesIn(outcome.Steps[0].Detail); len(got) != 1 || got[0] != "lo" {
		t.Errorf("a check saw the interfaces %v, want loopback alone", got)
	}
}

func interfacesIn(log string) []string {
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

func TestNothingACheckWroteReachesWhatIsPublished(t *testing.T) {
	publisher := &recordedPublisher{}
	made := realEngine(t, publisher)

	workflow := workflows.Workflow{Trigger: workflows.OnPublish, Steps: []workflows.Step{
		{Action: workflows.Check, Run: "echo smuggled > content/smuggled.md && echo wrote it"},
		{Action: workflows.Build},
		{Action: workflows.Publish},
	}}

	outcome, err := made.Run(context.Background(), workflow, workflows.Run{Site: os.DirFS(checkedSite)})
	if err != nil {
		t.Fatal(err)
	}
	if !outcome.Succeeded() {
		t.Fatalf("Succeeded() = false: %s", outcome.Sentence())
	}

	if _, err := fs.Stat(os.DirFS(checkedSite), "content/smuggled.md"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatal("a check wrote into the site fixture on disk")
	}
	if _, err := fs.Stat(publisher.given, "public/smuggled/index.html"); err == nil {
		t.Error("what a check wrote reached what is published")
	}
}

func TestACheckThatWillNotEndIsEnded(t *testing.T) {
	made := realEngine(t, &recordedPublisher{})

	workflow := workflows.Workflow{Trigger: workflows.OnPublish, Steps: []workflows.Step{
		{Action: workflows.Check, Run: "sleep 60", Timeout: 500 * time.Millisecond},
		{Action: workflows.Publish},
	}}

	started := time.Now()
	outcome, err := made.Run(context.Background(), workflow, workflows.Run{Site: os.DirFS(checkedSite)})
	if err != nil {
		t.Fatal(err)
	}

	if time.Since(started) > 30*time.Second {
		t.Error("the check was not ended at its limit")
	}
	if outcome.Succeeded() {
		t.Fatal("Succeeded() = true, want the check to have been ended")
	}
	if !strings.Contains(outcome.Sentence(), "took too long") {
		t.Errorf("Sentence() = %q", outcome.Sentence())
	}
}

func TestASiteWithNoWorkflowStillPublishes(t *testing.T) {
	publisher := &recordedPublisher{}
	made := realEngine(t, publisher)
	site := os.DirFS("../../test/fixtures/sites/small-site")

	workflow := publishWorkflow(t, site)
	outcome, err := made.Run(context.Background(), workflow, workflows.Run{Site: site})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !outcome.Succeeded() {
		t.Fatalf("Succeeded() = false: %s", outcome.Sentence())
	}
	if publisher.called != 1 {
		t.Errorf("publishing was asked %d times, want once", publisher.called)
	}
	if !outcome.Ran(workflows.Build) {
		t.Error("a project that publishes built pages did not build")
	}
}

func overlay(t *testing.T, base string, extra map[string]string) fs.FS {
	t.Helper()

	combined := fstest.MapFS{}
	err := fs.WalkDir(os.DirFS(base), ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		content, err := fs.ReadFile(os.DirFS(base), name)
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		combined[name] = &fstest.MapFile{Data: content, Mode: info.Mode().Perm()}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for name, content := range extra {
		combined[name] = &fstest.MapFile{Data: []byte(content), Mode: 0o644}
	}
	return combined
}
