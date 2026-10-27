// Package tenancy holds organisations, the projects inside them, and who
// belongs to each. It stores roles and hands them to internal/policy, which
// is the only place that decides what a role may do.
package tenancy

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms/internal/policy"
	"github.com/quiqr/chatcms/internal/store"
)

var (
	// ErrNotFound says the thing does not exist, or belongs to someone the
	// asker is not. The two are one answer so that the API is not a directory
	// of every customer's projects.
	ErrNotFound = errors.New("there is no such organisation or project")

	// ErrUnusableSlug says an address cannot be used.
	ErrUnusableSlug = errors.New("an address may hold only lower-case letters, digits and single hyphens")

	// ErrSlugTaken says the address is already in use.
	ErrSlugTaken = errors.New("that address is already taken")

	// ErrWouldLeaveNobodyInCharge says a removal would leave nobody able to
	// administer what is left.
	ErrWouldLeaveNobodyInCharge = errors.New("this would leave nobody who can administer it")
)

var slugShape = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// Organisation is a customer or an agency.
type Organisation struct {
	ID   uuid.UUID
	Slug string
	Name string
}

// Project is one site.
type Project struct {
	ID             uuid.UUID
	OrganisationID uuid.UUID
	Slug           string
	Name           string
	Domain         string
}

// Member is someone on a project, with the role they hold.
type Member struct {
	PersonID    uuid.UUID
	Email       string
	DisplayName string
	Role        policy.Role
	// StillToAccept says they have been invited and have not set a password.
	StillToAccept bool
}

// Store reads and writes organisations, projects and memberships.
type Store struct {
	db *store.DB
}

func NewStore(db *store.DB) *Store {
	return &Store{db: db}
}

// CreateOrganisation makes an organisation with its creator as owner.
func (s *Store) CreateOrganisation(ctx context.Context, slug, name string, ownerID uuid.UUID) (Organisation, error) {
	if err := checkSlug(slug); err != nil {
		return Organisation{}, err
	}

	row, err := s.db.Queries.CreateOrganisation(ctx, store.CreateOrganisationParams{Slug: slug, Name: name})
	if err != nil {
		if isUniqueViolation(err) {
			return Organisation{}, ErrSlugTaken
		}
		return Organisation{}, fmt.Errorf("creating the organisation failed: %w", err)
	}

	if err := s.db.Queries.SetOrganisationMember(ctx, store.SetOrganisationMemberParams{
		OrganisationID: row.ID,
		PersonID:       pgUUID(ownerID),
		Role:           string(policy.OrganisationOwner),
	}); err != nil {
		return Organisation{}, fmt.Errorf("making the creator an owner failed: %w", err)
	}

	return organisationFrom(row), nil
}

// OrganisationBySlug finds one by its address.
func (s *Store) OrganisationBySlug(ctx context.Context, slug string) (Organisation, error) {
	row, err := s.db.Queries.GetOrganisationBySlug(ctx, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return Organisation{}, ErrNotFound
	}
	if err != nil {
		return Organisation{}, fmt.Errorf("reading the organisation failed: %w", err)
	}
	return organisationFrom(row), nil
}

// CreateProject makes a project in an organisation.
func (s *Store) CreateProject(ctx context.Context, organisationID uuid.UUID, slug, name, domain string) (Project, error) {
	if err := checkSlug(slug); err != nil {
		return Project{}, err
	}

	row, err := s.db.Queries.CreateProject(ctx, store.CreateProjectParams{
		OrganisationID: pgUUID(organisationID),
		Slug:           slug,
		Name:           name,
		Domain:         domain,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return Project{}, ErrSlugTaken
		}
		return Project{}, fmt.Errorf("creating the project failed: %w", err)
	}
	return projectFrom(row), nil
}

// ProjectBySlug finds a project by its organisation's address and its own.
func (s *Store) ProjectBySlug(ctx context.Context, organisationSlug, projectSlug string) (Project, error) {
	row, err := s.db.Queries.GetProjectBySlug(ctx, store.GetProjectBySlugParams{
		Slug: organisationSlug, Slug_2: projectSlug,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Project{}, ErrNotFound
	}
	if err != nil {
		return Project{}, fmt.Errorf("reading the project failed: %w", err)
	}
	return projectFrom(row), nil
}

// ProjectsIn lists every project of an organisation. The caller filters by
// what the asker may reach.
func (s *Store) ProjectsIn(ctx context.Context, organisationID uuid.UUID) ([]Project, error) {
	rows, err := s.db.Queries.ListProjectsInOrganisation(ctx, pgUUID(organisationID))
	if err != nil {
		return nil, fmt.Errorf("listing the projects failed: %w", err)
	}

	projects := make([]Project, 0, len(rows))
	for _, row := range rows {
		projects = append(projects, projectFrom(row))
	}
	return projects, nil
}

// OrganisationsFor lists the organisations a person can reach, whether
// through membership of the organisation or a role on one of its projects.
func (s *Store) OrganisationsFor(ctx context.Context, personID uuid.UUID) ([]Organisation, error) {
	rows, err := s.db.Queries.ListOrganisationsForPerson(ctx, pgUUID(personID))
	if err != nil {
		return nil, fmt.Errorf("listing the organisations failed: %w", err)
	}

	out := make([]Organisation, 0, len(rows))
	for _, row := range rows {
		out = append(out, organisationFrom(row))
	}
	return out, nil
}

// ActorFor loads everything the policy needs to answer for this person.
func (s *Store) ActorFor(ctx context.Context, personID uuid.UUID) (policy.Actor, error) {
	actor := policy.Actor{
		PersonID:      personID.String(),
		Organisations: map[string]policy.OrganisationRole{},
		Projects:      map[string]policy.Role{},
	}

	// Administering is read from the account rather than from a membership,
	// so it is loaded here with everything else the policy needs.
	person, err := s.db.Queries.GetPerson(ctx, pgUUID(personID))
	if err != nil {
		return policy.Actor{}, fmt.Errorf("reading the person failed: %w", err)
	}
	actor.Administers = person.Administers

	organisations, err := s.db.Queries.ListOrganisationMembershipsForPerson(ctx, pgUUID(personID))
	if err != nil {
		return policy.Actor{}, fmt.Errorf("reading the memberships failed: %w", err)
	}
	for _, row := range organisations {
		actor.Organisations[uuid.UUID(row.OrganisationID.Bytes).String()] = policy.OrganisationRole(row.Role)
	}

	projects, err := s.db.Queries.ListProjectMembershipsForPerson(ctx, pgUUID(personID))
	if err != nil {
		return policy.Actor{}, fmt.Errorf("reading the memberships failed: %w", err)
	}
	for _, row := range projects {
		actor.Projects[uuid.UUID(row.ProjectID.Bytes).String()] = policy.Role(row.Role)
	}

	return actor, nil
}

// SetProjectRole gives someone a role on a project, replacing any they hold.
func (s *Store) SetProjectRole(ctx context.Context, projectID, personID uuid.UUID, role policy.Role) error {
	return s.db.Queries.SetProjectMember(ctx, store.SetProjectMemberParams{
		ProjectID: pgUUID(projectID),
		PersonID:  pgUUID(personID),
		Role:      string(role),
	})
}

// RemoveFromProject takes someone off a project, refusing when it would leave
// nobody able to administer it.
func (s *Store) RemoveFromProject(ctx context.Context, projectID, personID uuid.UUID) error {
	members, err := s.Members(ctx, projectID)
	if err != nil {
		return err
	}

	remaining := 0
	for _, member := range members {
		if member.PersonID == personID {
			continue
		}
		if member.Role == policy.RoleOwner || member.Role == policy.RoleImplementer {
			remaining++
		}
	}
	if remaining == 0 {
		return ErrWouldLeaveNobodyInCharge
	}

	return s.db.Queries.RemoveProjectMember(ctx, store.RemoveProjectMemberParams{
		ProjectID: pgUUID(projectID),
		PersonID:  pgUUID(personID),
	})
}

// RemoveFromOrganisation takes someone out, refusing when it would leave no
// owner.
func (s *Store) RemoveFromOrganisation(ctx context.Context, organisationID, personID uuid.UUID) error {
	owners, err := s.db.Queries.CountOrganisationOwners(ctx, pgUUID(organisationID))
	if err != nil {
		return fmt.Errorf("counting the owners failed: %w", err)
	}

	memberships, err := s.db.Queries.ListOrganisationMembershipsForPerson(ctx, pgUUID(personID))
	if err != nil {
		return fmt.Errorf("reading the memberships failed: %w", err)
	}

	for _, membership := range memberships {
		if uuid.UUID(membership.OrganisationID.Bytes) != organisationID {
			continue
		}
		if policy.OrganisationRole(membership.Role) == policy.OrganisationOwner && owners <= 1 {
			return ErrWouldLeaveNobodyInCharge
		}
	}

	return s.db.Queries.RemoveOrganisationMember(ctx, store.RemoveOrganisationMemberParams{
		OrganisationID: pgUUID(organisationID),
		PersonID:       pgUUID(personID),
	})
}

// SetOrganisationRole gives someone a role on an organisation.
func (s *Store) SetOrganisationRole(ctx context.Context, organisationID, personID uuid.UUID, role policy.OrganisationRole) error {
	return s.db.Queries.SetOrganisationMember(ctx, store.SetOrganisationMemberParams{
		OrganisationID: pgUUID(organisationID),
		PersonID:       pgUUID(personID),
		Role:           string(role),
	})
}

// Members lists the people on a project.
func (s *Store) Members(ctx context.Context, projectID uuid.UUID) ([]Member, error) {
	rows, err := s.db.Queries.ListProjectMembers(ctx, pgUUID(projectID))
	if err != nil {
		return nil, fmt.Errorf("listing the people failed: %w", err)
	}

	members := make([]Member, 0, len(rows))
	for _, row := range rows {
		members = append(members, Member{
			PersonID:      uuid.UUID(row.ID.Bytes),
			Email:         row.Email,
			DisplayName:   row.DisplayName,
			Role:          policy.Role(row.Role),
			StillToAccept: !row.HasPassword,
		})
	}
	return members, nil
}

func checkSlug(slug string) error {
	if !slugShape.MatchString(strings.TrimSpace(slug)) {
		return fmt.Errorf("%w: %q", ErrUnusableSlug, slug)
	}
	return nil
}

func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "SQLSTATE 23505")
}

func organisationFrom(row store.Organisation) Organisation {
	return Organisation{ID: uuid.UUID(row.ID.Bytes), Slug: row.Slug, Name: row.Name}
}

func projectFrom(row store.Project) Project {
	return Project{
		ID:             uuid.UUID(row.ID.Bytes),
		OrganisationID: uuid.UUID(row.OrganisationID.Bytes),
		Slug:           row.Slug,
		Name:           row.Name,
		Domain:         row.Domain,
	}
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

// EveryOrganisation is all of them, which only an administrator is shown.
func (s *Store) EveryOrganisation(ctx context.Context) ([]Organisation, error) {
	rows, err := s.db.Queries.ListEveryOrganisation(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading the organisations failed: %w", err)
	}

	out := make([]Organisation, 0, len(rows))
	for _, row := range rows {
		out = append(out, Organisation{
			ID:   uuid.UUID(row.ID.Bytes),
			Slug: row.Slug,
			Name: row.Name,
		})
	}
	return out, nil
}

// RenameOrganisation changes the name and leaves the address alone.
func (s *Store) RenameOrganisation(ctx context.Context, organisationID uuid.UUID, name string) (Organisation, error) {
	row, err := s.db.Queries.RenameOrganisation(ctx, store.RenameOrganisationParams{
		ID:   pgUUID(organisationID),
		Name: name,
	})
	if err != nil {
		return Organisation{}, fmt.Errorf("renaming the organisation failed: %w", err)
	}
	return organisationFrom(row), nil
}

// OrganisationMember is somebody in an organisation. It is its own type
// rather than Member, because what somebody holds on an organisation is not
// the same thing as what they hold on a project, and one type for both would
// say they were.
type OrganisationMember struct {
	PersonID    uuid.UUID
	Email       string
	DisplayName string
	Role        policy.OrganisationRole
	// StillToAccept says they have been invited and have not set a password.
	StillToAccept bool
}

// OrganisationMembers is who belongs to one, and what each holds there.
func (s *Store) OrganisationMembers(ctx context.Context, organisationID uuid.UUID) ([]OrganisationMember, error) {
	rows, err := s.db.Queries.ListOrganisationMembers(ctx, pgUUID(organisationID))
	if err != nil {
		return nil, fmt.Errorf("reading who is in the organisation failed: %w", err)
	}

	out := make([]OrganisationMember, 0, len(rows))
	for _, row := range rows {
		out = append(out, OrganisationMember{
			PersonID:      uuid.UUID(row.PersonID.Bytes),
			Email:         row.Email,
			DisplayName:   row.DisplayName,
			Role:          policy.OrganisationRole(row.Role),
			StillToAccept: row.PasswordHash == nil,
		})
	}
	return out, nil
}

// UpdateProject changes what a project is called and the address shown to
// customers. Its own address is not among them: links already sent are
// written against it.
func (s *Store) UpdateProject(ctx context.Context, projectID uuid.UUID, name, domain string) (Project, error) {
	row, err := s.db.Queries.UpdateProject(ctx, store.UpdateProjectParams{
		ID:     pgUUID(projectID),
		Name:   name,
		Domain: domain,
	})
	if err != nil {
		return Project{}, fmt.Errorf("changing the project failed: %w", err)
	}
	return projectFrom(row), nil
}

// RemoveProject forgets a project. What is in git is the customer's and is
// not touched.
func (s *Store) RemoveProject(ctx context.Context, projectID uuid.UUID) error {
	if err := s.db.Queries.RemoveProject(ctx, pgUUID(projectID)); err != nil {
		return fmt.Errorf("removing the project failed: %w", err)
	}
	return nil
}
