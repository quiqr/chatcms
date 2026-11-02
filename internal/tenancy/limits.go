package tenancy

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/policy"
	"github.com/quiqr/chatcms/internal/store"
)

// Usage is how full an organisation is: what it holds, and what it is
// allowed. Allowed is nil where there is no limit, which is what a new
// organisation has.
//
// Everything that shows how full an organisation is reads this, so that the
// organisation's own screens and the platform admin's cannot disagree.
type Usage struct {
	Members  Counted
	Projects Counted
}

// Counted is one of the two things an organisation can run out of.
type Counted struct {
	Held    int
	Allowed *int
}

// Limited says whether there is a limit at all.
func (c Counted) Limited() bool { return c.Allowed != nil }

// Full says whether one more would be one too many.
func (c Counted) Full() bool { return c.Allowed != nil && c.Held >= *c.Allowed }

// Left is how many more there is room for, and is meaningless with no limit.
func (c Counted) Left() int {
	if c.Allowed == nil {
		return 0
	}
	if left := *c.Allowed - c.Held; left > 0 {
		return left
	}
	return 0
}

// UsageFor reads how full an organisation is.
func (s *Store) UsageFor(ctx context.Context, organisationID uuid.UUID) (Usage, error) {
	row, err := s.db.Queries.OrganisationUsage(ctx, pgUUID(organisationID))
	if err != nil {
		return Usage{}, fmt.Errorf("reading how full the organisation is failed: %w", err)
	}

	return Usage{
		Members:  Counted{Held: int(row.Members), Allowed: asInt(row.MaxMembers)},
		Projects: Counted{Held: int(row.Projects), Allowed: asInt(row.MaxProjects)},
	}, nil
}

// SetLimits writes what an organisation is allowed. Nil is no limit. A limit
// below what is already held is accepted and removes nothing: it stops what
// is new until usage is under it again.
func (s *Store) SetLimits(ctx context.Context, organisationID uuid.UUID, members, projects *int) error {
	if (members != nil && *members < 1) || (projects != nil && *projects < 1) {
		return &httpapi.RefusedError{
			Key:    "refused.limit-below-one",
			Reason: "a limit is a number of at least one, or none at all",
		}
	}

	if _, err := s.db.Queries.SetOrganisationLimits(ctx, store.SetOrganisationLimitsParams{
		ID:          pgUUID(organisationID),
		MaxMembers:  asInt32(members),
		MaxProjects: asInt32(projects),
	}); err != nil {
		return fmt.Errorf("writing the limits failed: %w", err)
	}
	return nil
}

// WithRoomForAMember runs something that adds a member, refusing first when
// the seat would be one too many. The count and the act are in one
// transaction, so two people inviting at the same moment cannot both get the
// last seat.
func (s *Store) WithRoomForAMember(ctx context.Context, organisationID uuid.UUID, add func(*Store) error) error {
	return s.inOneGo(ctx, func(within *Store) error {
		if err := within.RoomForAMember(ctx, organisationID); err != nil {
			return err
		}
		return add(within)
	})
}

// WithRoomForAMemberUnlessAlreadyIn is the same, for an act that may or may
// not add a seat: changing the role of somebody already in the organisation
// takes no new one, and a limit that refused it would make a full
// organisation one nobody can reorganise.
func (s *Store) WithRoomForAMemberUnlessAlreadyIn(ctx context.Context, organisationID, personID uuid.UUID, add func(*Store) error) error {
	return s.inOneGo(ctx, func(within *Store) error {
		already, err := within.IsInOrganisation(ctx, organisationID, personID)
		if err != nil {
			return err
		}
		if !already {
			if err := within.RoomForAMember(ctx, organisationID); err != nil {
				return err
			}
		}
		return add(within)
	})
}

// IsInOrganisation says whether somebody already holds a seat.
func (s *Store) IsInOrganisation(ctx context.Context, organisationID, personID uuid.UUID) (bool, error) {
	memberships, err := s.db.Queries.ListOrganisationMembershipsForPerson(ctx, pgUUID(personID))
	if err != nil {
		return false, fmt.Errorf("reading the memberships failed: %w", err)
	}
	for _, membership := range memberships {
		if uuid.UUID(membership.OrganisationID.Bytes) == organisationID {
			return true, nil
		}
	}
	return false, nil
}

// WithRoomForAProject is the same for a site.
func (s *Store) WithRoomForAProject(ctx context.Context, organisationID uuid.UUID, add func(*Store) error) error {
	return s.inOneGo(ctx, func(within *Store) error {
		if err := within.RoomForAProject(ctx, organisationID); err != nil {
			return err
		}
		return add(within)
	})
}

// inOneGo gives a function a store that reads and writes inside a single
// transaction, and rolls it back when the function refuses.
func (s *Store) inOneGo(ctx context.Context, do func(*Store) error) error {
	tx, err := s.db.Pool().Begin(ctx)
	if err != nil {
		return fmt.Errorf("starting a transaction failed: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := do(s.within(tx)); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("finishing the transaction failed: %w", err)
	}
	return nil
}

// RoomForAMember refuses when a seat would be one too many. The refusal says
// what it would make and what the limit is, because "that is not allowed"
// leaves whoever reads it with nothing to do next.
func (s *Store) RoomForAMember(ctx context.Context, organisationID uuid.UUID) error {
	usage, err := s.UsageFor(ctx, organisationID)
	if err != nil {
		return err
	}
	if !usage.Members.Full() {
		return nil
	}

	return &httpapi.RefusedError{
		Key: "refused.no-room-for-a-member",
		Reason: fmt.Sprintf(
			"that makes %d members, and the limit is %d; remove somebody, or ask the platform admin to raise the limit",
			usage.Members.Held+1, *usage.Members.Allowed),
		Detail: map[string]any{
			"would":   usage.Members.Held + 1,
			"allowed": *usage.Members.Allowed,
		},
	}
}

// RoomForAProject refuses when a site would be one too many.
func (s *Store) RoomForAProject(ctx context.Context, organisationID uuid.UUID) error {
	usage, err := s.UsageFor(ctx, organisationID)
	if err != nil {
		return err
	}
	if !usage.Projects.Full() {
		return nil
	}

	return &httpapi.RefusedError{
		Key: "refused.no-room-for-a-site",
		Reason: fmt.Sprintf(
			"that makes %d sites, and the limit is %d; archive or remove one, or ask the platform admin to raise the limit",
			usage.Projects.Held+1, *usage.Projects.Allowed),
		Detail: map[string]any{
			"would":   usage.Projects.Held + 1,
			"allowed": *usage.Projects.Allowed,
		},
	}
}

// within is this store with every query running inside the transaction.
func (s *Store) within(tx pgx.Tx) *Store {
	return &Store{db: s.db.Within(tx)}
}

func asInt(value *int32) *int {
	if value == nil {
		return nil
	}
	held := int(*value)
	return &held
}

func asInt32(value *int) *int32 {
	if value == nil {
		return nil
	}
	held := int32(*value)
	return &held
}

// PutSomebodyInWithRoom gives somebody a role in an organisation, refusing
// when that would take a seat there is no room for.
func (s *Store) PutSomebodyInWithRoom(ctx context.Context, organisationID, personID uuid.UUID, role policy.OrganisationRole) error {
	return s.WithRoomForAMemberUnlessAlreadyIn(ctx, organisationID, personID, func(within *Store) error {
		return within.SetOrganisationRole(ctx, organisationID, personID, role)
	})
}
