package tenancy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/auth"
	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/policy"
)

// Service answers the HTTP layer's questions about organisations, projects
// and people. It loads what the policy needs and asks it; it never decides.
type Service struct {
	copies  Copies
	zips    Copies
	store   *Store
	inviter *Inviter
	// waiting answers how much is waiting to be published. It is injected
	// because tenancy is about who may reach what, and opening a site to
	// count versions is somebody else's job. Without it the count is zero,
	// which is what a project with no site has.
	waiting Waiting
	// preferences answers what somebody chose about how the application
	// behaves, which rides along in the one answer a screen asks for.
	preferences Preferences
}

// Preferences is the little of identity that tenancy needs.
type Preferences interface {
	AppearanceFor(ctx context.Context, personID uuid.UUID) (string, error)
	LanguageFor(ctx context.Context, personID uuid.UUID) (string, error)
}

// KnowsWhatTheyChose gives the service somewhere to ask.
func (s *Service) KnowsWhatTheyChose(preferences Preferences) { s.preferences = preferences }

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

	actor, err := s.actor(ctx, who)
	if err != nil {
		return nil, err
	}

	views := make([]httpapi.OrganisationView, 0, len(organisations))
	for _, organisation := range organisations {
		views = append(views, httpapi.OrganisationView{
			Slug: organisation.Slug,
			Name: organisation.Name,
			// What decides whether a way into the organisation's own screens
			// is offered. Asked of the policy rather than of the role, so
			// that a platform admin who administers nothing is not offered
			// one either.
			YouAdminister: policy.Can(actor, policy.ActionManagePeople,
				policy.Organisation(organisation.ID.String())),
		})
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

	// Counted and made together, so that two people making a site at the
	// same moment cannot both take the last one.
	var project Project
	if err := s.store.WithRoomForAProject(ctx, organisation.ID, func(within *Store) error {
		made, err := within.CreateProject(ctx, organisation.ID, slug, name, domain)
		if err != nil {
			return err
		}
		project = made
		return nil
	}); err != nil {
		return httpapi.ProjectView{}, refusal(err)
	}

	// Whoever made it reaches it through the organisation they administer;
	// giving the row as well keeps the project's own list of people honest.
	if err := s.store.SetProjectRole(ctx, project.ID, mustParse(who.ID), policy.RoleImplementer); err != nil {
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
		return httpapi.MemberView{}, &httpapi.RefusedError{
			Key:    "refused.not-a-role",
			Reason: ErrNotARole.Error(),
		}
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
		Archived: project.Archived,
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
		views = append(views, httpapi.OrganisationView{
			Slug: organisation.Slug,
			Name: organisation.Name,
			// What decides whether a way into the organisation's own screens
			// is offered. Asked of the policy rather than of the role, so
			// that a platform admin who administers nothing is not offered
			// one either.
			YouAdminister: policy.Can(actor, policy.ActionManagePeople,
				policy.Organisation(organisation.ID.String())),
		})
	}

	answer := httpapi.WhoAmIView{
		Person: httpapi.PersonView{
			ID:          who.ID,
			Email:       who.Email,
			DisplayName: who.DisplayName,
		},
		Administers:   policy.Can(actor, policy.ActionAdministerInstallation, policy.Resource{}),
		Organisations: views,
		Appearance:    s.appearanceFor(ctx, personID),
		Language:      s.languageFor(ctx, personID),
	}
	if who.ActingAs() {
		answer.ActingAs = &httpapi.ActingAsView{
			ReallyEmail: who.Really.Email,
			ReallyName:  who.Really.Said(),
		}
	}
	return answer, nil
}

// languageFor is what this person chose. A preference that cannot be read
// is not a reason to refuse a screen; the browser decides instead.
func (s *Service) languageFor(ctx context.Context, personID uuid.UUID) string {
	if s.preferences == nil {
		return ""
	}
	chosen, err := s.preferences.LanguageFor(ctx, personID)
	if err != nil {
		return ""
	}
	return chosen
}

// appearanceFor is what this person chose. A preference that cannot be read
// is not a reason to refuse a screen; the machine decides instead.
func (s *Service) appearanceFor(ctx context.Context, personID uuid.UUID) string {
	if s.preferences == nil {
		return ""
	}
	chosen, err := s.preferences.AppearanceFor(ctx, personID)
	if err != nil {
		return ""
	}
	return chosen
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
		return httpapi.OrganisationView{}, &httpapi.RefusedError{Key: "refused.organisation-needs-name", Reason: "an organisation needs a name"}
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
			LastSeen:      s.lastSeen(ctx, person.PersonID),
		})
	}
	return views, nil
}

// lastSeen is when somebody was last here, or nothing when they never were.
// A screen that says "never" about somebody who has not accepted yet is
// saying the same thing twice; it says so only for somebody who has.
func (s *Service) lastSeen(ctx context.Context, personID uuid.UUID) *time.Time {
	when, ever, err := s.inviter.auth.People().LastSeen(ctx, personID)
	if err != nil || !ever {
		return nil
	}
	return &when
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
		return httpapi.ProjectView{}, &httpapi.RefusedError{Key: "refused.project-needs-name", Reason: "a project needs a name"}
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

// Copies is what the service needs to make one site from another. It is an
// interface so that the service does not have to know how a working copy is
// kept, and so that a test can watch what was asked for.
type Copies interface {
	CopyTo(ctx context.Context, from, to uuid.UUID, remote string, author gitsync.Identity, message string) error
	Zip(ctx context.Context, projectID uuid.UUID, into io.Writer) error
}

// MakesCopies gives the service somewhere to ask.
func (s *Service) MakesCopies(copies Copies) {
	s.copies = copies
	s.zips = copies
}

// Duplicate makes a separate copy of a site in the same organisation.
//
// Three things are copied always, because they are what a site is: its
// pages, its media and its model. Two are copied only if asked for: the
// people, and where it publishes. Publishing is off by default because a
// copy that published where the original publishes would overwrite a live
// website the first time anybody pressed Publish.
func (s *Service) Duplicate(ctx context.Context, who httpapi.Actor, organisation, project, slug, name string, withPeople, withPublishing bool) (httpapi.ProjectView, error) {
	original, actor, _, err := s.reachProject(ctx, who, organisation, project, policy.ActionSeeTechnicalDetail)
	if err != nil {
		return httpapi.ProjectView{}, err
	}
	organisationID := original.OrganisationID
	if !policy.Can(actor, policy.ActionCreateProject, policy.Organisation(organisationID.String())) {
		return httpapi.ProjectView{}, httpapi.ErrNotAllowed
	}
	if strings.TrimSpace(name) == strings.TrimSpace(original.Name) {
		return httpapi.ProjectView{}, &httpapi.RefusedError{
			Key:    "refused.copy-needs-its-own-name",
			Reason: "that name is already taken; give the copy its own name",
		}
	}

	var copy Project
	if err := s.store.WithRoomForAProject(ctx, organisationID, func(within *Store) error {
		made, err := within.CreateProject(ctx, organisationID, slug, name, "")
		if err != nil {
			return err
		}
		copy = made
		return nil
	}); err != nil {
		return httpapi.ProjectView{}, refusal(err)
	}

	// Where the copy publishes, only if somebody asked for it. Without it
	// the copy has no remote at all, which is what makes pressing Publish
	// on a copy harmless until somebody says where it should go.
	remote := ""
	if withPublishing {
		row, err := s.store.db.Queries.GetProject(ctx, pgUUID(original.ID))
		if err != nil {
			return httpapi.ProjectView{}, fmt.Errorf("reading where the site publishes failed: %w", err)
		}
		remote = row.Remote
	}

	// The copy's first version is attributed to whoever asked for it,
	// which is what the history of the copy should say.
	author := gitsync.Identity{Name: who.Said(), Email: who.Email}
	if err := s.copies.CopyTo(ctx, original.ID, copy.ID, remote, author,
		"Copied from "+original.Name); err != nil {
		// The row would otherwise be a site with no files behind it.
		_ = s.store.RemoveProject(ctx, copy.ID)
		return httpapi.ProjectView{}, err
	}

	if withPeople {
		if err := s.store.CopyThePeople(ctx, original.ID, copy.ID); err != nil {
			return httpapi.ProjectView{}, err
		}
	} else if err := s.store.SetProjectRole(ctx, copy.ID, mustParse(who.ID), policy.RoleImplementer); err != nil {
		return httpapi.ProjectView{}, err
	}

	actor, err = s.actor(ctx, who)
	if err != nil {
		return httpapi.ProjectView{}, err
	}
	return projectView(copy, actor, policy.Project(organisationID.String(), copy.ID.String())), nil
}

// MoveTo puts a site in another organisation. Whoever does it administers
// both, or administers the installation.
func (s *Service) MoveTo(ctx context.Context, who httpapi.Actor, organisation, project, destination string) error {
	found, actor, _, err := s.reachProject(ctx, who, organisation, project, policy.ActionSeeTechnicalDetail)
	if err != nil {
		return err
	}
	from := found.OrganisationID

	to, err := s.store.OrganisationBySlug(ctx, destination)
	if err != nil {
		return err
	}
	if to.ID == from {
		return &httpapi.RefusedError{
			Key:    "refused.already-there",
			Reason: "that site is already in that organisation",
		}
	}

	if !policy.MayMoveTo(actor, policy.Organisation(from.String()), policy.Organisation(to.ID.String())) {
		return httpapi.ErrNotAllowed
	}

	return refusal(s.store.MoveProject(ctx, found.ID, to.ID))
}

// Archive stops a site counting against its organisation's limit of sites.
// Nothing is removed, and what is published stays published.
func (s *Service) Archive(ctx context.Context, who httpapi.Actor, organisation, project string, archived bool) error {
	found, _, _, err := s.reachProject(ctx, who, organisation, project, policy.ActionDeleteProject)
	if err != nil {
		return err
	}
	if archived {
		return s.store.Archive(ctx, found.ID)
	}
	return refusal(s.store.BringBack(ctx, found.ID))
}

// WriteZip writes the site's files and answers what the file should be
// called. Whoever may see a site's technical detail may take a copy of it
// away; an editor may read a page, which is not the same thing.
func (s *Service) WriteZip(ctx context.Context, who httpapi.Actor, organisation, project string, to io.Writer) (string, error) {
	found, _, _, err := s.reachProject(ctx, who, organisation, project, policy.ActionSeeTechnicalDetail)
	if err != nil {
		return "", err
	}
	if s.zips == nil {
		return "", fmt.Errorf("nothing here can make a zip")
	}
	if err := s.zips.Zip(ctx, found.ID, to); err != nil {
		return "", err
	}
	return found.Slug, nil
}

// InviteToOrganisation brings somebody into an organisation, which its owner
// manages rather than having to ask whoever runs the service.
func (s *Service) InviteToOrganisation(ctx context.Context, who httpapi.Actor, slug, email, role string) (httpapi.MemberView, error) {
	organisation, actor, err := s.reach(ctx, who, slug)
	if err != nil {
		return httpapi.MemberView{}, err
	}
	if !policy.Can(actor, policy.ActionManagePeople, policy.Organisation(organisation.ID.String())) {
		return httpapi.MemberView{}, httpapi.ErrNotAllowed
	}

	invitation, err := s.inviter.InviteToOrganisation(ctx, organisation.ID, email,
		policy.OrganisationRole(role))
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

// ChangeOrganisationRole changes what somebody holds in an organisation.
func (s *Service) ChangeOrganisationRole(ctx context.Context, who httpapi.Actor, slug, personID, role string) error {
	organisation, actor, err := s.reach(ctx, who, slug)
	if err != nil {
		return err
	}
	if !policy.Can(actor, policy.ActionManagePeople, policy.Organisation(organisation.ID.String())) {
		return httpapi.ErrNotAllowed
	}

	asked := policy.OrganisationRole(role)
	if asked != policy.OrganisationAdmin && asked != policy.OrganisationMember {
		return &httpapi.RefusedError{Key: "refused.not-role-organisation-has", Reason: "that is not a role an organisation has"}
	}

	person, err := uuid.Parse(personID)
	if err != nil {
		return httpapi.ErrMissing
	}
	return refusal(s.store.SetOrganisationRole(ctx, organisation.ID, person, asked))
}

// RemoveFromOrganisation takes somebody out of one.
func (s *Service) RemoveFromOrganisation(ctx context.Context, who httpapi.Actor, slug, personID string) error {
	organisation, actor, err := s.reach(ctx, who, slug)
	if err != nil {
		return err
	}
	if !policy.Can(actor, policy.ActionManagePeople, policy.Organisation(organisation.ID.String())) {
		return httpapi.ErrNotAllowed
	}

	person, err := uuid.Parse(personID)
	if err != nil {
		return httpapi.ErrMissing
	}
	return refusal(s.store.RemoveFromOrganisation(ctx, organisation.ID, person))
}

// SendTheInvitationAgain issues a new way in and stops the previous one.
func (s *Service) SendTheInvitationAgain(ctx context.Context, who httpapi.Actor, slug, personID string) error {
	organisation, actor, err := s.reach(ctx, who, slug)
	if err != nil {
		return err
	}
	if !policy.Can(actor, policy.ActionManagePeople, policy.Organisation(organisation.ID.String())) {
		return httpapi.ErrNotAllowed
	}

	person, err := uuid.Parse(personID)
	if err != nil {
		return httpapi.ErrMissing
	}
	return s.inviter.SendAgain(ctx, person)
}

// UsageOf is how full an organisation is, for anybody who belongs to it. It
// is not a secret from its own members: a screen that offers a New project
// button has to know whether there is room for one.
func (s *Service) UsageOf(ctx context.Context, who httpapi.Actor, organisationSlug string) (httpapi.UsageView, error) {
	organisation, _, err := s.reach(ctx, who, organisationSlug)
	if err != nil {
		return httpapi.UsageView{}, err
	}

	usage, err := s.store.UsageFor(ctx, organisation.ID)
	if err != nil {
		return httpapi.UsageView{}, err
	}

	return httpapi.UsageView{
		Members:  counted(usage.Members),
		Projects: counted(usage.Projects),
	}, nil
}

func counted(from Counted) httpapi.CountedView {
	return httpapi.CountedView{Held: from.Held, Allowed: from.Allowed, Full: from.Full()}
}

// ManageableOrganisation is one organisation as whoever administers it sees
// it. One answer rather than three, because every screen in the area needs
// the counts in its shell whichever one of them is open.
func (s *Service) ManageableOrganisation(ctx context.Context, who httpapi.Actor, slug string) (httpapi.OrganisationAreaView, error) {
	organisation, actor, err := s.reach(ctx, who, slug)
	if err != nil {
		return httpapi.OrganisationAreaView{}, err
	}

	// Managing the organisation is the same permission that invites somebody
	// to it; an organisation its own members cannot manage is still one they
	// can see the projects of.
	resource := policy.Organisation(organisation.ID.String())
	if !policy.Can(actor, policy.ActionManagePeople, resource) {
		return httpapi.OrganisationAreaView{}, httpapi.ErrNotAllowed
	}

	usage, err := s.store.UsageFor(ctx, organisation.ID)
	if err != nil {
		return httpapi.OrganisationAreaView{}, err
	}

	people, err := s.store.MembersWithTheirProjects(ctx, organisation.ID)
	if err != nil {
		return httpapi.OrganisationAreaView{}, err
	}

	projects, err := s.store.ProjectsIn(ctx, organisation.ID)
	if err != nil {
		return httpapi.OrganisationAreaView{}, err
	}

	area := httpapi.OrganisationAreaView{
		Slug:     organisation.Slug,
		Name:     organisation.Name,
		Usage:    httpapi.UsageView{Members: counted(usage.Members), Projects: counted(usage.Projects)},
		People:   make([]httpapi.OrganisationPerson, 0, len(people)),
		Projects: make([]httpapi.ProjectView, 0, len(projects)),
	}

	for _, person := range people {
		said := httpapi.OrganisationPerson{
			PersonID:      person.PersonID.String(),
			Email:         person.Email,
			DisplayName:   person.DisplayName,
			Role:          string(person.Role),
			StillToAccept: person.StillToAccept,
			Projects:      make([]httpapi.BelongingView, 0, len(person.Projects)),
		}
		if person.EverSeen {
			seen := person.LastSeen
			said.LastSeen = &seen
		}
		for _, project := range person.Projects {
			said.Projects = append(said.Projects, httpapi.BelongingView{
				Slug: project.Slug, Name: project.Name, Role: project.Role,
			})
		}
		area.People = append(area.People, said)
	}

	for _, project := range projects {
		view := projectView(project, actor, policy.Project(organisation.ID.String(), project.ID.String()))
		view.Unpublished = s.unpublishedIn(ctx, project)
		area.Projects = append(area.Projects, view)
	}

	return area, nil
}
