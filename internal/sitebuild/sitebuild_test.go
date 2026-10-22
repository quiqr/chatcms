package sitebuild_test

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/quiqr/chatcms/internal/runner"
	"github.com/quiqr/chatcms/internal/sitebuild"
)

const fixture = "../../test/fixtures/sites/small-site"

func hugoBin(t *testing.T) string {
	t.Helper()

	if named := os.Getenv("CHATCMS_HUGO_BIN"); named != "" {
		return named
	}
	found, err := exec.LookPath("hugo")
	if err != nil {
		t.Skipf("no Hugo on this machine: %v", err)
	}
	return found
}

func newBuilder(t *testing.T) *sitebuild.Builder {
	t.Helper()

	sandbox, err := runner.NewSandbox(runner.Config{WorkRoot: t.TempDir()})
	if err != nil {
		if errors.Is(err, runner.ErrNoIsolation) {
			t.Skipf("no isolation available here: %v", err)
		}
		t.Fatal(err)
	}
	builder, err := sitebuild.New(sandbox, hugoBin(t), sitebuild.Limits{Timeout: 2 * time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	return builder
}

func TestABuiltSiteMatchesHugoRunDirectly(t *testing.T) {
	builder := newBuilder(t)

	result, err := builder.Build(context.Background(), os.DirFS(fixture), sitebuild.Options{})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	defer result.Release()

	if !result.Succeeded() {
		t.Fatalf("Succeeded() = false, exit %d, log:\n%s", result.ExitCode, result.Log)
	}

	want := buildDirectly(t)
	got := readAll(t, result.Outputs, sitebuild.Output)

	compare(t, want, got)
}

func buildDirectly(t *testing.T) map[string]string {
	t.Helper()

	into := filepath.Join(t.TempDir(), "public")
	out, err := exec.Command(hugoBin(t),
		"--source", fixture,
		"--destination", into,
		"--cleanDestinationDir",
		"--cacheDir", filepath.Join(t.TempDir(), "cache"),
		"--environment", "production",
		"--noBuildLock",
	).CombinedOutput()
	if err != nil {
		t.Fatalf("Hugo run directly failed: %v\n%s", err, out)
	}
	return readAll(t, os.DirFS(into), ".")
}

func readAll(t *testing.T, from fs.FS, root string) map[string]string {
	t.Helper()

	files := map[string]string{}
	err := fs.WalkDir(from, root, func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		content, err := fs.ReadFile(from, name)
		if err != nil {
			return err
		}
		relative := strings.TrimPrefix(strings.TrimPrefix(name, root), "/")
		files[relative] = string(content)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func compare(t *testing.T, want, got map[string]string) {
	t.Helper()

	if len(want) == 0 {
		t.Fatal("the direct build produced nothing; the fixture is wrong")
	}
	for _, name := range names(want) {
		content, present := got[name]
		if !present {
			t.Errorf("%s: missing from the sandboxed build", name)
			continue
		}
		if content != want[name] {
			t.Errorf("%s: differs between the two builds", name)
		}
	}
	for _, name := range names(got) {
		if _, present := want[name]; !present {
			t.Errorf("%s: the sandboxed build produced a file the direct build did not", name)
		}
	}
}

func names(files map[string]string) []string {
	out := make([]string, 0, len(files))
	for name := range files {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func TestTheBaseURLIsApplied(t *testing.T) {
	builder := newBuilder(t)

	result, err := builder.Build(context.Background(), os.DirFS(fixture), sitebuild.Options{
		BaseURL: "https://the-customer.example/",
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	defer result.Release()

	page, err := fs.ReadFile(result.Outputs, "public/sitemap.xml")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(page), "https://the-customer.example/") {
		t.Errorf("the base URL was not applied:\n%s", page)
	}
}

func TestASiteThatFailsToBuildSaysWhy(t *testing.T) {
	builder := newBuilder(t)

	broken := fstest.MapFS{
		"hugo.toml":          {Data: []byte("title = \"Broken\"\n")},
		"layouts/index.html": {Data: []byte("{{ .ThisIsNotAThing }}\n")},
		"content/_index.md":  {Data: []byte("---\ntitle: Home\n---\n")},
	}

	result, err := builder.Build(context.Background(), broken, sitebuild.Options{})
	if err != nil {
		t.Fatalf("Build() error = %v, want a failed build rather than an error", err)
	}
	defer result.Release()

	if result.Succeeded() {
		t.Fatal("Succeeded() = true, want the broken site to fail")
	}
	if !strings.Contains(strings.ToLower(result.Log), "error") {
		t.Errorf("Log does not say what went wrong:\n%s", result.Log)
	}
}

func TestABuildThatTakesTooLongIsEnded(t *testing.T) {
	sandbox, err := runner.NewSandbox(runner.Config{WorkRoot: t.TempDir()})
	if err != nil {
		if errors.Is(err, runner.ErrNoIsolation) {
			t.Skipf("no isolation available here: %v", err)
		}
		t.Fatal(err)
	}
	builder, err := sitebuild.New(sandbox, hugoBin(t), sitebuild.Limits{Timeout: 50 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}

	result, err := builder.Build(context.Background(), os.DirFS(fixture), sitebuild.Options{})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	defer result.Release()

	if result.Reached != runner.LimitTime {
		t.Errorf("Reached = %q, want %q", result.Reached, runner.LimitTime)
	}
}

func TestNewRefusesAnIncompleteBuilder(t *testing.T) {
	if _, err := sitebuild.New(nil, "hugo", sitebuild.Limits{}); err == nil {
		t.Error("New() with no runner error = nil, want an error")
	}
	sandbox := &stubRunner{}
	if _, err := sitebuild.New(sandbox, "", sitebuild.Limits{}); !errors.Is(err, sitebuild.ErrNoHugo) {
		t.Errorf("New() with no Hugo error = %v, want ErrNoHugo", err)
	}
}

func TestBuildRefusesWithNoSite(t *testing.T) {
	builder, err := sitebuild.New(&stubRunner{}, "hugo", sitebuild.Limits{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := builder.Build(context.Background(), nil, sitebuild.Options{}); err == nil {
		t.Error("Build() with no site error = nil, want an error")
	}
}

func TestTheLimitsHaveDefaults(t *testing.T) {
	stub := &stubRunner{}
	builder, err := sitebuild.New(stub, "hugo", sitebuild.Limits{})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := builder.Build(context.Background(), fstest.MapFS{}, sitebuild.Options{}); err != nil {
		t.Fatal(err)
	}

	if stub.task.Timeout != sitebuild.DefaultTimeout {
		t.Errorf("Timeout = %v, want the default", stub.task.Timeout)
	}
	if stub.task.MemoryBytes != sitebuild.DefaultMemoryBytes {
		t.Errorf("MemoryBytes = %d, want the default", stub.task.MemoryBytes)
	}
	if err := runner.Validate(stub.task); err != nil {
		t.Errorf("the task a build makes is not valid: %v", err)
	}
}

func TestDraftsAreOffUnlessAskedFor(t *testing.T) {
	stub := &stubRunner{}
	builder, err := sitebuild.New(stub, "hugo", sitebuild.Limits{})
	if err != nil {
		t.Fatal(err)
	}

	builder.Build(context.Background(), fstest.MapFS{}, sitebuild.Options{})
	if contains(stub.task.Arguments, "--buildDrafts") {
		t.Error("drafts were built without being asked for")
	}

	builder.Build(context.Background(), fstest.MapFS{}, sitebuild.Options{Drafts: true})
	if !contains(stub.task.Arguments, "--buildDrafts") {
		t.Error("drafts were asked for and not built")
	}
}

func TestABuildAsksForNoNetwork(t *testing.T) {
	stub := &stubRunner{}
	builder, err := sitebuild.New(stub, "hugo", sitebuild.Limits{})
	if err != nil {
		t.Fatal(err)
	}

	builder.Build(context.Background(), fstest.MapFS{}, sitebuild.Options{})

	if stub.task.Network {
		t.Error("a site build asked for the network")
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

type stubRunner struct {
	task runner.Task
}

func (s *stubRunner) Run(ctx context.Context, task runner.Task) (*runner.Result, error) {
	s.task = task
	return &runner.Result{}, nil
}
