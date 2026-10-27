package tenancy

import (
	"context"
	"errors"
	"strings"

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
	// waiting answers how much is waiting to be published. It is injected
	// because tenancy is about who may reach what, and opening a site to
	// count versions is somebody else's job. Without it the count is zero,
	// which is what a project with no site has.
	waiting Waiting
}

// Waiting is how much of a project is waiting to go out. One source, so that
// the number on a button and the number on the publish screen cannot
// disagree, which is how they came to.
type Waiting interface {
	UnpublishedIn(ctx context.Context, projectID string) (int, error)
}

// KnowsWhatIsWaiting gives the service somewhere to ask.
func (s *Service) KnowsWhatIsWaiting(waiting Waiting) { s.waiting = waiting }

func (s *Service) unpublishedIn(ctx context.Context, project Project) int {
	if s.waiting == nil {
		return 0
	}
	count, err := s.waiting.UnpublishedIn(ctx, project.ID.String())
	if err != nil {
		// A site that cannot be read is not a reason to refuse the screen
		// that lists it; the publish screen says what is wrong.
		return 0
	}
	return count
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

	// Bringing an organisation into being is held over the installation, not
	// over anything in it, so the resource is empty and only an administrator
	// is answered yes.
	actor, err := s.actor(ctx, who)
	if err != nil {
		return httpapi.OrganisationView{}, err
	}
	if !policy.Can(actor, policy.ActionCreateOrganisation, policy.Resource{}) {
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
		view := projectView(project, actor, resource)
		view.Unpublished = s.unpublishedIn(ctx, project)
		views = append(views, view)
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

	view := projectView(project, actor, resource)
	view.Unpublished = s.unpublishedIn(ctx, project)
	return view, nil
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
		// Somebody who administers the installation reaches every
		// organisation, because managing them is what administering is for.
		// It still grants nothing over what is inside them.
		seesEvery := policy.Can(actor, policy.ActionSeeEveryOrganisation, policy.Resource{})
		if !seesEvery && !s.holdsAnyProjectIn(ctx, actor, organisation) {
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

// WhoAmI answers what a screen needs before it can draw its shell: who is
// looking, whether they administer the installation, and which organisations
// they may switch to.
func (s *Service) WhoAmI(ctx context.Context, who httpapi.Actor) (httpapi.WhoAmIView, error) {
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return httpapi.WhoAmIView{}, httpapi.ErrNotAllowed
	}

	actor, err := s.actor(ctx, who)
	if err != nil {
		return httpapi.WhoAmIView{}, err
	}

	organisations, err := s.store.OrganisationsFor(ctx, personID)
	if err != nil {
		return httpapi.WhoAmIView{}, err
	}

	// An administrator sees every organisation, because managing them is
	// what administering is for. They are not thereby in any of them.
	if policy.Can(actor, policy.ActionSeeEveryOrganisation, policy.Resource{}) {
		organisations, err = s.store.EveryOrganisation(ctx)
		if err != nil {
			return httpapi.WhoAmIView{}, err
		}
	}

	views := make([]httpapi.OrganisationView, 0, len(organisations))
	for _, organisation := range organisations {
		views = append(views, httpapi.OrganisationView{Slug: organisation.Slug, Name: organisation.Name})
	}

	return httpapi.WhoAmIView{
		Person: httpapi.PersonView{
			ID:          who.ID,
			Email:       who.Email,
			DisplayName: who.DisplayName,
		},
		Administers:   policy.Can(actor, policy.ActionAdministerInstallation, policy.Resource{}),
		Organisations: views,
	}, nil
}

// Rename gives an organisation a different name. Its address does not change
// with it, because links already written would stop working.
func (s *Service) Rename(ctx context.Context, who httpapi.Actor, slug, name string) (httpapi.OrganisationView, error) {
	actor, err := s.actor(ctx, who)
	if err != nil {
		return httpapi.OrganisationView{}, err
	}
	if !policy.Can(actor, policy.ActionRenameOrganisation, policy.Resource{}) {
		return httpapi.OrganisationView{}, httpapi.ErrNotAllowed
	}

	if strings.TrimSpace(name) == "" {
		return httpapi.OrganisationView{}, &httpapi.RefusedError{Reason: "an organisation needs a name"}
	}

	organisation, err := s.store.OrganisationBySlug(ctx, slug)
	if err != nil {
		return httpapi.OrganisationView{}, err
	}

	renamed, err := s.store.RenameOrganisation(ctx, organisation.ID, strings.TrimSpace(name))
	if err != nil {
		return httpapi.OrganisationView{}, err
	}
	return httpapi.OrganisationView{Slug: renamed.Slug, Name: renamed.Name}, nil
}

// OrganisationPeople is who belongs to an organisation, which its owner
// manages and anybody in it may see.
func (s *Service) OrganisationPeople(ctx context.Context, who httpapi.Actor, slug string) ([]httpapi.MemberView, error) {
	organisation, actor, err := s.reach(ctx, who, slug)
	if err != nil {
		return nil, err
	}

	// Seeing who else is here is part of belonging. Changing it is not.
	if !policy.Can(actor, policy.ActionReadContent, policy.Organisation(organisation.ID.String())) &&
		!policy.Can(actor, policy.ActionSeeEveryOrganisation, policy.Resource{}) {
		return nil, httpapi.ErrMissing
	}

	people, err := s.store.OrganisationMembers(ctx, organisation.ID)
	if err != nil {
		return nil, err
	}

	views := make([]httpapi.MemberView, 0, len(people))
	for _, person := range people {
		views = append(views, httpapi.MemberView{
			PersonID:      person.PersonID.String(),
			Email:         person.Email,
			DisplayName:   person.DisplayName,
			Role:          string(person.Role),
			StillToAccept: person.StillToAccept,
		})
	}
	return views, nil
}

// ChangeProject gives a project a different name and a different address to
// show customers. Its own address is not changed: every link already sent is
// written against it, and milestone 01 has no way to keep an old one working.
func (s *Service) ChangeProject(ctx context.Context, who httpapi.Actor, organisation, project, name, domain string) (httpapi.ProjectView, error) {
	found, actor, resource, err := s.reachProject(ctx, who, organisation, project, policy.ActionChangeSettings)
	if err != nil {
		return httpapi.ProjectView{}, err
	}

	if strings.TrimSpace(name) == "" {
		return httpapi.ProjectView{}, &httpapi.RefusedError{Reason: "a project needs a name"}
	}

	changed, err := s.store.UpdateProject(ctx, found.ID, strings.TrimSpace(name), strings.TrimSpace(domain))
	if err != nil {
		return httpapi.ProjectView{}, err
	}
	view := projectView(changed, actor, resource)
	view.Unpublished = s.unpublishedIn(ctx, changed)
	return view, nil
}

// RemoveProject forgets a project. What is in git is the customer's, and
// ChatCMS forgetting a project must never be a way to lose it.
func (s *Service) RemoveProject(ctx context.Context, who httpapi.Actor, organisation, project string) error {
	found, _, _, err := s.reachProject(ctx, who, organisation, project, policy.ActionDeleteProject)
	if err != nil {
		return err
	}
	return s.store.RemoveProject(ctx, found.ID)
}
