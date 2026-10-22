package publish_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/policy"
	"github.com/quiqr/chatcms/internal/publish"
	"github.com/quiqr/chatcms/internal/workflows"
)

type sources struct {
	site     *site
	settings workflows.Settings
	baseURL  string
	closed   int
	err      error
}

func (s *sources) Site(ctx context.Context, projectID uuid.UUID) (publish.Site, func() error, error) {
	if s.err != nil {
		return nil, nil, s.err
	}
	return s.site, func() error { s.closed++; return nil }, nil
}

func (s *sources) Settings(ctx context.Context, projectID uuid.UUID) (workflows.Settings, error) {
	return s.settings, nil
}

func (s *sources) BaseURL(ctx context.Context, projectID uuid.UUID) (string, error) {
	return s.baseURL, nil
}

func (s *sources) Resource(ctx context.Context, projectID uuid.UUID) (policy.Resource, error) {
	return policy.Project(organisation, projectID.String()), nil
}

const organisation = "11111111-1111-1111-1111-111111111111"

type engine struct {
	outcome  *workflows.Outcome
	err      error
	workflow workflows.Workflow
	run      workflows.Run
}

func (e *engine) Run(ctx context.Context, workflow workflows.Workflow, run workflows.Run) (*workflows.Outcome, error) {
	e.workflow, e.run = workflow, run
	if e.err != nil {
		return nil, e.err
	}
	if e.outcome == nil {
		return &workflows.Outcome{Trigger: workflows.OnPublish}, nil
	}
	return e.outcome, nil
}

type records struct {
	started  []publish.Record
	finished []publish.Record
	last     publish.Record
	lastErr  error
	startErr error
}

func (r *records) Start(ctx context.Context, projectID uuid.UUID, personID *uuid.UUID, note string, summary publish.Summary) (publish.Record, error) {
	if r.startErr != nil {
		return publish.Record{}, r.startErr
	}
	record := publish.Record{
		ID: uuid.New(), ProjectID: projectID, PersonID: personID, Note: note,
		State: publish.Running, ChangeCount: summary.Count(), Summary: summary.Changes,
		StartedAt: time.Now(),
	}
	r.started = append(r.started, record)
	return record, nil
}

func (r *records) Finish(ctx context.Context, id uuid.UUID, state publish.State, steps []workflows.StepOutcome, trouble, detail string) (publish.Record, error) {
	now := time.Now()
	record := publish.Record{ID: id, State: state, Steps: steps, Trouble: trouble, Detail: detail, FinishedAt: &now}
	for _, started := range r.started {
		if started.ID == id {
			record.ProjectID = started.ProjectID
			record.ChangeCount = started.ChangeCount
		}
	}
	r.finished = append(r.finished, record)
	return record, nil
}

func (r *records) Get(ctx context.Context, id uuid.UUID) (publish.Record, error) {
	for _, record := range append(r.finished, r.started...) {
		if record.ID == id {
			return record, nil
		}
	}
	return publish.Record{}, publish.ErrNoSuchPublish
}

func (r *records) List(ctx context.Context, projectID uuid.UUID, limit int32) ([]publish.Record, error) {
	return r.finished, nil
}

func (r *records) LastSucceeded(ctx context.Context, projectID uuid.UUID) (publish.Record, error) {
	if r.lastErr != nil {
		return publish.Record{}, r.lastErr
	}
	return r.last, nil
}

func (r *records) InterruptLeftovers(ctx context.Context) (int, error) { return 1, nil }

func anEditor(project uuid.UUID) policy.Actor {
	return policy.Actor{
		PersonID: uuid.NewString(),
		Projects: map[string]policy.Role{project.String(): policy.RoleEditor},
	}
}

func aViewer(project uuid.UUID) policy.Actor {
	return policy.Actor{
		PersonID: uuid.NewString(),
		Projects: map[string]policy.Role{project.String(): policy.RoleViewer},
	}
}

func anImplementer(project uuid.UUID) policy.Actor {
	return policy.Actor{
		PersonID: uuid.NewString(),
		Projects: map[string]policy.Role{project.String(): policy.RoleImplementer},
	}
}

func service(t *testing.T, src *sources, eng *engine, rec *records) *publish.Service {
	t.Helper()

	made, err := publish.NewService(src, eng, rec)
	if err != nil {
		t.Fatal(err)
	}
	return made
}

func TestReviewShowsWhatChanged(t *testing.T) {
	project := uuid.New()
	src := &sources{site: anEditedHomePage()}
	rec := &records{lastErr: publish.ErrNoSuchPublish}

	summary, err := service(t, src, &engine{}, rec).Review(context.Background(), anEditor(project), project)
	if err != nil {
		t.Fatalf("Review() error = %v", err)
	}

	if summary.Count() != 1 {
		t.Errorf("Review() = %v, want one change", summary.Changes)
	}
	if src.closed != 1 {
		t.Errorf("the site was released %d times, want once", src.closed)
	}
}

func TestReviewSaysWhenTheSiteWasLastPublished(t *testing.T) {
	project := uuid.New()
	when := time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC)
	rec := &records{last: publish.Record{State: publish.Succeeded, FinishedAt: &when}}

	summary, err := service(t, &sources{site: anEditedHomePage()}, &engine{}, rec).
		Review(context.Background(), anEditor(project), project)
	if err != nil {
		t.Fatal(err)
	}

	if summary.LastPublished == nil || !summary.LastPublished.Equal(when) {
		t.Errorf("LastPublished = %v, want %v", summary.LastPublished, when)
	}
}

func TestAViewerMayLookButNotPublish(t *testing.T) {
	project := uuid.New()
	made := service(t, &sources{site: anEditedHomePage()}, &engine{}, &records{lastErr: publish.ErrNoSuchPublish})

	if _, err := made.Review(context.Background(), aViewer(project), project); err != nil {
		t.Errorf("Review() as a viewer error = %v, want it allowed", err)
	}
	if _, err := made.Run(context.Background(), aViewer(project), project, ""); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("Run() as a viewer error = %v, want ErrNotAllowed", err)
	}
}

func TestSomebodyWithNoPartInTheProjectIsToldItIsNotThere(t *testing.T) {
	project := uuid.New()
	made := service(t, &sources{site: anEditedHomePage()}, &engine{}, &records{})

	stranger := policy.Actor{PersonID: uuid.NewString()}
	for name, call := range map[string]func() error{
		"review": func() error { _, err := made.Review(context.Background(), stranger, project); return err },
		"run":    func() error { _, err := made.Run(context.Background(), stranger, project, ""); return err },
		"history": func() error {
			_, err := made.History(context.Background(), stranger, project, 0)
			return err
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := call(); !errors.Is(err, httpapi.ErrMissing) {
				t.Errorf("error = %v, want ErrMissing", err)
			}
		})
	}
}

func TestAPublishThatWorkedIsRecorded(t *testing.T) {
	project := uuid.New()
	rec := &records{lastErr: publish.ErrNoSuchPublish}
	eng := &engine{}

	record, err := service(t, &sources{site: anEditedHomePage(), baseURL: "https://the-customer.example/"}, eng, rec).
		Run(context.Background(), anEditor(project), project, "Spring prices")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if record.State != publish.Succeeded {
		t.Errorf("State = %q, want %q", record.State, publish.Succeeded)
	}
	if len(rec.started) != 1 || rec.started[0].Note != "Spring prices" {
		t.Errorf("the note was not recorded: %v", rec.started)
	}
	if rec.started[0].ChangeCount != 1 {
		t.Errorf("ChangeCount = %d, want what the review showed", rec.started[0].ChangeCount)
	}
	if eng.run.BaseURL != "https://the-customer.example/" {
		t.Errorf("the workflow ran with BaseURL %q", eng.run.BaseURL)
	}
	if eng.run.ProjectID != project {
		t.Error("the workflow ran against another project")
	}
}

func TestAFailedPublishCarriesTheSentenceAndTheDetailApart(t *testing.T) {
	project := uuid.New()
	rec := &records{lastErr: publish.ErrNoSuchPublish}
	eng := &engine{outcome: &workflows.Outcome{
		Trigger: workflows.OnPublish,
		Steps: []workflows.StepOutcome{{
			Title: "Checking pages and links", Action: workflows.Check,
			State: workflows.Failed, Detail: "fatal: three links are broken",
		}},
		Trouble: &workflows.Trouble{
			Because: "A check on this site found something that needs fixing first.",
		},
	}}

	record, err := service(t, &sources{site: anEditedHomePage()}, eng, rec).
		Run(context.Background(), anEditor(project), project, "")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if record.State != publish.Failed {
		t.Errorf("State = %q, want %q", record.State, publish.Failed)
	}
	if !strings.Contains(record.Trouble, "Nothing changed on your live site") {
		t.Errorf("Trouble = %q, want the reassurance", record.Trouble)
	}
	if strings.Contains(record.Trouble, "fatal") {
		t.Error("the raw output ended up in what an editor reads")
	}
	if !strings.Contains(record.Detail, "three links are broken") {
		t.Errorf("Detail = %q, want the raw output", record.Detail)
	}
}

func TestAPublishThatCouldNotStartIsStillRecorded(t *testing.T) {
	project := uuid.New()
	rec := &records{lastErr: publish.ErrNoSuchPublish}
	eng := &engine{err: errors.New("the engine fell over")}

	record, err := service(t, &sources{site: anEditedHomePage()}, eng, rec).
		Run(context.Background(), anEditor(project), project, "")
	if err != nil {
		t.Fatalf("Run() error = %v, want the failure recorded", err)
	}

	if record.State != publish.Failed {
		t.Errorf("State = %q, want %q", record.State, publish.Failed)
	}
	if !strings.Contains(record.Trouble, "edits are safe") {
		t.Errorf("Trouble = %q", record.Trouble)
	}
}

func TestTheSitesOwnWorkflowIsUsed(t *testing.T) {
	project := uuid.New()
	declared := anEditedHomePage()
	declared.now["quiqr/workflows/publish.yaml"] = &fstest.MapFile{
		Data: []byte("on: publish\nsteps:\n  - action: check\n    run: ./check.sh\n  - action: publish\n"),
	}
	eng := &engine{}

	if _, err := service(t, &sources{site: declared}, eng, &records{lastErr: publish.ErrNoSuchPublish}).
		Run(context.Background(), anEditor(project), project, ""); err != nil {
		t.Fatal(err)
	}

	if !eng.workflow.Does(workflows.Check) {
		t.Errorf("the workflow that ran was %v, want the one the site declared", eng.workflow.Steps)
	}
}

func TestASiteWithNoWorkflowGetsTheDefaultForItsSettings(t *testing.T) {
	project := uuid.New()
	eng := &engine{}
	src := &sources{site: anEditedHomePage(), settings: workflows.Settings{ChatCMSBuilds: true}}

	if _, err := service(t, src, eng, &records{lastErr: publish.ErrNoSuchPublish}).
		Run(context.Background(), anEditor(project), project, ""); err != nil {
		t.Fatal(err)
	}

	if !eng.workflow.Does(workflows.Build) || !eng.workflow.Does(workflows.Publish) {
		t.Errorf("the workflow that ran was %v, want build then publish", eng.workflow.Steps)
	}
}

func TestOnlySomebodyAllowedTechnicalDetailSeesIt(t *testing.T) {
	project := uuid.New()
	rec := &records{lastErr: publish.ErrNoSuchPublish}
	eng := &engine{outcome: &workflows.Outcome{
		Trigger: workflows.OnPublish,
		Steps:   []workflows.StepOutcome{{Title: "Building the site", State: workflows.Failed, Detail: "template error at line 7"}},
		Trouble: &workflows.Trouble{Because: "The site could not be built."},
	}}
	made := service(t, &sources{site: anEditedHomePage()}, eng, rec)

	record, err := made.Run(context.Background(), anEditor(project), project, "")
	if err != nil {
		t.Fatal(err)
	}

	forTheEditor, err := made.Get(context.Background(), anEditor(project), project, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	if forTheEditor.Detail != "" {
		t.Errorf("an editor was given the raw output: %q", forTheEditor.Detail)
	}
	if forTheEditor.Trouble == "" {
		t.Error("an editor was given nothing to read")
	}

	forTheImplementer, err := made.Get(context.Background(), anImplementer(project), project, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(forTheImplementer.Detail, "template error") {
		t.Errorf("an implementer was not given the raw output: %q", forTheImplementer.Detail)
	}
}

func TestAPublishFromAnotherProjectIsNotFound(t *testing.T) {
	project := uuid.New()
	rec := &records{lastErr: publish.ErrNoSuchPublish}
	made := service(t, &sources{site: anEditedHomePage()}, &engine{}, rec)

	record, err := made.Run(context.Background(), anEditor(project), project, "")
	if err != nil {
		t.Fatal(err)
	}

	elsewhere := uuid.New()
	if _, err := made.Get(context.Background(), anEditor(elsewhere), elsewhere, record.ID); !errors.Is(err, publish.ErrNoSuchPublish) {
		t.Errorf("Get() across projects error = %v, want ErrNoSuchPublish", err)
	}
}

func TestHistoryHidesTheRawOutputFromAnEditor(t *testing.T) {
	project := uuid.New()
	rec := &records{
		lastErr: publish.ErrNoSuchPublish,
		finished: []publish.Record{{
			ID: uuid.New(), ProjectID: project, State: publish.Failed,
			Trouble: "We couldn't publish this time.", Detail: "the raw output",
		}},
	}
	made := service(t, &sources{site: anEditedHomePage()}, &engine{}, rec)

	forTheEditor, err := made.History(context.Background(), anEditor(project), project, 0)
	if err != nil {
		t.Fatal(err)
	}
	if forTheEditor[0].Detail != "" {
		t.Error("an editor was given the raw output in the history")
	}

	rec.finished[0].Detail = "the raw output"
	forTheImplementer, err := made.History(context.Background(), anImplementer(project), project, 0)
	if err != nil {
		t.Fatal(err)
	}
	if forTheImplementer[0].Detail == "" {
		t.Error("an implementer was not given the raw output in the history")
	}
}

func TestNewServiceRefusesAMissingPart(t *testing.T) {
	cases := map[string]struct {
		sources publish.Sources
		engine  publish.Engine
		records publish.Recorder
	}{
		"no sources": {nil, &engine{}, &records{}},
		"no engine":  {&sources{}, nil, &records{}},
		"no records": {&sources{}, &engine{}, nil},
	}

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := publish.NewService(test.sources, test.engine, test.records); !errors.Is(err, publish.ErrIncomplete) {
				t.Errorf("NewService() error = %v, want ErrIncomplete", err)
			}
		})
	}
}

func TestTheSiteIsReleasedEvenWhenAPublishFails(t *testing.T) {
	project := uuid.New()
	src := &sources{site: anEditedHomePage()}
	eng := &engine{err: errors.New("the engine fell over")}

	if _, err := service(t, src, eng, &records{lastErr: publish.ErrNoSuchPublish}).
		Run(context.Background(), anEditor(project), project, ""); err != nil {
		t.Fatal(err)
	}

	if src.closed != 1 {
		t.Errorf("the site was released %d times, want once", src.closed)
	}
}
