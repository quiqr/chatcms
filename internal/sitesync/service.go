package sitesync

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/forms"
	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/model"
	"github.com/quiqr/chatcms/internal/policy"
	"github.com/quiqr/chatcms/internal/publish"
)

// Site is what saying how a site stands needs from its working copy.
type Site interface {
	Status(ctx context.Context) (gitsync.Status, error)
	Sync(ctx context.Context) (gitsync.SyncResult, error)
	Resolve(ctx context.Context, resolutions map[string][]byte) (gitsync.SyncResult, error)
	Abandon(ctx context.Context) error
	History(ctx context.Context, limit int) ([]gitsync.Version, error)
	Model() (*model.Model, error)
}

// Sources opens a project's site and says what a person may do with it.
type Sources interface {
	Site(ctx context.Context, projectID uuid.UUID) (Site, func() error, error)
	Resource(ctx context.Context, projectID uuid.UUID) (policy.Resource, error)
	// Technical says where a site keeps its work, for somebody allowed to see
	// such things. Nobody else is ever shown it.
	Technical(ctx context.Context, projectID uuid.UUID) (Technical, error)
}

// Technical is the detail an implementer may see on hover. It is the only
// place in this package where a git word is allowed.
type Technical struct {
	Line   string
	Remote string
	Head   string
}

var ErrIncomplete = errors.New("sitesync: the service is missing a part it needs")

type Service struct {
	sources Sources
}

func NewService(sources Sources) (*Service, error) {
	if sources == nil {
		return nil, ErrIncomplete
	}
	return &Service{sources: sources}, nil
}

// State is how a site stands, in the five words the designs use.
type State string

const (
	// Saved says everything somebody typed is in the site.
	Saved State = "saved"
	// Saving says something is on its way.
	Saving State = "saving"
	// Offline says the site could not be reached and the work is here.
	Offline State = "offline"
	// Arrived says somebody else changed the site and it has been brought in.
	Arrived State = "arrived"
	// Asking says two people changed the same thing and somebody has to say
	// which to keep.
	Asking State = "asking"
)

// Standing is everything the indicator and the sync screen show.
type Standing struct {
	State State
	// Unpublished is how much is waiting to go out.
	Unpublished int
	// ArrivedFrom names the people whose work was brought in, most recent
	// first, so the announcement can say who.
	ArrivedFrom []string
	// Conflicts is what two people changed, as pages and fields.
	Conflicts []PageConflict
	// Technical is only filled in for somebody allowed technical detail.
	Technical *Technical
}

// Standing says how a site stands without changing anything.
func (s *Service) Standing(ctx context.Context, actor policy.Actor, projectID uuid.UUID) (Standing, error) {
	if err := s.allowed(ctx, actor, policy.ActionReadContent, projectID); err != nil {
		return Standing{}, err
	}

	site, done, err := s.sources.Site(ctx, projectID)
	if err != nil {
		return Standing{}, err
	}
	defer done()

	status, err := site.Status(ctx)
	if err != nil {
		return Standing{}, err
	}

	standing := Standing{State: Saved, Unpublished: status.Unpublished}
	if status.UnsavedWork {
		standing.State = Saving
	}
	if status.ExchangeInProgress {
		standing.State = Asking
	}

	s.addTechnical(ctx, actor, projectID, &standing)
	return standing, nil
}

// Bring brings in whatever somebody else changed. Where two people changed
// the same thing it does not choose: it says what the question is.
func (s *Service) Bring(ctx context.Context, actor policy.Actor, projectID uuid.UUID) (Standing, error) {
	if err := s.allowed(ctx, actor, policy.ActionEditContent, projectID); err != nil {
		return Standing{}, err
	}

	site, done, err := s.sources.Site(ctx, projectID)
	if err != nil {
		return Standing{}, err
	}
	defer done()

	result, err := site.Sync(ctx)
	if errors.Is(err, gitsync.ErrRemoteUnreachable) {
		standing := Standing{State: Offline}
		s.addTechnical(ctx, actor, projectID, &standing)
		return standing, nil
	}
	if err != nil {
		return Standing{}, err
	}

	return s.after(ctx, actor, projectID, site, result)
}

func (s *Service) after(
	ctx context.Context,
	actor policy.Actor,
	projectID uuid.UUID,
	site Site,
	result gitsync.SyncResult,
) (Standing, error) {
	status, err := site.Status(ctx)
	if err != nil {
		return Standing{}, err
	}

	standing := Standing{State: Saved, Unpublished: status.Unpublished}
	for _, version := range result.Arrived {
		if version.Author.Name != "" && !carries(standing.ArrivedFrom, version.Author.Name) {
			standing.ArrivedFrom = append(standing.ArrivedFrom, version.Author.Name)
		}
	}
	if len(result.Arrived) > 0 {
		standing.State = Arrived
	}

	if result.Conflict != nil {
		m, _ := site.Model()
		standing.State = Asking
		standing.Conflicts = Describe(result.Conflict, m, strings.Join(standing.ArrivedFrom, " and "))
	}

	s.addTechnical(ctx, actor, projectID, &standing)
	return standing, nil
}

func carries(names []string, name string) bool {
	for _, held := range names {
		if held == name {
			return true
		}
	}
	return false
}

// Choice is what to keep for one field: what was written here, or what
// arrived.
type Choice string

const (
	Mine   Choice = "mine"
	Theirs Choice = "theirs"
)

// Answer says which version to keep, field by field, for one page.
type Answer struct {
	// File is the page being answered about.
	File string
	// Fields maps a field's path to which version to keep. A field left out
	// keeps what was written here.
	Fields map[string]Choice
	// Whole, for a page whose difference could not be put as fields, says to
	// keep one whole version.
	Whole Choice
}

// Answered takes what somebody chose and finishes the exchange.
func (s *Service) Answered(ctx context.Context, actor policy.Actor, projectID uuid.UUID, answers []Answer) (Standing, error) {
	if err := s.allowed(ctx, actor, policy.ActionEditContent, projectID); err != nil {
		return Standing{}, err
	}

	site, done, err := s.sources.Site(ctx, projectID)
	if err != nil {
		return Standing{}, err
	}
	defer done()

	result, err := site.Sync(ctx)
	if err != nil && !errors.Is(err, gitsync.ErrExchangeInProgress) {
		return Standing{}, err
	}
	if result.Conflict == nil {
		// Nothing is waiting on an answer any more, which can happen if
		// somebody answered in another window.
		return s.after(ctx, actor, projectID, site, result)
	}

	m, _ := site.Model()
	resolutions := map[string][]byte{}
	for _, path := range result.Conflict.Paths {
		answer, found := answerFor(answers, path.Path)
		if !found {
			continue
		}
		content, err := resolve(path, m, answer)
		if err != nil {
			return Standing{}, &httpapi.RefusedError{Reason: err.Error()}
		}
		resolutions[path.Path] = content
	}

	if len(resolutions) != len(result.Conflict.Paths) {
		return Standing{}, &httpapi.RefusedError{
			Key:    "refused.every-page-two-people-changed",
			Reason: "every page that two people changed needs an answer before this can finish",
		}
	}

	finished, err := site.Resolve(ctx, resolutions)
	if err != nil {
		return Standing{}, err
	}
	return s.after(ctx, actor, projectID, site, finished)
}

func answerFor(answers []Answer, file string) (Answer, bool) {
	for _, answer := range answers {
		if answer.File == file {
			return answer, true
		}
	}
	return Answer{}, false
}

// resolve builds the content to keep: whichever whole version was chosen, or
// the version written here with the fields somebody wanted from the other.
func resolve(path gitsync.ConflictedPath, m *model.Model, answer Answer) ([]byte, error) {
	if answer.Whole == Theirs {
		return path.Remote, nil
	}
	if answer.Whole == Mine || len(answer.Fields) == 0 {
		return path.Local, nil
	}

	place := publish.LocateFields(m, path.Path)
	mine, err := forms.Load(place, path.Path, path.Local)
	if err != nil {
		return nil, fmt.Errorf("this page could not be read: %w", err)
	}
	theirs, err := forms.Load(place, path.Path, path.Remote)
	if err != nil {
		return nil, fmt.Errorf("the version that arrived could not be read: %w", err)
	}

	for field, choice := range answer.Fields {
		if choice != Theirs {
			continue
		}
		if field == forms.BodyKey {
			if err := mine.Set(forms.BodyKey, theirs.Body()); err != nil {
				return nil, err
			}
			continue
		}
		held, found := theirs.Get(field)
		if !found {
			if err := mine.Clear(field); err != nil {
				return nil, err
			}
			continue
		}
		if err := mine.Set(field, held); err != nil {
			return nil, err
		}
	}
	return mine.Bytes()
}

// Abandon gives up on an exchange and puts the site back as it was, for
// somebody who would rather start again than answer.
func (s *Service) Abandon(ctx context.Context, actor policy.Actor, projectID uuid.UUID) (Standing, error) {
	if err := s.allowed(ctx, actor, policy.ActionEditContent, projectID); err != nil {
		return Standing{}, err
	}

	site, done, err := s.sources.Site(ctx, projectID)
	if err != nil {
		return Standing{}, err
	}
	defer done()

	if err := site.Abandon(ctx); err != nil {
		return Standing{}, err
	}
	return s.Standing(ctx, actor, projectID)
}

func (s *Service) addTechnical(ctx context.Context, actor policy.Actor, projectID uuid.UUID, standing *Standing) {
	resource, err := s.sources.Resource(ctx, projectID)
	if err != nil || !policy.Can(actor, policy.ActionSeeTechnicalDetail, resource) {
		return
	}
	if technical, err := s.sources.Technical(ctx, projectID); err == nil {
		standing.Technical = &technical
	}
}

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
