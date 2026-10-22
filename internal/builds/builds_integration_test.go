//go:build integration

package builds_test

import (
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/auth"
	"github.com/quiqr/chatcms/internal/builds"
	"github.com/quiqr/chatcms/internal/store"
	"github.com/quiqr/chatcms/internal/tenancy"
	"github.com/quiqr/chatcms/internal/testsupport"
)

func records(t *testing.T) (*builds.Records, uuid.UUID) {
	store, projects := recordsWithProjects(t)
	return store, projects(t)
}

func recordsWithProjects(t *testing.T) (*builds.Records, func(*testing.T) uuid.UUID) {
	t.Helper()

	db, err := store.Open(t.Context(), testsupport.Postgres(t))
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	t.Cleanup(db.Close)
	if err := db.Migrate(t.Context()); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	tenants := tenancy.NewStore(db)
	people := auth.NewService(db, slog.New(slog.NewJSONHandler(io.Discard, nil)), "https://chatcms.example").People()
	owner, err := people.Create(t.Context(), "builds-"+uuid.NewString()+"@example.org", "The Owner", "a long enough passphrase")
	if err != nil {
		t.Fatal(err)
	}
	organisation, err := tenants.CreateOrganisation(t.Context(), "builds-"+uuid.NewString()[:8], "Builds", owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	return builds.NewRecords(db.Queries), func(t *testing.T) uuid.UUID {
		t.Helper()

		project, err := tenants.CreateProject(t.Context(), organisation.ID, "site-"+uuid.NewString()[:8], "A Site", "")
		if err != nil {
			t.Fatal(err)
		}
		return project.ID
	}
}

func TestABuildIsRecordedFromStartToFinish(t *testing.T) {
	store, project := records(t)

	started, err := store.Start(t.Context(), project, "the editor pressed publish")
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if started.State != builds.Running {
		t.Errorf("State = %q, want %q", started.State, builds.Running)
	}
	if started.FinishedAt != nil {
		t.Error("a build that just started already has a finish")
	}

	code := int32(0)
	finished, err := store.Finish(t.Context(), started.ID, builds.Succeeded, &code, "built 5 pages")
	if err != nil {
		t.Fatalf("Finish() error = %v", err)
	}
	if finished.State != builds.Succeeded || finished.FinishedAt == nil {
		t.Errorf("finished = %+v, want a finished build", finished)
	}
	if finished.Duration() < 0 {
		t.Errorf("Duration() = %v, want it to be sound", finished.Duration())
	}

	read, err := store.Get(t.Context(), started.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if read.Log != "built 5 pages" {
		t.Errorf("Log = %q, want what was recorded", read.Log)
	}
	if read.Reason != "the editor pressed publish" {
		t.Errorf("Reason = %q, want what was recorded", read.Reason)
	}
}

func TestABuildCannotBeFinishedTwice(t *testing.T) {
	store, project := records(t)

	started, err := store.Start(t.Context(), project, "a publish")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Finish(t.Context(), started.ID, builds.Succeeded, nil, ""); err != nil {
		t.Fatal(err)
	}

	_, err = store.Finish(t.Context(), started.ID, builds.Failed, nil, "a second opinion")
	if !errors.Is(err, builds.ErrNoSuchBuild) {
		t.Fatalf("a second Finish() error = %v, want ErrNoSuchBuild", err)
	}

	read, err := store.Get(t.Context(), started.ID)
	if err != nil {
		t.Fatal(err)
	}
	if read.State != builds.Succeeded {
		t.Errorf("State = %q, want the first finish to stand", read.State)
	}
}

func TestEveryStateTheRecordAcceptsIsOneTheCodeUses(t *testing.T) {
	store, project := records(t)

	for _, state := range []builds.State{
		builds.Succeeded, builds.Failed, builds.OutOfTime, builds.OutOfMemory, builds.Interrupted,
	} {
		started, err := store.Start(t.Context(), project, "a publish")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := store.Finish(t.Context(), started.ID, state, nil, ""); err != nil {
			t.Errorf("Finish() with %q error = %v, want the database to accept it", state, err)
		}
	}
}

func TestAStateTheCodeDoesNotUseIsRefusedByTheDatabase(t *testing.T) {
	store, project := records(t)

	started, err := store.Start(t.Context(), project, "a publish")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.Finish(t.Context(), started.ID, builds.State("mostly-fine"), nil, ""); err == nil {
		t.Fatal("Finish() with an unknown state error = nil, want the database to refuse it")
	}
}

func TestBuildsAreListedRecentFirstAndPerProject(t *testing.T) {
	store, projects := recordsWithProjects(t)
	project := projects(t)
	other := projects(t)

	for _, reason := range []string{"first", "second", "third"} {
		started, err := store.Start(t.Context(), project, reason)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := store.Finish(t.Context(), started.ID, builds.Succeeded, nil, ""); err != nil {
			t.Fatal(err)
		}
	}
	elsewhere, err := store.Start(t.Context(), other, "somebody else's build")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Finish(t.Context(), elsewhere.ID, builds.Succeeded, nil, ""); err != nil {
		t.Fatal(err)
	}

	listed, err := store.List(t.Context(), project, 10)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(listed) != 3 {
		t.Fatalf("listed %d builds, want this project's three", len(listed))
	}
	if listed[0].Reason != "third" {
		t.Errorf("first listed = %q, want the most recent", listed[0].Reason)
	}
	for _, build := range listed {
		if build.ProjectID != project {
			t.Errorf("listed a build from project %s", build.ProjectID)
		}
	}

	shorter, err := store.List(t.Context(), project, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(shorter) != 2 {
		t.Errorf("listed %d builds with a limit of two", len(shorter))
	}
}

func TestABuildLeftRunningIsMarkedInterrupted(t *testing.T) {
	store, project := records(t)

	started, err := store.Start(t.Context(), project, "a publish that never came back")
	if err != nil {
		t.Fatal(err)
	}

	count, err := store.InterruptLeftovers(t.Context())
	if err != nil {
		t.Fatalf("InterruptLeftovers() error = %v", err)
	}
	if count < 1 {
		t.Errorf("InterruptLeftovers() = %d, want at least the one left running", count)
	}

	read, err := store.Get(t.Context(), started.ID)
	if err != nil {
		t.Fatal(err)
	}
	if read.State != builds.Interrupted {
		t.Fatalf("State = %q, want %q", read.State, builds.Interrupted)
	}
	if read.FinishedAt == nil {
		t.Error("an interrupted build has no finish")
	}
	if !strings.Contains(read.Log, "still running when the server stopped") {
		t.Errorf("Log = %q, want it to say what happened", read.Log)
	}

	again, err := store.InterruptLeftovers(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if again != 0 {
		t.Errorf("a second InterruptLeftovers() = %d, want nothing left to do", again)
	}
}

func TestAnUnknownBuildIsReportedAsSuch(t *testing.T) {
	store, _ := records(t)

	if _, err := store.Get(t.Context(), uuid.New()); !errors.Is(err, builds.ErrNoSuchBuild) {
		t.Errorf("Get() on an unknown build error = %v, want ErrNoSuchBuild", err)
	}
}
