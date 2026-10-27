package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/policy"
	"github.com/quiqr/chatcms/internal/tenancy"
)

// Tenancy is what the service needs from the organisations it shapes. It is
// an interface so that the tests can watch what was asked for without a
// database.
type Tenancy interface {
	ActorFor(ctx context.Context, personID uuid.UUID) (policy.Actor, error)
	OrganisationMembers(ctx context.Context, organisationID uuid.UUID) ([]tenancy.OrganisationMember, error)
	ProjectsIn(ctx context.Context, organisationID uuid.UUID) ([]tenancy.Project, error)
	RemoveOrganisation(ctx context.Context, organisationID uuid.UUID) error
	CreateOrganisation(ctx context.Context, slug, name string, ownerID uuid.UUID) (tenancy.Organisation, error)
	OrganisationBySlug(ctx context.Context, slug string) (tenancy.Organisation, error)
	SetOrganisationRole(ctx context.Context, organisationID, personID uuid.UUID, role policy.OrganisationRole) error
	RemoveFromOrganisation(ctx context.Context, organisationID, personID uuid.UUID) error
}

// Service is every act that needs administration of the installation. Each
// one asks the policy first and writes down what it did afterwards.
type Service struct {
	store    *Store
	tenancy  Tenancy
	accounts Accounts
}

// Accounts is what administering needs from identity: an account can be
// stopped and started. Named here rather than imported whole so that admin
// depends on the little of auth it uses.
type Accounts interface {
	TurnAccountOff(ctx context.Context, who httpapi.Actor, personID uuid.UUID) error
	TurnAccountOn(ctx context.Context, personID uuid.UUID) error
}

func NewService(store *Store, tenancy Tenancy, accounts Accounts) *Service {
	return &Service{store: store, tenancy: tenancy, accounts: accounts}
}

// mayAdminister is the one gate. Every method here goes through it, and the
// resource is empty because an action over the installation is not held over
// anything within it.
func (s *Service) mayAdminister(ctx context.Context, who httpapi.Actor, action policy.Action) (policy.Actor, error) {
	if who.Anonymous || who.ID == "" {
		return policy.Nobody(), httpapi.ErrNotAllowed
	}
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return policy.Nobody(), httpapi.ErrNotAllowed
	}

	actor, err := s.tenancy.ActorFor(ctx, personID)
	if err != nil {
		return policy.Nobody(), err
	}
	if !policy.Can(actor, action, policy.Resource{}) {
		return policy.Nobody(), httpapi.ErrNotAllowed
	}
	return actor, nil
}

// MakeOrganisation brings one into being and gives it an owner, who is
// somebody else unless the administrator deliberately names themselves.
func (s *Service) MakeOrganisation(ctx context.Context, who httpapi.Actor, slug, name, ownerEmail string) (httpapi.OrganisationView, error) {
	if _, err := s.mayAdminister(ctx, who, policy.ActionCreateOrganisation); err != nil {
		return httpapi.OrganisationView{}, err
	}

	owner, err := s.personByEmail(ctx, ownerEmail)
	if err != nil {
		return httpapi.OrganisationView{}, err
	}

	organisation, err := s.tenancy.CreateOrganisation(ctx, slug, name, owner.id)
	if err != nil {
		return httpapi.OrganisationView{}, refusal(err)
	}

	s.record(ctx, who, MadeAnOrganisation, organisation.ID.String(), organisation.Name, map[string]any{
		"ownedBy": owner.email,
	})
	return httpapi.OrganisationView{Slug: organisation.Slug, Name: organisation.Name}, nil
}

// PutSomebodyInto gives a person a role in an organisation, which is how an
// administrator hands one over without joining it themselves.
func (s *Service) PutSomebodyInto(ctx context.Context, who httpapi.Actor, slug, email string, role policy.OrganisationRole) error {
	if _, err := s.mayAdminister(ctx, who, policy.ActionManageOrganisationPeople); err != nil {
		return err
	}

	organisation, err := s.tenancy.OrganisationBySlug(ctx, slug)
	if err != nil {
		return err
	}
	person, err := s.personByEmail(ctx, email)
	if err != nil {
		return err
	}

	if err := s.tenancy.SetOrganisationRole(ctx, organisation.ID, person.id, role); err != nil {
		return refusal(err)
	}

	s.record(ctx, who, PutSomebodyIn, organisation.ID.String(), organisation.Name, map[string]any{
		"person": person.email,
		"as":     string(role),
	})
	return nil
}

// TakeSomebodyOut removes a person from an organisation.
func (s *Service) TakeSomebodyOut(ctx context.Context, who httpapi.Actor, slug, email string) error {
	if _, err := s.mayAdminister(ctx, who, policy.ActionManageOrganisationPeople); err != nil {
		return err
	}

	organisation, err := s.tenancy.OrganisationBySlug(ctx, slug)
	if err != nil {
		return err
	}
	person, err := s.personByEmail(ctx, email)
	if err != nil {
		return err
	}

	if err := s.tenancy.RemoveFromOrganisation(ctx, organisation.ID, person.id); err != nil {
		return refusal(err)
	}

	s.record(ctx, who, TookSomebodyOut, organisation.ID.String(), organisation.Name, map[string]any{
		"person": person.email,
	})
	return nil
}

// Appoint makes somebody else an administrator.
func (s *Service) Appoint(ctx context.Context, who httpapi.Actor, email string) error {
	if _, err := s.mayAdminister(ctx, who, policy.ActionAdministerInstallation); err != nil {
		return err
	}

	person, err := s.personByEmail(ctx, email)
	if err != nil {
		return err
	}
	if err := s.store.MakeAdministrator(ctx, person.id); err != nil {
		return err
	}

	s.record(ctx, who, AppointedAnAdmin, person.id.String(), person.email, nil)
	return nil
}

// StandDown takes administration away, keeping the last one.
func (s *Service) StandDown(ctx context.Context, who httpapi.Actor, email string) error {
	if _, err := s.mayAdminister(ctx, who, policy.ActionAdministerInstallation); err != nil {
		return err
	}

	person, err := s.personByEmail(ctx, email)
	if err != nil {
		return err
	}
	if err := s.store.StopAdministrating(ctx, person.id); err != nil {
		if errors.Is(err, ErrLastAdministrator) {
			return &httpapi.RefusedError{Reason: ErrLastAdministrator.Error()}
		}
		return err
	}

	s.record(ctx, who, StoodDownAnAdmin, person.id.String(), person.email, nil)
	return nil
}

// TurnAccountOff stops an account being used anywhere on this installation.
// It belongs here rather than with an organisation's people because an
// account is held by the installation: somebody in two organisations would
// otherwise be locked out of one by the other.
func (s *Service) TurnAccountOff(ctx context.Context, who httpapi.Actor, email string) error {
	if _, err := s.mayAdminister(ctx, who, policy.ActionAdministerInstallation); err != nil {
		return err
	}

	person, err := s.personByEmail(ctx, email)
	if err != nil {
		return err
	}
	if err := s.accounts.TurnAccountOff(ctx, who, person.id); err != nil {
		return err
	}

	s.record(ctx, who, TurnedAnAccountOff, person.id.String(), person.email, nil)
	return nil
}

// TurnAccountOn lets it be used again, with the password it had.
func (s *Service) TurnAccountOn(ctx context.Context, who httpapi.Actor, email string) error {
	if _, err := s.mayAdminister(ctx, who, policy.ActionAdministerInstallation); err != nil {
		return err
	}

	person, err := s.personByEmail(ctx, email)
	if err != nil {
		return err
	}
	if err := s.accounts.TurnAccountOn(ctx, person.id); err != nil {
		return err
	}

	s.record(ctx, who, TurnedAnAccountOn, person.id.String(), person.email, nil)
	return nil
}

// Records is what has been done.
func (s *Service) Records(ctx context.Context, who httpapi.Actor, most int32) ([]Record, error) {
	if _, err := s.mayAdminister(ctx, who, policy.ActionAdministerInstallation); err != nil {
		return nil, err
	}
	return s.store.Records(ctx, most)
}

// record writes what happened. A failure to write it is logged by the caller
// of the service rather than failing the act itself: the organisation has
// already been made, and refusing afterwards would be a lie.
func (s *Service) record(ctx context.Context, who httpapi.Actor, did What, subject, subjectSaid string, detail map[string]any) {
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return
	}
	said := who.DisplayName
	if said == "" {
		said = who.Email
	}
	_ = s.store.Write(ctx, personID, said, did, subject, subjectSaid, detail)
}

type person struct {
	id    uuid.UUID
	email string
}

func (s *Service) personByEmail(ctx context.Context, email string) (person, error) {
	trimmed := strings.TrimSpace(email)
	if trimmed == "" {
		return person{}, &httpapi.RefusedError{Reason: "an email address is needed"}
	}

	row, err := s.store.db.Queries.PersonByEmail(ctx, trimmed)
	if err != nil {
		if isNoRows(err) {
			return person{}, &httpapi.RefusedError{Reason: ErrNoSuchPerson.Error()}
		}
		return person{}, fmt.Errorf("looking for that person failed: %w", err)
	}
	return person{id: uuid.UUID(row.ID.Bytes), email: row.Email}, nil
}

func refusal(err error) error {
	var refused *httpapi.RefusedError
	if errors.As(err, &refused) {
		return err
	}
	if errors.Is(err, tenancy.ErrSlugTaken) {
		return &httpapi.RefusedError{Reason: "that address is already taken"}
	}
	return err
}

// Organisation is one organisation with everything about it, which is what
// an administrator opens.
func (s *Service) Organisation(ctx context.Context, who httpapi.Actor, slug string) (httpapi.OrganisationWholeView, error) {
	if _, err := s.mayAdminister(ctx, who, policy.ActionSeeEveryOrganisation); err != nil {
		return httpapi.OrganisationWholeView{}, err
	}

	organisation, err := s.tenancy.OrganisationBySlug(ctx, slug)
	if err != nil {
		return httpapi.OrganisationWholeView{}, err
	}

	people, err := s.tenancy.OrganisationMembers(ctx, organisation.ID)
	if err != nil {
		return httpapi.OrganisationWholeView{}, err
	}
	projects, err := s.tenancy.ProjectsIn(ctx, organisation.ID)
	if err != nil {
		return httpapi.OrganisationWholeView{}, err
	}

	whole := httpapi.OrganisationWholeView{
		Slug:     organisation.Slug,
		Name:     organisation.Name,
		People:   make([]httpapi.MemberView, 0, len(people)),
		Projects: make([]httpapi.ProjectView, 0, len(projects)),
	}
	for _, person := range people {
		whole.People = append(whole.People, httpapi.MemberView{
			PersonID:      person.PersonID.String(),
			Email:         person.Email,
			DisplayName:   person.DisplayName,
			Role:          string(person.Role),
			StillToAccept: person.StillToAccept,
		})
	}
	for _, project := range projects {
		whole.Projects = append(whole.Projects, httpapi.ProjectView{
			Slug:   project.Slug,
			Name:   project.Name,
			Domain: project.Domain,
		})
	}
	return whole, nil
}

// RemoveOrganisation forgets one. It is refused while the organisation still
// holds projects: removing one with sites in it would make several
// customers' work unreachable through a single confirmation.
func (s *Service) RemoveOrganisation(ctx context.Context, who httpapi.Actor, slug string) error {
	if _, err := s.mayAdminister(ctx, who, policy.ActionRemoveOrganisation); err != nil {
		return err
	}

	organisation, err := s.tenancy.OrganisationBySlug(ctx, slug)
	if err != nil {
		return err
	}

	projects, err := s.tenancy.ProjectsIn(ctx, organisation.ID)
	if err != nil {
		return err
	}
	if len(projects) > 0 {
		names := make([]string, 0, len(projects))
		for _, project := range projects {
			names = append(names, project.Name)
		}
		return &httpapi.RefusedError{
			Reason: "this organisation still holds " + strings.Join(names, ", ") +
				"; remove them first",
		}
	}

	if err := s.tenancy.RemoveOrganisation(ctx, organisation.ID); err != nil {
		return err
	}

	s.record(ctx, who, RemovedAnOrganisation, organisation.ID.String(), organisation.Name, nil)
	return nil
}
