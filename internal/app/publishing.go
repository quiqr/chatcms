package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms/internal/auth"
	"github.com/quiqr/chatcms/internal/content"
	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/history"
	"github.com/quiqr/chatcms/internal/jobs"
	"github.com/quiqr/chatcms/internal/media"
	"github.com/quiqr/chatcms/internal/model"
	"github.com/quiqr/chatcms/internal/onboarding"
	"github.com/quiqr/chatcms/internal/policy"
	"github.com/quiqr/chatcms/internal/publish"
	"github.com/quiqr/chatcms/internal/store"
	"github.com/quiqr/chatcms/internal/workflows"
)

// publishSources is what a publish needs from the rest of the server: a
// project's working copy, its publishing settings and what it is called.
type publishSources struct {
	sites   *gitsync.Sites
	queries *store.Queries
	baseURL string
}

type workingCopy struct {
	gitsync.Repo
}

func (c workingCopy) Model() (*model.Model, error) {
	assembly, err := model.Assemble(c.FS())
	if err != nil {
		return nil, err
	}
	return model.Decode(assembly)
}

func (p publishSources) Site(ctx context.Context, projectID uuid.UUID) (publish.Site, func() error, error) {
	repository, err := p.sites.Open(ctx, projectID.String())
	if err != nil {
		return nil, nil, err
	}
	return workingCopy{Repo: repository}, repository.Close, nil
}

func (p publishSources) Settings(ctx context.Context, projectID uuid.UUID) (workflows.Settings, error) {
	row, err := p.queries.GetProject(ctx, pgUUID(projectID))
	if err != nil {
		return workflows.Settings{}, err
	}
	return workflows.Settings{ChatCMSBuilds: row.PublishForm == string(onboarding.PublishBuilt)}, nil
}

func (p publishSources) BaseURL(ctx context.Context, projectID uuid.UUID) (string, error) {
	row, err := p.queries.GetProject(ctx, pgUUID(projectID))
	if err != nil {
		return "", err
	}
	if row.Domain != "" {
		return "https://" + row.Domain + "/", nil
	}
	return "", nil
}

func (p publishSources) Resource(ctx context.Context, projectID uuid.UUID) (policy.Resource, error) {
	row, err := p.queries.GetProject(ctx, pgUUID(projectID))
	if err != nil {
		return policy.Resource{}, err
	}
	return policy.Project(uuid.UUID(row.OrganisationID.Bytes).String(), projectID.String()), nil
}

// publishTarget answers where and in what form a project publishes, which is
// what the project's own settings say.
func publishTarget(queries *store.Queries) publish.Targets {
	return func(ctx context.Context, projectID uuid.UUID) (publish.Target, error) {
		row, err := queries.GetProject(ctx, pgUUID(projectID))
		if err != nil {
			return publish.Target{}, err
		}

		target := publish.Target{
			BuiltPages: row.PublishForm == string(onboarding.PublishBuilt),
			Message:    "Published from Quiqr ChatCMS",
		}
		if row.PublishTo == string(onboarding.PublishToItsOwnLine) {
			target.Line = row.PublishBranch
		}
		return target, nil
	}
}

var _ publish.Sources = publishSources{}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

// historySources is what the history needs from the rest of the server: a
// project's working copy, what people wrote when they published, and who is
// acting.
type historySources struct {
	publishSources
	records *publish.Records
	people  *auth.Service
}

type historyCopy struct {
	workingCopy
}

func (h historySources) Site(ctx context.Context, projectID uuid.UUID) (history.Site, func() error, error) {
	repository, err := h.sites.Open(ctx, projectID.String())
	if err != nil {
		return nil, nil, err
	}
	return historyCopy{workingCopy{Repo: repository}}, repository.Close, nil
}

// Notes is what somebody wrote when they published, by the version that
// publish put live. A publish record does not name a version, so the version
// live at the time is the most recent one made before it finished.
func (h historySources) Notes(ctx context.Context, projectID uuid.UUID) (map[string]string, error) {
	records, err := h.records.List(ctx, projectID, 200)
	if err != nil {
		return nil, err
	}

	repository, err := h.sites.Open(ctx, projectID.String())
	if err != nil {
		return nil, err
	}
	defer repository.Close()

	versions, err := repository.History(ctx, history.DefaultLength)
	if err != nil {
		return nil, err
	}

	notes := map[string]string{}
	for _, record := range records {
		if record.Note == "" || record.FinishedAt == nil {
			continue
		}
		for _, version := range versions {
			if !version.Time.After(*record.FinishedAt) {
				if _, taken := notes[version.ID]; !taken {
					notes[version.ID] = record.Note
				}
				break
			}
		}
	}
	return notes, nil
}

func (h historySources) Who(ctx context.Context, personID uuid.UUID) (gitsync.Identity, error) {
	person, err := h.people.People().ByID(ctx, personID)
	if err != nil {
		return gitsync.Identity{}, err
	}
	return gitsync.Identity{Name: person.DisplayName, Email: person.Email}, nil
}

var _ history.Sources = historySources{}

// contentSources is what editing one page needs: a project's working copy,
// who is editing, and what they may do.
type contentSources struct {
	publishSources
	people *auth.Service
}

func (c contentSources) Site(ctx context.Context, projectID uuid.UUID) (content.Site, func() error, error) {
	repository, err := c.sites.Open(ctx, projectID.String())
	if err != nil {
		return nil, nil, err
	}
	return workingCopy{Repo: repository}, repository.Close, nil
}

func (c contentSources) Who(ctx context.Context, personID uuid.UUID) (gitsync.Identity, error) {
	person, err := c.people.People().ByID(ctx, personID)
	if err != nil {
		return gitsync.Identity{}, err
	}
	return gitsync.Identity{Name: person.DisplayName, Email: person.Email}, nil
}

// queueSync asks for the site to be brought level with its remote soon, so
// that an edit reaches it without anybody waiting for the next round.
type queueSync struct {
	queue *jobs.Queue
}

func (q queueSync) SyncSoon(ctx context.Context, projectID uuid.UUID) error {
	if q.queue == nil {
		return nil
	}
	_, err := q.queue.Enqueue(ctx, gitsync.SyncJobKind, projectID.String(), nil)
	return err
}

var _ content.Sources = contentSources{}
var _ content.Syncer = queueSync{}

// mediaSources is what the picture library needs: the same working copy and
// the same person lookup as editing a page.
type mediaSources struct {
	contentSources
}

func (m mediaSources) Site(ctx context.Context, projectID uuid.UUID) (media.Site, func() error, error) {
	repository, err := m.sites.Open(ctx, projectID.String())
	if err != nil {
		return nil, nil, err
	}
	return workingCopy{Repo: repository}, repository.Close, nil
}

var _ media.Sources = mediaSources{}
