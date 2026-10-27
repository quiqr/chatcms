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
		return &httpapi.RefusedError{Key: "refused.not-appearance", Reason: "that is not an appearance"}
	}

	if err := s.people.db.Queries.SetAppearance(ctx, store.SetAppearanceParams{
		ID:         pgUUID(personID),
		Appearance: chosen,
	}); err != nil {
		return fmt.Errorf("remembering that failed: %w", err)
	}
	return nil
}

// Language is which language somebody reads the interface in. Empty is not
// a third language: it means they have chosen none and the browser decides,
// which is what a Dutch customer gets before they have found this screen.
type Language string

const (
	FollowTheBrowser Language = ""
	English          Language = "en"
	Dutch            Language = "nl"
)

// SetLanguage records what somebody chose, for their own account only.
func (s *Service) SetLanguage(ctx context.Context, who httpapi.Actor, chosen string) error {
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return httpapi.ErrNotAllowed
	}

	switch Language(chosen) {
	case FollowTheBrowser, English, Dutch:
	default:
		return &httpapi.RefusedError{
			Key:    "refused.not-language-interface-has",
			Reason: "that is not a language this interface has",
		}
	}

	if err := s.people.db.Queries.SetLanguage(ctx, store.SetLanguageParams{
		ID:       pgUUID(personID),
		Language: chosen,
	}); err != nil {
		return fmt.Errorf("remembering that failed: %w", err)
	}
	return nil
}

// LanguageFor is what somebody chose, or nothing when they have not.
func (s *Service) LanguageFor(ctx context.Context, personID uuid.UUID) (string, error) {
	person, err := s.people.db.Queries.GetPerson(ctx, pgUUID(personID))
	if err != nil {
		return "", fmt.Errorf("reading the preference failed: %w", err)
	}
	return person.Language, nil
}

// AppearanceFor is what somebody chose, or nothing when they have not.
func (s *Service) AppearanceFor(ctx context.Context, personID uuid.UUID) (string, error) {
	person, err := s.people.db.Queries.GetPerson(ctx, pgUUID(personID))
	if err != nil {
		return "", fmt.Errorf("reading the preference failed: %w", err)
	}
	return person.Appearance, nil
}
