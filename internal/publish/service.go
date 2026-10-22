package publish

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/model"
	"github.com/quiqr/chatcms/internal/policy"
	"github.com/quiqr/chatcms/internal/workflows"
)

// Site is everything a publish needs from one project's working copy.
type Site interface {
	Reader
	Model() (*model.Model, error)
}

// Sources opens a project's site and reads its settings. The caller supplies
// it, so this package does not reach into the database or the git engine.
type Sources interface {
	Site(ctx context.Context, projectID uuid.UUID) (Site, func() error, error)
	Settings(ctx context.Context, projectID uuid.UUID) (workflows.Settings, error)
	BaseURL(ctx context.Context, projectID uuid.UUID) (string, error)
	// Resource says what a project is, in the terms the policy function
	// asks about, so this package decides nothing about roles itself.
	Resource(ctx context.Context, projectID uuid.UUID) (policy.Resource, error)
}

type Engine interface {
	Run(ctx context.Context, workflow workflows.Workflow, run workflows.Run) (*workflows.Outcome, error)
}

type Recorder interface {
	Start(ctx context.Context, projectID uuid.UUID, personID *uuid.UUID, note string, summary Summary) (Record, error)
	Finish(ctx context.Context, id uuid.UUID, state State, steps []workflows.StepOutcome, trouble, detail string) (Record, error)
	Get(ctx context.Context, id uuid.UUID) (Record, error)
	List(ctx context.Context, projectID uuid.UUID, limit int32) ([]Record, error)
	LastSucceeded(ctx context.Context, projectID uuid.UUID) (Record, error)
	InterruptLeftovers(ctx context.Context) (int, error)
}

var ErrIncomplete = errors.New("publish: the service is missing a part it needs")

type Service struct {
	sources Sources
	engine  Engine
	records Recorder
}

func NewService(sources Sources, engine Engine, records Recorder) (*Service, error) {
	if sources == nil || engine == nil || records == nil {
		return nil, ErrIncomplete
	}
	return &Service{sources: sources, engine: engine, records: records}, nil
}

// Review is what a person sees before deciding to publish.
func (s *Service) Review(ctx context.Context, actor policy.Actor, projectID uuid.UUID) (Summary, error) {
	if err := s.allowed(ctx, actor, policy.ActionReadContent, projectID); err != nil {
		return Summary{}, err
	}

	site, done, err := s.sources.Site(ctx, projectID)
	if err != nil {
		return Summary{}, err
	}
	defer done()

	summary, err := Describe(ctx, site, modelOf(site))
	if err != nil {
		return Summary{}, err
	}
	if last, err := s.records.LastSucceeded(ctx, projectID); err == nil && last.FinishedAt != nil {
		summary.LastPublished = last.FinishedAt
	}
	return summary, nil
}

func modelOf(site Site) *model.Model {
	m, err := site.Model()
	if err != nil {
		return nil
	}
	return m
}

const DefaultListLength = 50

// Run publishes a project: work out what changed, record that it started, run
// the site's publish workflow, then record how it went.
func (s *Service) Run(ctx context.Context, actor policy.Actor, projectID uuid.UUID, note string) (Record, error) {
	if err := s.allowed(ctx, actor, policy.ActionPublish, projectID); err != nil {
		return Record{}, err
	}

	site, done, err := s.sources.Site(ctx, projectID)
	if err != nil {
		return Record{}, err
	}
	defer done()

	m := modelOf(site)
	summary, err := Describe(ctx, site, m)
	if err != nil {
		return Record{}, err
	}

	settings, err := s.sources.Settings(ctx, projectID)
	if err != nil {
		return Record{}, err
	}
	baseURL, err := s.sources.BaseURL(ctx, projectID)
	if err != nil {
		return Record{}, err
	}

	declared, err := workflows.Load(site.FS())
	if err != nil {
		return Record{}, err
	}

	person := personOf(actor)
	started, err := s.records.Start(ctx, projectID, person, note, summary)
	if err != nil {
		return Record{}, err
	}

	outcome, err := s.engine.Run(ctx, declared.For(workflows.OnPublish, settings), workflows.Run{
		ProjectID: projectID,
		Site:      site.FS(),
		BaseURL:   baseURL,
	})
	if err != nil {
		return s.records.Finish(ctx, started.ID, Failed, nil,
			"We couldn't publish this time. Nothing changed on your live site, and your edits are safe.",
			err.Error())
	}

	state := Succeeded
	trouble := ""
	if !outcome.Succeeded() {
		state = Failed
		trouble = outcome.Sentence()
	}
	return s.records.Finish(ctx, started.ID, state, outcome.Steps, trouble, detailOf(outcome))
}

func detailOf(outcome *workflows.Outcome) string {
	said := []string{}
	for _, step := range outcome.Steps {
		if step.Detail != "" {
			said = append(said, step.Title+":\n"+step.Detail)
		}
	}
	return strings.Join(said, "\n\n")
}

func personOf(actor policy.Actor) *uuid.UUID {
	id, err := uuid.Parse(actor.PersonID)
	if err != nil {
		return nil
	}
	return &id
}

func (s *Service) Get(ctx context.Context, actor policy.Actor, projectID, id uuid.UUID) (Record, error) {
	if err := s.allowed(ctx, actor, policy.ActionReadContent, projectID); err != nil {
		return Record{}, err
	}

	record, err := s.records.Get(ctx, id)
	if err != nil {
		return Record{}, err
	}
	if record.ProjectID != projectID {
		return Record{}, ErrNoSuchPublish
	}
	if s.allowed(ctx, actor, policy.ActionSeeTechnicalDetail, projectID) != nil {
		record.Detail = ""
	}
	return record, nil
}

func (s *Service) History(ctx context.Context, actor policy.Actor, projectID uuid.UUID, limit int32) ([]Record, error) {
	if err := s.allowed(ctx, actor, policy.ActionReadContent, projectID); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = DefaultListLength
	}

	records, err := s.records.List(ctx, projectID, limit)
	if err != nil {
		return nil, err
	}
	if s.allowed(ctx, actor, policy.ActionSeeTechnicalDetail, projectID) != nil {
		for index := range records {
			records[index].Detail = ""
		}
	}
	return records, nil
}

func (s *Service) InterruptLeftovers(ctx context.Context) (int, error) {
	return s.records.InterruptLeftovers(ctx)
}

// allowed asks the one policy function. Somebody who may not read the
// project at all is told it is not there, so that the existence of another
// customer's site is not something this answers.
func (s *Service) allowed(ctx context.Context, actor policy.Actor, action policy.Action, projectID uuid.UUID) error {
	resource, err := s.sources.Resource(ctx, projectID)
	if err != nil {
		return err
	}
	if !policy.Can(actor, policy.ActionReadContent, resource) {
		return httpapi.ErrMissing
	}
	if !policy.Can(actor, action, resource) {
		return httpapi.ErrNotAllowed
	}
	return nil
}
