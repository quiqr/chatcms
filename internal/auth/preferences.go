package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/store"
)

// Appearance is which way somebody likes the interface drawn. Empty is not a
// third appearance: it means they have chosen nothing and the machine
// decides, which is what somebody who never opens the screen gets.
type Appearance string

const (
	FollowTheMachine Appearance = ""
	Light            Appearance = "light"
	Dark             Appearance = "dark"
)

// SetAppearance records what somebody chose, for their own account only.
func (s *Service) SetAppearance(ctx context.Context, who httpapi.Actor, chosen string) error {
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return httpapi.ErrNotAllowed
	}

	switch Appearance(chosen) {
	case FollowTheMachine, Light, Dark:
	default:
		return &httpapi.RefusedError{Reason: "that is not an appearance"}
	}

	if err := s.people.db.Queries.SetAppearance(ctx, store.SetAppearanceParams{
		ID:         pgUUID(personID),
		Appearance: chosen,
	}); err != nil {
		return fmt.Errorf("remembering that failed: %w", err)
	}
	return nil
}

// AppearanceFor is what somebody chose, or nothing when they have not.
func (s *Service) AppearanceFor(ctx context.Context, personID uuid.UUID) (string, error) {
	person, err := s.people.db.Queries.GetPerson(ctx, pgUUID(personID))
	if err != nil {
		return "", fmt.Errorf("reading the preference failed: %w", err)
	}
	return person.Appearance, nil
}
