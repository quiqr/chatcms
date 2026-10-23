package history

import (
	"context"
	"errors"
	"io/fs"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/model"
	"github.com/quiqr/chatcms/internal/policy"
	"github.com/quiqr/chatcms/internal/publish"
)

// Site is what the history needs from one project's working copy.
type Site interface {
	History(ctx context.Context, limit int) ([]gitsync.Version, error)
	Status(ctx context.Context) (gitsync.Status, error)
	ChangesIn(ctx context.Context, version string) ([]gitsync.Change, error)
	At(ctx context.Context, version string) (fs.FS, error)
	Before(ctx context.Context, version string) (fs.FS, error)
	GoBackTo(ctx context.Context, version string, author gitsync.Identity, message string) (gitsync.Version, error)
	ReadPublished(ctx context.Context, name string) ([]byte, error)
	Model() (*model.Model, error)
	FS() fs.FS
}

// Sources opens a project's site and says what a person may do with it.
type Sources interface {
	Site(ctx context.Context, projectID uuid.UUID) (Site, func() error, error)
	Resource(ctx context.Context, projectID uuid.UUID) (policy.Resource, error)
	// Notes is what people wrote when they published, by the version each
	// publish put live.
	Notes(ctx context.Context, projectID uuid.UUID) (map[string]string, error)
	// Who is the person acting, as the history should record them.
	Who(ctx context.Context, personID uuid.UUID) (gitsync.Identity, error)
}

// Publisher puts the site live. Going back offers to publish right away, and
// this is what that calls.
type Publisher interface {
	Run(ctx context.Context, actor policy.Actor, projectID uuid.UUID, note string) (publish.Record, error)
}

var (
	ErrIncomplete = errors.New("history: the service is missing a part it needs")
	// ErrNothingToUndo says the site already holds what that version left.
	ErrNothingToUndo = errors.New("history: the site is already as that version left it")
)

const DefaultLength = 100

type Service struct {
	sources   Sources
	publisher Publisher
	now       func() time.Time
}

func NewService(sources Sources, publisher Publisher) (*Service, error) {
	if sources == nil || publisher == nil {
		return nil, ErrIncomplete
	}
	return &Service{sources: sources, publisher: publisher, now: time.Now}, nil
}

// Timeline is the history of the site, grouped by period, most recent first.
func (s *Service) Timeline(ctx context.Context, actor policy.Actor, projectID uuid.UUID) (Timeline, error) {
	if err := s.allowed(ctx, actor, policy.ActionReadContent, projectID); err != nil {
		return Timeline{}, err
	}

	site, done, err := s.sources.Site(ctx, projectID)
	if err != nil {
		return Timeline{}, err
	}
	defer done()

	versions, err := site.History(ctx, DefaultLength)
	if err != nil {
		return Timeline{}, err
	}
	status, err := site.Status(ctx)
	if err != nil {
		return Timeline{}, err
	}
	notes, err := s.sources.Notes(ctx, projectID)
	if err != nil {
		return Timeline{}, err
	}

	return Build(versions, notes, liveVersion(versions, status.Unpublished), status.Unpublished, s.now()), nil
}

// liveVersion is the version the published site holds: the most recent one
// that is not still waiting to be published.
func liveVersion(versions []gitsync.Version, unpublished int) string {
	if unpublished >= len(versions) {
		return ""
	}
	return versions[unpublished].ID
}

// Contents is what one version changed, as pages and fields.
func (s *Service) Contents(ctx context.Context, actor policy.Actor, projectID uuid.UUID, version string) (publish.Summary, error) {
	if err := s.allowed(ctx, actor, policy.ActionReadContent, projectID); err != nil {
		return publish.Summary{}, err
	}

	site, done, err := s.sources.Site(ctx, projectID)
	if err != nil {
		return publish.Summary{}, err
	}
	defer done()

	asItWas, err := site.At(ctx, version)
	if err != nil {
		return publish.Summary{}, notThere(err)
	}

	m, _ := site.Model()
	return publish.Describe(ctx, &oneVersion{site: site, version: version, content: asItWas}, m)
}

// oneVersion answers the questions a summary asks, but about one version
// rather than about everything waiting to be published: what it changed, what
// the files said before it, and what they say in it.
type oneVersion struct {
	site    Site
	version string
	content fs.FS
}

func (o *oneVersion) Changes(ctx context.Context) ([]gitsync.Change, error) {
	return o.site.ChangesIn(ctx, o.version)
}

func (o *oneVersion) ReadPublished(ctx context.Context, name string) ([]byte, error) {
	before, err := o.site.Before(ctx, o.version)
	if err != nil {
		return nil, gitsync.ErrNotPublished
	}
	content, err := fs.ReadFile(before, name)
	if err != nil {
		return nil, gitsync.ErrNotPublished
	}
	return content, nil
}

func (o *oneVersion) FS() fs.FS { return o.content }

// AsItWas is the whole site as one version left it, for previewing.
func (s *Service) AsItWas(ctx context.Context, actor policy.Actor, projectID uuid.UUID, version string) (fs.FS, func() error, error) {
	if err := s.allowed(ctx, actor, policy.ActionReadContent, projectID); err != nil {
		return nil, nil, err
	}

	site, done, err := s.sources.Site(ctx, projectID)
	if err != nil {
		return nil, nil, err
	}

	asItWas, err := site.At(ctx, version)
	if err != nil {
		done()
		return nil, nil, notThere(err)
	}
	return asItWas, done, nil
}

// Warning is what somebody is told before they go back, so the confirmation
// says what will happen rather than asking them to trust it.
type Warning struct {
	// SetAside is how much unpublished work will stop being in the site. It
	// is not lost: it stays in the history, and going back to it brings it
	// out again.
	SetAside int `json:"setAside"`
	// Title is what the version they are going back to is called.
	Title string `json:"title"`
	// When is when that version was made.
	When time.Time `json:"when"`
}

// Before answers what going back to a version would do.
func (s *Service) Before(ctx context.Context, actor policy.Actor, projectID uuid.UUID, version string) (Warning, error) {
	if err := s.allowed(ctx, actor, policy.ActionGoBackToAVersion, projectID); err != nil {
		return Warning{}, err
	}

	site, done, err := s.sources.Site(ctx, projectID)
	if err != nil {
		return Warning{}, err
	}
	defer done()

	versions, err := site.History(ctx, DefaultLength)
	if err != nil {
		return Warning{}, err
	}
	status, err := site.Status(ctx)
	if err != nil {
		return Warning{}, err
	}
	notes, err := s.sources.Notes(ctx, projectID)
	if err != nil {
		return Warning{}, err
	}

	for _, candidate := range versions {
		if candidate.ID != version {
			continue
		}
		return Warning{
			SetAside: status.Unpublished,
			Title:    titleOf(candidate, notes[candidate.ID]),
			When:     candidate.Time,
		}, nil
	}
	return Warning{}, httpapi.ErrMissing
}

// Result is what happened when somebody went back.
type Result struct {
	Version   string          `json:"version"`
	Title     string          `json:"title"`
	Published *publish.Record `json:"published,omitempty"`
}

// GoBackTo makes the site hold what an earlier version held, and publishes it
// straight away when asked. Nothing is thrown away: the versions in between
// stay in the history, so this can itself be undone.
func (s *Service) GoBackTo(
	ctx context.Context,
	actor policy.Actor,
	projectID uuid.UUID,
	version string,
	publishRightAway bool,
) (Result, error) {
	if err := s.allowed(ctx, actor, policy.ActionGoBackToAVersion, projectID); err != nil {
		return Result{}, err
	}

	warning, err := s.Before(ctx, actor, projectID, version)
	if err != nil {
		return Result{}, err
	}

	who, err := s.who(ctx, actor)
	if err != nil {
		return Result{}, err
	}

	made, err := s.goBack(ctx, projectID, version, who, "Go back to: "+warning.Title)
	if err != nil {
		return Result{}, err
	}

	result := Result{Version: made.ID, Title: warning.Title}
	if !publishRightAway {
		return result, nil
	}

	record, err := s.publisher.Run(ctx, actor, projectID, "Went back to: "+warning.Title)
	if err != nil {
		return result, err
	}
	result.Published = &record
	return result, nil
}

func (s *Service) goBack(ctx context.Context, projectID uuid.UUID, version string, who gitsync.Identity, message string) (gitsync.Version, error) {
	site, done, err := s.sources.Site(ctx, projectID)
	if err != nil {
		return gitsync.Version{}, err
	}
	defer done()

	made, err := site.GoBackTo(ctx, version, who, message)
	if errors.Is(err, gitsync.ErrAlreadyThere) {
		return gitsync.Version{}, ErrNothingToUndo
	}
	return made, notThere(err)
}

func (s *Service) who(ctx context.Context, actor policy.Actor) (gitsync.Identity, error) {
	personID, err := uuid.Parse(actor.PersonID)
	if err != nil {
		return gitsync.Identity{}, httpapi.ErrNotAllowed
	}
	return s.sources.Who(ctx, personID)
}

func notThere(err error) error {
	if errors.Is(err, gitsync.ErrNoSuchVersion) {
		return httpapi.ErrMissing
	}
	return err
}

// allowed asks the one policy function. Somebody who may not read the project
// at all is told it is not there, the same as for a project that does not
// exist.
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
