package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms/internal/gitsync"
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
