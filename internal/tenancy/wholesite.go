package tenancy

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/policy"
	"github.com/quiqr/chatcms/internal/store"
)

// MoveProject puts a site in another organisation. Nothing on disk moves: a
// working copy and a remote are named by the project, not by the
// organisation, so git is told nothing at all.
//
// The people who hold a role on the site become members of the destination,
// because somebody who edits a site and belongs to no organisation that
// holds it would lose it the moment the move finished.
func (s *Store) MoveProject(ctx context.Context, projectID, to uuid.UUID) error {
	theirs, err := s.db.Queries.ListProjectMemberIDs(ctx, pgUUID(projectID))
	if err != nil {
		return fmt.Errorf("reading who is on the site failed: %w", err)
	}

	return s.inOneGo(ctx, func(within *Store) error {
		// Counted before anything moves: the people who would become
		// members are what the destination's limit is asked about.
		coming := 0
		for _, person := range theirs {
			already, err := within.IsInOrganisation(ctx, to, uuid.UUID(person.Bytes))
			if err != nil {
				return err
			}
			if !already {
				coming++
			}
		}
		if err := within.roomForThisMany(ctx, to, coming); err != nil {
			return err
		}

		if err := within.db.Queries.MoveProject(ctx, store.MoveProjectParams{
			ID:             pgUUID(projectID),
			OrganisationID: pgUUID(to),
		}); err != nil {
			return fmt.Errorf("moving the site failed: %w", err)
		}

		for _, person := range theirs {
			already, err := within.IsInOrganisation(ctx, to, uuid.UUID(person.Bytes))
			if err != nil {
				return err
			}
			if already {
				continue
			}
			if err := within.SetOrganisationRole(ctx, to, uuid.UUID(person.Bytes),
				policy.OrganisationMember); err != nil {
				return err
			}
		}
		return nil
	})
}

// roomForThisMany refuses when adding this many would take an organisation
// over its limit of members. Asking about several at once is what moving a
// site needs: the people come as one act or not at all.
func (s *Store) roomForThisMany(ctx context.Context, organisationID uuid.UUID, many int) error {
	if many == 0 {
		return nil
	}

	usage, err := s.UsageFor(ctx, organisationID)
	if err != nil {
		return err
	}
	if !usage.Members.Limited() || usage.Members.Held+many <= *usage.Members.Allowed {
		return nil
	}

	return &httpapi.RefusedError{
		Key: "refused.no-room-for-the-people",
		Reason: fmt.Sprintf(
			"that makes %d members, and the limit is %d; the people on the site come with it",
			usage.Members.Held+many, *usage.Members.Allowed),
		Detail: map[string]any{
			"would":   usage.Members.Held + many,
			"allowed": *usage.Members.Allowed,
		},
	}
}

// Archive stops a site counting against its organisation's limit, and stops
// it being in the way. Nothing is removed, and what is published stays
// published.
func (s *Store) Archive(ctx context.Context, projectID uuid.UUID) error {
	if err := s.db.Queries.ArchiveProject(ctx, pgUUID(projectID)); err != nil {
		return fmt.Errorf("archiving the site failed: %w", err)
	}
	return nil
}

// BringBack is the other way.
func (s *Store) BringBack(ctx context.Context, projectID uuid.UUID) error {
	return s.inOneGo(ctx, func(within *Store) error {
		row, err := within.db.Queries.GetProject(ctx, pgUUID(projectID))
		if err != nil {
			return fmt.Errorf("reading the site failed: %w", err)
		}
		project := projectFrom(row)
		// It counts again the moment it is back, so the limit is asked
		// first, exactly as it would be for a new one.
		if err := within.RoomForAProject(ctx, project.OrganisationID); err != nil {
			return err
		}
		if err := within.db.Queries.BringProjectBack(ctx, pgUUID(projectID)); err != nil {
			return fmt.Errorf("bringing the site back failed: %w", err)
		}
		return nil
	})
}

// CopyThePeople gives the copy the roles the original had. It is only asked
// for when somebody ticked the box: a copy is usually made to try something,
// and handing everybody a role on it is not what trying something means.
func (s *Store) CopyThePeople(ctx context.Context, from, to uuid.UUID) error {
	members, err := s.Members(ctx, from)
	if err != nil {
		return err
	}
	for _, member := range members {
		if err := s.SetProjectRole(ctx, to, member.PersonID, member.Role); err != nil {
			return err
		}
	}
	return nil
}
