package admin

import (
	"context"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/policy"
)

// Web is the service as the HTTP layer wants it: the same acts, with the
// views that layer speaks in.
type Web struct {
	service *Service
}

func NewWeb(service *Service) *Web { return &Web{service: service} }

func (w *Web) MakeOrganisation(ctx context.Context, who httpapi.Actor, slug, name, ownerEmail string) (httpapi.OrganisationView, error) {
	return w.service.MakeOrganisation(ctx, who, slug, name, ownerEmail)
}

func (w *Web) PutSomebodyInto(ctx context.Context, who httpapi.Actor, slug, email, role string) error {
	// An unknown word is refused rather than quietly treated as the lesser of
	// the two: falling back to the smaller role when somebody meant the
	// larger one is a silent demotion nobody is told about.
	asked := policy.OrganisationRole(role)
	if asked != policy.OrganisationOwner && asked != policy.OrganisationMember {
		return &httpapi.RefusedError{Key: "refused.not-role-organisation-has", Reason: "that is not a role an organisation has"}
	}
	return w.service.PutSomebodyInto(ctx, who, slug, email, asked)
}

func (w *Web) TakeSomebodyOut(ctx context.Context, who httpapi.Actor, slug, email string) error {
	return w.service.TakeSomebodyOut(ctx, who, slug, email)
}

func (w *Web) Appoint(ctx context.Context, who httpapi.Actor, email string) error {
	return w.service.Appoint(ctx, who, email)
}

func (w *Web) StandDown(ctx context.Context, who httpapi.Actor, email string) error {
	return w.service.StandDown(ctx, who, email)
}

func (w *Web) Organisation(ctx context.Context, who httpapi.Actor, slug string) (httpapi.OrganisationWholeView, error) {
	return w.service.Organisation(ctx, who, slug)
}

func (w *Web) RemoveOrganisation(ctx context.Context, who httpapi.Actor, slug string) error {
	return w.service.RemoveOrganisation(ctx, who, slug)
}

func (w *Web) TurnAccountOff(ctx context.Context, who httpapi.Actor, email string) error {
	return w.service.TurnAccountOff(ctx, who, email)
}

func (w *Web) TurnAccountOn(ctx context.Context, who httpapi.Actor, email string) error {
	return w.service.TurnAccountOn(ctx, who, email)
}

func (w *Web) Records(ctx context.Context, who httpapi.Actor, most int32) ([]httpapi.RecordView, error) {
	records, err := w.service.Records(ctx, who, most)
	if err != nil {
		return nil, err
	}

	out := make([]httpapi.RecordView, 0, len(records))
	for _, record := range records {
		out = append(out, httpapi.RecordView{
			Who:    record.PersonSaid,
			Did:    string(record.Did),
			To:     record.SubjectSaid,
			Detail: record.Detail,
			At:     record.At,
		})
	}
	return out, nil
}

func (w *Web) Administrators(ctx context.Context, who httpapi.Actor) ([]httpapi.AdministratorView, error) {
	if _, err := w.service.mayAdminister(ctx, who, policy.ActionAdministerInstallation); err != nil {
		return nil, err
	}

	rows, err := w.service.store.db.Queries.ListAdministrators(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]httpapi.AdministratorView, 0, len(rows))
	for _, row := range rows {
		out = append(out, httpapi.AdministratorView{
			Email:       row.Email,
			DisplayName: row.DisplayName,
			Yourself:    uuid.UUID(row.ID.Bytes).String() == who.ID,
		})
	}
	return out, nil
}

// MayAdminister says whether this person administers the installation, so
// that the interface can leave out what it must rather than offering it and
// refusing afterwards.
func (w *Web) MayAdminister(ctx context.Context, who httpapi.Actor) bool {
	_, err := w.service.mayAdminister(ctx, who, policy.ActionAdministerInstallation)
	return err == nil
}

func (w *Web) Organisations(ctx context.Context, who httpapi.Actor) ([]httpapi.OrganisationSummaryView, error) {
	if _, err := w.service.mayAdminister(ctx, who, policy.ActionSeeEveryOrganisation); err != nil {
		return nil, err
	}

	rows, err := w.service.store.db.Queries.ListEveryOrganisation(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]httpapi.OrganisationSummaryView, 0, len(rows))
	for _, row := range rows {
		out = append(out, httpapi.OrganisationSummaryView{
			Slug:     row.Slug,
			Name:     row.Name,
			Projects: row.ProjectCount,
			People:   row.PeopleCount,
		})
	}
	return out, nil
}

func (w *Web) People(ctx context.Context, who httpapi.Actor, search string, most int32) (httpapi.PeopleView, error) {
	people, altogether, err := w.service.People(ctx, who, search, most)
	if err != nil {
		return httpapi.PeopleView{}, err
	}

	out := httpapi.PeopleView{
		People:     make([]httpapi.PersonWholeView, 0, len(people)),
		Altogether: altogether,
		Shown:      int64(len(people)),
	}
	for _, person := range people {
		out.People = append(out.People, personView(person))
	}
	return out, nil
}

func (w *Web) Person(ctx context.Context, who httpapi.Actor, email string) (httpapi.PersonWholeView, error) {
	person, err := w.service.Person(ctx, who, email)
	if err != nil {
		return httpapi.PersonWholeView{}, err
	}
	return personView(person), nil
}

func (w *Web) MakePerson(ctx context.Context, who httpapi.Actor, email, displayName string) (httpapi.PersonWholeView, error) {
	made, err := w.service.MakePerson(ctx, who, email, displayName)
	if err != nil {
		return httpapi.PersonWholeView{}, err
	}
	return personView(made), nil
}

func personView(person PersonWithReach) httpapi.PersonWholeView {
	view := httpapi.PersonWholeView{
		PersonID:      person.ID,
		Email:         person.Email,
		DisplayName:   person.DisplayName,
		Administers:   person.Administers,
		AccountOn:     person.AccountOn,
		StillToAccept: person.StillToAccept,
		Organisations: make([]httpapi.BelongingView, 0, len(person.Organisations)),
		Projects:      make([]httpapi.BelongingView, 0, len(person.Projects)),
	}
	if person.EverSeen {
		seen := person.LastSeen
		view.LastSeen = &seen
	}
	for _, where := range person.Organisations {
		view.Organisations = append(view.Organisations, httpapi.BelongingView{
			Slug: where.Slug, Name: where.Name, Role: where.Role,
		})
	}
	for _, where := range person.Projects {
		view.Projects = append(view.Projects, httpapi.BelongingView{
			Slug: where.Slug, Name: where.Name, Organisation: where.Organisation, Role: where.Role,
		})
	}
	return view
}

func (w *Web) ActAs(ctx context.Context, who httpapi.Actor, token, email string) error {
	return w.service.ActAs(ctx, who, token, email)
}

func (w *Web) StopActingAs(ctx context.Context, who httpapi.Actor, token string) error {
	return w.service.StopActingAs(ctx, who, token)
}
