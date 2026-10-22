package onboarding

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/policy"
	"github.com/quiqr/chatcms/internal/store"
	"github.com/quiqr/chatcms/internal/tenancy"
)

// Service answers the HTTP layer's questions about setting a project up. It
// asks internal/policy whether the person may, and never decides itself.
type Service struct {
	db       *store.DB
	tenancy  *tenancy.Store
	keys     *Keys
	importer *Importer
}

func NewService(db *store.DB, tenancyStore *tenancy.Store, keys *Keys, importer *Importer) *Service {
	return &Service{db: db, tenancy: tenancyStore, keys: keys, importer: importer}
}

// reach finds a project the asker may act on in the given way, answering the
// same for one they cannot reach as for one that is not there.
func (s *Service) reach(
	ctx context.Context,
	who httpapi.Actor,
	organisationSlug, projectSlug string,
	action policy.Action,
) (tenancy.Project, error) {
	if who.Anonymous || who.ID == "" {
		return tenancy.Project{}, httpapi.ErrNotAllowed
	}
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return tenancy.Project{}, httpapi.ErrNotAllowed
	}

	actor, err := s.tenancy.ActorFor(ctx, personID)
	if err != nil {
		return tenancy.Project{}, err
	}

	project, err := s.tenancy.ProjectBySlug(ctx, organisationSlug, projectSlug)
	if errors.Is(err, tenancy.ErrNotFound) {
		return tenancy.Project{}, httpapi.ErrMissing
	}
	if err != nil {
		return tenancy.Project{}, err
	}

	resource := policy.Project(project.OrganisationID.String(), project.ID.String())
	if !policy.Can(actor, policy.ActionReadContent, resource) {
		return tenancy.Project{}, httpapi.ErrMissing
	}
	if !policy.Can(actor, action, resource) {
		return tenancy.Project{}, httpapi.ErrNotAllowed
	}
	return project, nil
}

const keyAdvice = "Add this key to the repository's deploy keys on your git host, with write " +
	"access, then test the connection."

func (s *Service) GenerateKey(ctx context.Context, who httpapi.Actor, organisation, project string) (httpapi.KeyView, error) {
	found, err := s.reach(ctx, who, organisation, project, policy.ActionChangeSettings)
	if err != nil {
		return httpapi.KeyView{}, err
	}

	public, err := s.keys.Generate(ctx, found.ID, "chatcms "+organisation+"/"+project)
	if err != nil {
		return httpapi.KeyView{}, err
	}
	return httpapi.KeyView{PublicKey: public, Advice: keyAdvice}, nil
}

func (s *Service) ReadKey(ctx context.Context, who httpapi.Actor, organisation, project string) (httpapi.KeyView, error) {
	found, err := s.reach(ctx, who, organisation, project, policy.ActionSeeTechnicalDetail)
	if err != nil {
		return httpapi.KeyView{}, err
	}

	public, err := s.keys.PublicKey(ctx, found.ID)
	if errors.Is(err, ErrNoKeyYet) {
		return httpapi.KeyView{Advice: "This project has no key yet. Make one to continue."}, nil
	}
	if err != nil {
		return httpapi.KeyView{}, err
	}
	return httpapi.KeyView{PublicKey: public, Advice: keyAdvice}, nil
}

func (s *Service) ImportSite(ctx context.Context, who httpapi.Actor, organisation, project, kind, address, remote string) (httpapi.ImportView, error) {
	found, err := s.reach(ctx, who, organisation, project, policy.ActionChangeSettings)
	if err != nil {
		return httpapi.ImportView{}, err
	}

	result, err := s.importer.Import(ctx, found.ID, Request{
		Kind: SourceKind(kind), Address: address, Remote: remote,
	})
	if err != nil {
		return httpapi.ImportView{}, importRefusal(err)
	}

	return importView(result.Kind, result.Remote, result.Inspection), nil
}

func (s *Service) InspectSite(ctx context.Context, who httpapi.Actor, organisation, project string) (httpapi.ImportView, error) {
	found, err := s.reach(ctx, who, organisation, project, policy.ActionReadContent)
	if err != nil {
		return httpapi.ImportView{}, err
	}

	row, err := s.db.Queries.GetProject(ctx, pgtype.UUID{Bytes: found.ID, Valid: true})
	if err != nil {
		return httpapi.ImportView{}, err
	}
	if SourceKind(row.SourceKind) == SourceNone {
		return httpapi.ImportView{Kind: string(SourceNone)}, nil
	}

	inspection, err := s.importer.Inspect(ctx, found.ID)
	if err != nil {
		return httpapi.ImportView{}, err
	}
	return importView(SourceKind(row.SourceKind), row.Remote, inspection), nil
}

func (s *Service) TestConnection(ctx context.Context, who httpapi.Actor, organisation, project string) (httpapi.ConnectionView, error) {
	found, err := s.reach(ctx, who, organisation, project, policy.ActionSeeTechnicalDetail)
	if err != nil {
		return httpapi.ConnectionView{}, err
	}

	row, err := s.db.Queries.GetProject(ctx, pgtype.UUID{Bytes: found.ID, Valid: true})
	if err != nil {
		return httpapi.ConnectionView{}, err
	}

	connection, err := s.importer.TestConnection(ctx, found.ID, row.Remote)
	if err != nil {
		return httpapi.ConnectionView{}, err
	}
	return httpapi.ConnectionView{
		Good:      connection.Good,
		Message:   connection.Message,
		PublicKey: connection.PublicKey,
	}, nil
}

func (s *Service) ReadPublishing(ctx context.Context, who httpapi.Actor, organisation, project string) (httpapi.PublishingView, error) {
	found, err := s.reach(ctx, who, organisation, project, policy.ActionReadContent)
	if err != nil {
		return httpapi.PublishingView{}, err
	}

	row, err := s.db.Queries.GetProject(ctx, pgtype.UUID{Bytes: found.ID, Valid: true})
	if err != nil {
		return httpapi.PublishingView{}, err
	}
	return httpapi.PublishingView{To: row.PublishTo, Branch: row.PublishBranch, Form: row.PublishForm}, nil
}

func (s *Service) SetPublishing(ctx context.Context, who httpapi.Actor, organisation, project, to, branch, form string) (httpapi.PublishingView, error) {
	found, err := s.reach(ctx, who, organisation, project, policy.ActionChangeSettings)
	if err != nil {
		return httpapi.PublishingView{}, err
	}

	if err := s.importer.SetPublishing(ctx, found.ID, Publishing{
		To: PublishTo(to), Branch: branch, Form: PublishForm(form),
	}); err != nil {
		return httpapi.PublishingView{}, importRefusal(err)
	}

	return s.ReadPublishing(ctx, who, organisation, project)
}

func importView(kind SourceKind, remote string, inspection Inspection) httpapi.ImportView {
	return httpapi.ImportView{
		Kind:        string(kind),
		Remote:      remote,
		IsHugoSite:  inspection.IsHugoSite,
		HasModel:    inspection.HasModel,
		Ready:       inspection.Ready(),
		Pages:       inspection.Pages,
		Collections: inspection.Collections,
		Problem:     inspection.Problem,
		LookedFor:   inspection.LookedFor,
	}
}

// importRefusal turns a failure an implementer can act on into one the HTTP
// layer shows, and leaves anything else as it is.
func importRefusal(err error) error {
	switch {
	case errors.Is(err, ErrUnusableBranchName),
		errors.Is(err, ErrNoSuchTemplate),
		errors.Is(err, ErrExportUnreadable),
		errors.Is(err, ErrNoKeyYet):
		return &httpapi.RefusedError{Reason: err.Error()}
	}

	// Anything git said is worth showing: it is usually the address.
	if err != nil {
		return &httpapi.RefusedError{Reason: fmt.Sprintf("the site could not be imported: %v", err)}
	}
	return nil
}
