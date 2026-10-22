package tenancy

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/auth"
	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/policy"
)

// Service answers the HTTP layer's questions about organisations, projects
// and people. It loads what the policy needs and asks it; it never decides.
type Service struct {
	store   *Store
	inviter *Inviter
}

func NewService(store *Store, service *auth.Service) *Service {
	return &Service{store: store, inviter: NewInviter(store, service)}
}

// Store exposes the store, which the tests and later epics use.
func (s *Service) Store() *Store { return s.store }

func (s *Service) actor(ctx context.Context, who httpapi.Actor) (policy.Actor, error) {
	if who.Anonymous || who.ID == "" {
		return policy.Nobody(), nil
	}
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return policy.Nobody(), nil
	}
	return s.store.ActorFor(ctx, personID)
}

func (s *Service) ListOrganisations(ctx context.Context, who httpapi.Actor) ([]httpapi.OrganisationView, error) {
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return nil, httpapi.ErrNotAllowed
	}

	organisations, err := s.store.OrganisationsFor(ctx, personID)
	if err != nil {
		return nil, err
	}

	views := make([]httpapi.OrganisationView, 0, len(organisations))
	for _, organisation := range organisations {
		views = append(views, httpapi.OrganisationView{Slug: organisation.Slug, Name: organisation.Name})
	}
	return views, nil
}

func (s *Service) CreateOrganisation(ctx context.Context, who httpapi.Actor, slug, name string) (httpapi.OrganisationView, error) {
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return httpapi.OrganisationView{}, httpapi.ErrNotAllowed
	}

	organisation, err := s.store.CreateOrganisation(ctx, slug, name, personID)
	if err != nil {
		return httpapi.OrganisationView{}, refusal(err)
	}
	return httpapi.OrganisationView{Slug: organisation.Slug, Name: organisation.Name}, nil
}

func (s *Service) ListProjects(ctx context.Context, who httpapi.Actor, organisationSlug string) ([]httpapi.ProjectView, error) {
	organisation, actor, err := s.reach(ctx, who, organisationSlug)
	if err != nil {
		return nil, err
	}

	projects, err := s.store.ProjectsIn(ctx, organisation.ID)
	if err != nil {
		return nil, err
	}

	views := make([]httpapi.ProjectView, 0, len(projects))
	for _, project := range projects {
		resource := policy.Project(organisation.ID.String(), project.ID.String())
		if !policy.Can(actor, policy.ActionReadContent, resource) {
			continue
		}
		views = append(views, projectView(project, actor, resource))
	}
	return views, nil
}

func (s *Service) CreateProject(ctx context.Context, who httpapi.Actor, organisationSlug, slug, name, domain string) (httpapi.ProjectView, error) {
	organisation, actor, err := s.reach(ctx, who, organisationSlug)
	if err != nil {
		return httpapi.ProjectView{}, err
	}

	if !policy.Can(actor, policy.ActionCreateProject, policy.Organisation(organisation.ID.String())) {
		return httpapi.ProjectView{}, httpapi.ErrNotAllowed
	}

	project, err := s.store.CreateProject(ctx, organisation.ID, slug, name, domain)
	if err != nil {
		return httpapi.ProjectView{}, refusal(err)
	}

	// The creator holds the owner role through the organisation; giving the
	// row as well keeps the project's own list of people honest.
	if err := s.store.SetProjectRole(ctx, project.ID, mustParse(who.ID), policy.RoleOwner); err != nil {
		return httpapi.ProjectView{}, err
	}

	actor, err = s.actor(ctx, who)
	if err != nil {
		return httpapi.ProjectView{}, err
	}
	resource := policy.Project(organisation.ID.String(), project.ID.String())
	return projectView(project, actor, resource), nil
}

func (s *Service) GetProject(ctx context.Context, who httpapi.Actor, organisationSlug, projectSlug string) (httpapi.ProjectView, error) {
	project, actor, resource, err := s.reachProject(ctx, who, organisationSlug, projectSlug, policy.ActionReadContent)
	if err != nil {
		return httpapi.ProjectView{}, err
	}
	return projectView(project, actor, resource), nil
}

func (s *Service) ListPeople(ctx context.Context, who httpapi.Actor, organisationSlug, projectSlug string) ([]httpapi.MemberView, error) {
	project, _, _, err := s.reachProject(ctx, who, organisationSlug, projectSlug, policy.ActionManagePeople)
	if err != nil {
		return nil, err
	}

	members, err := s.store.Members(ctx, project.ID)
	if err != nil {
		return nil, err
	}

	views := make([]httpapi.MemberView, 0, len(members))
	for _, member := range members {
		views = append(views, memberView(member))
	}
	return views, nil
}

func (s *Service) InvitePerson(ctx context.Context, who httpapi.Actor, organisationSlug, projectSlug, email, role string) (httpapi.MemberView, error) {
	project, _, _, err := s.reachProject(ctx, who, organisationSlug, projectSlug, policy.ActionManagePeople)
	if err != nil {
		return httpapi.MemberView{}, err
	}

	invitation, err := s.inviter.Invite(ctx, project.ID, email, policy.Role(role))
	if err != nil {
		return httpapi.MemberView{}, refusal(err)
	}

	return httpapi.MemberView{
		PersonID:      invitation.PersonID.String(),
		Email:         invitation.Email,
		Role:          string(invitation.Role),
		StillToAccept: invitation.LinkIssued,
	}, nil
}

func (s *Service) ChangeRole(ctx context.Context, who httpapi.Actor, organisationSlug, projectSlug, personID, role string) (httpapi.MemberView, error) {
	project, _, _, err := s.reachProject(ctx, who, organisationSlug, projectSlug, policy.ActionManagePeople)
	if err != nil {
		return httpapi.MemberView{}, err
	}

	person, err := uuid.Parse(personID)
	if err != nil {
		return httpapi.MemberView{}, httpapi.ErrMissing
	}
	if !knownRole(policy.Role(role)) {
		return httpapi.MemberView{}, &httpapi.RefusedError{Reason: ErrNotARole.Error()}
	}

	if err := s.store.SetProjectRole(ctx, project.ID, person, policy.Role(role)); err != nil {
		return httpapi.MemberView{}, err
	}

	members, err := s.store.Members(ctx, project.ID)
	if err != nil {
		return httpapi.MemberView{}, err
	}
	for _, member := range members {
		if member.PersonID == person {
			return memberView(member), nil
		}
	}
	return httpapi.MemberView{}, httpapi.ErrMissing
}

func (s *Service) RemovePerson(ctx context.Context, who httpapi.Actor, organisationSlug, projectSlug, personID string) error {
	project, _, _, err := s.reachProject(ctx, who, organisationSlug, projectSlug, policy.ActionManagePeople)
	if err != nil {
		return err
	}

	person, err := uuid.Parse(personID)
	if err != nil {
		return httpapi.ErrMissing
	}

	if err := s.store.RemoveFromProject(ctx, project.ID, person); err != nil {
		return refusal(err)
	}
	return nil
}

// reach finds an organisation the asker can see, answering the same for one
// they cannot reach as for one that is not there.
func (s *Service) reach(ctx context.Context, who httpapi.Actor, organisationSlug string) (Organisation, policy.Actor, error) {
	actor, err := s.actor(ctx, who)
	if err != nil {
		return Organisation{}, policy.Actor{}, err
	}

	organisation, err := s.store.OrganisationBySlug(ctx, organisationSlug)
	if errors.Is(err, ErrNotFound) {
		return Organisation{}, policy.Actor{}, httpapi.ErrMissing
	}
	if err != nil {
		return Organisation{}, policy.Actor{}, err
	}

	if _, member := actor.Organisations[organisation.ID.String()]; !member {
		if !s.holdsAnyProjectIn(ctx, actor, organisation) {
			return Organisation{}, policy.Actor{}, httpapi.ErrMissing
		}
	}

	return organisation, actor, nil
}

func (s *Service) holdsAnyProjectIn(ctx context.Context, actor policy.Actor, organisation Organisation) bool {
	projects, err := s.store.ProjectsIn(ctx, organisation.ID)
	if err != nil {
		return false
	}
	for _, project := range projects {
		if _, held := actor.Projects[project.ID.String()]; held {
			return true
		}
	}
	return false
}

func (s *Service) reachProject(
	ctx context.Context,
	who httpapi.Actor,
	organisationSlug, projectSlug string,
	action policy.Action,
) (Project, policy.Actor, policy.Resource, error) {
	actor, err := s.actor(ctx, who)
	if err != nil {
		return Project{}, policy.Actor{}, policy.Resource{}, err
	}

	project, err := s.store.ProjectBySlug(ctx, organisationSlug, projectSlug)
	if errors.Is(err, ErrNotFound) {
		return Project{}, policy.Actor{}, policy.Resource{}, httpapi.ErrMissing
	}
	if err != nil {
		return Project{}, policy.Actor{}, policy.Resource{}, err
	}

	resource := policy.Project(project.OrganisationID.String(), project.ID.String())

	// Something the asker cannot reach at all looks like something that is
	// not there; something they can see but may not act on is refused.
	if !policy.Can(actor, policy.ActionReadContent, resource) {
		return Project{}, policy.Actor{}, policy.Resource{}, httpapi.ErrMissing
	}
	if !policy.Can(actor, action, resource) {
		return Project{}, policy.Actor{}, policy.Resource{}, httpapi.ErrNotAllowed
	}

	return project, actor, resource, nil
}

func projectView(project Project, actor policy.Actor, resource policy.Resource) httpapi.ProjectView {
	role, _ := policy.EffectiveRole(actor, resource)

	may := policy.Allowed(actor, resource)
	names := make([]string, 0, len(may))
	for _, action := range may {
		names = append(names, string(action))
	}

	return httpapi.ProjectView{
		Slug:     project.Slug,
		Name:     project.Name,
		Domain:   project.Domain,
		YourRole: string(role),
		YouMay:   names,
	}
}

func memberView(member Member) httpapi.MemberView {
	return httpapi.MemberView{
		PersonID:      member.PersonID.String(),
		Email:         member.Email,
		DisplayName:   member.DisplayName,
		Role:          string(member.Role),
		StillToAccept: member.StillToAccept,
	}
}

// refusal turns a store failure into the answer the HTTP layer gives. A
// refusal carries its reason, because the reason is worth showing: an address
// that is taken, or a removal that would leave nobody in charge.
func refusal(err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return httpapi.ErrMissing

	case errors.Is(err, ErrUnusableSlug),
		errors.Is(err, ErrSlugTaken),
		errors.Is(err, ErrWouldLeaveNobodyInCharge),
		errors.Is(err, ErrNotARole),
		errors.Is(err, ErrNoAddress):
		return &httpapi.RefusedError{Reason: err.Error()}
	}
	return err
}

func mustParse(id string) uuid.UUID {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil
	}
	return parsed
}
