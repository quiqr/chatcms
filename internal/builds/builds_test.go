package builds_test

import (
	"context"
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/builds"
	"github.com/quiqr/chatcms/internal/runner"
	"github.com/quiqr/chatcms/internal/sitebuild"
)

type remembered struct {
	started     []builds.Build
	finished    []builds.Build
	startErr    error
	finishErr   error
	interrupted int
}

func (r *remembered) Start(ctx context.Context, projectID uuid.UUID, reason string) (builds.Build, error) {
	if r.startErr != nil {
		return builds.Build{}, r.startErr
	}
	build := builds.Build{
		ID:        uuid.New(),
		ProjectID: projectID,
		Reason:    reason,
		State:     builds.Running,
		StartedAt: time.Now(),
	}
	r.started = append(r.started, build)
	return build, nil
}

func (r *remembered) Finish(ctx context.Context, id uuid.UUID, state builds.State, exitCode *int32, log string) (builds.Build, error) {
	if r.finishErr != nil {
		return builds.Build{}, r.finishErr
	}
	now := time.Now()
	build := builds.Build{ID: id, State: state, ExitCode: exitCode, Log: log, FinishedAt: &now}
	r.finished = append(r.finished, build)
	return build, nil
}

func (r *remembered) Get(ctx context.Context, id uuid.UUID) (builds.Build, error) {
	for _, build := range append(r.finished, r.started...) {
		if build.ID == id {
			return build, nil
		}
	}
	return builds.Build{}, builds.ErrNoSuchBuild
}

func (r *remembered) List(ctx context.Context, projectID uuid.UUID, limit int32) ([]builds.Build, error) {
	out := []builds.Build{}
	for _, build := range r.started {
		if build.ProjectID == projectID && int32(len(out)) < limit {
			out = append(out, build)
		}
	}
	return out, nil
}

func (r *remembered) InterruptLeftovers(ctx context.Context) (int, error) {
	return r.interrupted, nil
}

type fixedBuilder struct {
	result *runner.Result
	err    error
	seen   sitebuild.Options
}

func (b *fixedBuilder) Build(ctx context.Context, site fs.FS, options sitebuild.Options) (*runner.Result, error) {
	b.seen = options
	return b.result, b.err
}

func aSite(ctx context.Context, projectID uuid.UUID) (fs.FS, error) {
	return fstest.MapFS{"hugo.toml": {Data: []byte("title = \"a site\"\n")}}, nil
}

func newService(t *testing.T, records *remembered, builder *fixedBuilder) *builds.Service {
	t.Helper()

	service, err := builds.NewService(records, builder, aSite)
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func TestABuildThatWorkedIsRecordedWithItsOutput(t *testing.T) {
	records := &remembered{}
	service := newService(t, records, &fixedBuilder{
		result: &runner.Result{ExitCode: 0, Log: "built 5 pages"},
	})

	outcome, err := service.Run(context.Background(), uuid.New(), "a publish", sitebuild.Options{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	defer outcome.Release()

	if outcome.Build.State != builds.Succeeded {
		t.Errorf("State = %q, want %q", outcome.Build.State, builds.Succeeded)
	}
	if len(records.started) != 1 || len(records.finished) != 1 {
		t.Fatalf("recorded %d starts and %d finishes, want one of each", len(records.started), len(records.finished))
	}
	if records.finished[0].Log != "built 5 pages" {
		t.Errorf("Log = %q, want what the build said", records.finished[0].Log)
	}
}

func TestEveryWayABuildCanEndIsRecordedAsItsOwnState(t *testing.T) {
	cases := map[string]struct {
		result *runner.Result
		want   builds.State
	}{
		"worked":        {&runner.Result{ExitCode: 0}, builds.Succeeded},
		"failed":        {&runner.Result{ExitCode: 1}, builds.Failed},
		"out of time":   {&runner.Result{ExitCode: -1, Reached: runner.LimitTime}, builds.OutOfTime},
		"out of memory": {&runner.Result{ExitCode: 2, Reached: runner.LimitMemory}, builds.OutOfMemory},
	}

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			records := &remembered{}
			service := newService(t, records, &fixedBuilder{result: test.result})

			outcome, err := service.Run(context.Background(), uuid.New(), "a publish", sitebuild.Options{})
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}
			defer outcome.Release()

			if outcome.Build.State != test.want {
				t.Errorf("State = %q, want %q", outcome.Build.State, test.want)
			}
		})
	}
}

func TestOnlyASucceededBuildHandsBackItsOutput(t *testing.T) {
	released := false
	records := &remembered{}
	service := newService(t, records, &fixedBuilder{
		result: resultWithOutputs(&released, 1),
	})

	outcome, err := service.Run(context.Background(), uuid.New(), "a publish", sitebuild.Options{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if outcome.Outputs != nil {
		t.Error("a failed build handed back its output")
	}
	if !released {
		t.Error("a failed build did not release its workspace")
	}
}

func resultWithOutputs(released *bool, exitCode int) *runner.Result {
	return &runner.Result{
		ExitCode:  exitCode,
		Outputs:   fstest.MapFS{"public/index.html": {Data: []byte("page")}},
		OnRelease: func() error { *released = true; return nil },
	}
}

func TestABuildThatCouldNotStartIsStillRecorded(t *testing.T) {
	records := &remembered{}
	service := newService(t, records, &fixedBuilder{err: runner.ErrNoIsolation})

	outcome, err := service.Run(context.Background(), uuid.New(), "a publish", sitebuild.Options{})
	if err != nil {
		t.Fatalf("Run() error = %v, want the failure to be recorded rather than returned", err)
	}

	if outcome.Build.State != builds.Failed {
		t.Errorf("State = %q, want %q", outcome.Build.State, builds.Failed)
	}
	if !strings.Contains(records.finished[0].Log, "could not be started") {
		t.Errorf("Log = %q, want it to say the build never ran", records.finished[0].Log)
	}
}

func TestASiteThatCouldNotBeReadIsRecordedAsAFailure(t *testing.T) {
	records := &remembered{}
	service, err := builds.NewService(records, &fixedBuilder{result: &runner.Result{}},
		func(ctx context.Context, projectID uuid.UUID) (fs.FS, error) {
			return nil, errors.New("the working copy is not there")
		})
	if err != nil {
		t.Fatal(err)
	}

	outcome, err := service.Run(context.Background(), uuid.New(), "a publish", sitebuild.Options{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if outcome.Build.State != builds.Failed {
		t.Errorf("State = %q, want %q", outcome.Build.State, builds.Failed)
	}
	if !strings.Contains(records.finished[0].Log, "working copy is not there") {
		t.Errorf("Log = %q, want it to carry why", records.finished[0].Log)
	}
}

func TestAVeryLongLogIsShortenedBeforeItIsStored(t *testing.T) {
	records := &remembered{}
	service := newService(t, records, &fixedBuilder{
		result: &runner.Result{ExitCode: 1, Log: strings.Repeat("noise\n", 200000) + "THE REAL ERROR"},
	})

	outcome, err := service.Run(context.Background(), uuid.New(), "a publish", sitebuild.Options{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	defer outcome.Release()

	stored := records.finished[0].Log
	if len(stored) > builds.MaxLogBytes+200 {
		t.Errorf("the stored log is %d bytes, want it bounded to %d", len(stored), builds.MaxLogBytes)
	}
	if !strings.HasSuffix(stored, "THE REAL ERROR") {
		t.Error("the stored log does not end with the error")
	}
	if !strings.Contains(stored, "dropped") {
		t.Error("the stored log does not say output was dropped")
	}
}

func TestNewServiceRefusesAMissingPart(t *testing.T) {
	cases := map[string]struct {
		records builds.Recorder
		builder builds.SiteBuilder
		sites   builds.Siteloader
	}{
		"no records": {nil, &fixedBuilder{}, aSite},
		"no builder": {&remembered{}, nil, aSite},
		"no sites":   {&remembered{}, &fixedBuilder{}, nil},
	}

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := builds.NewService(test.records, test.builder, test.sites); !errors.Is(err, builds.ErrIncomplete) {
				t.Errorf("NewService() error = %v, want ErrIncomplete", err)
			}
		})
	}
}

func TestListingHasADefaultLength(t *testing.T) {
	records := &remembered{}
	service := newService(t, records, &fixedBuilder{result: &runner.Result{}})
	project := uuid.New()
	for i := 0; i < 3; i++ {
		outcome, err := service.Run(context.Background(), project, "a publish", sitebuild.Options{})
		if err != nil {
			t.Fatal(err)
		}
		outcome.Release()
	}

	listed, err := service.List(context.Background(), project, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 3 {
		t.Errorf("listed %d builds, want 3", len(listed))
	}
}

func TestDurationIsZeroWhileABuildIsStillRunning(t *testing.T) {
	running := builds.Build{State: builds.Running, StartedAt: time.Now()}
	if running.Duration() != 0 {
		t.Errorf("Duration() = %v, want 0 while it is running", running.Duration())
	}
	if running.State.Finished() {
		t.Error("Finished() = true for a running build")
	}

	finish := running.StartedAt.Add(90 * time.Second)
	done := builds.Build{State: builds.Succeeded, StartedAt: running.StartedAt, FinishedAt: &finish}
	if done.Duration() != 90*time.Second {
		t.Errorf("Duration() = %v, want 90s", done.Duration())
	}
	if !done.State.Finished() {
		t.Error("Finished() = false for a finished build")
	}
}

func TestInterruptLeftoversIsPassedOn(t *testing.T) {
	records := &remembered{interrupted: 2}
	service := newService(t, records, &fixedBuilder{})

	count, err := service.InterruptLeftovers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("InterruptLeftovers() = %d, want 2", count)
	}
}

func TestGetReachesTheRecords(t *testing.T) {
	records := &remembered{}
	service := newService(t, records, &fixedBuilder{result: &runner.Result{}})
	project := uuid.New()
	outcome, err := service.Run(context.Background(), project, "a publish", sitebuild.Options{})
	if err != nil {
		t.Fatal(err)
	}
	outcome.Release()

	if _, err := service.Get(context.Background(), outcome.Build.ID); err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if _, err := service.Get(context.Background(), uuid.New()); !errors.Is(err, builds.ErrNoSuchBuild) {
		t.Errorf("Get() on an unknown build error = %v, want ErrNoSuchBuild", err)
	}
}

func TestOutcomeReleaseIsSafeOnNothing(t *testing.T) {
	var absent *builds.Outcome
	if err := absent.Release(); err != nil {
		t.Errorf("Release() on nothing = %v, want nil", err)
	}
}
