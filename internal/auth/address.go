package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/mail"
	"github.com/quiqr/chatcms/internal/store"
)

// KindNewAddress is a link that proves a new email address.
const KindNewAddress TokenKind = "new-address"

var (
	// ErrAddressTaken says somebody else already signs in with it.
	ErrAddressTaken = errors.New("that address already belongs to an account")

	// ErrNoAddressWaiting says there is nothing to confirm.
	ErrNoAddressWaiting = errors.New("there is no address waiting to be confirmed")
)

// AskForANewAddress records an address and sends a link to it. Signing in
// keeps using the old one until the link is followed: a typo must not be
// able to lock somebody out of their own account.
func (s *Service) AskForANewAddress(ctx context.Context, who httpapi.Actor, address string) error {
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return httpapi.ErrNotAllowed
	}

	wanted := strings.ToLower(strings.TrimSpace(address))
	if wanted == "" {
		return &httpapi.RefusedError{Reason: "an address cannot be empty"}
	}
	if wanted == strings.ToLower(who.Email) {
		return &httpapi.RefusedError{Reason: "that is already your address"}
	}

	if _, err := s.people.ByEmail(ctx, wanted); err == nil {
		return &httpapi.RefusedError{Reason: ErrAddressTaken.Error()}
	} else if !errors.Is(err, ErrNoSuchPerson) {
		return err
	}

	if err := s.people.db.Queries.SetEmailWaiting(ctx, store.SetEmailWaitingParams{
		ID:           pgUUID(personID),
		EmailWaiting: &wanted,
	}); err != nil {
		return fmt.Errorf("remembering the new address failed: %w", err)
	}

	token, expires, err := s.credentials.Issue(ctx, KindNewAddress, personID)
	if err != nil {
		return err
	}

	// Sent to the new address, because following the link is what proves
	// somebody reads it.
	s.send(ctx, "new address", Person{ID: personID, Email: wanted}, token, expires,
		mail.NewAddress(s.confirmLinkFor(token)))
	return nil
}

// ProveTheNewAddress spends the link and makes the waiting address the one
// this account signs in with.
func (s *Service) ProveTheNewAddress(ctx context.Context, token string) (httpapi.Actor, error) {
	personID, err := s.credentials.SpendWithoutAPassword(ctx, token, KindNewAddress)
	if err != nil {
		return httpapi.Anonymous(), httpapi.ErrLinkUnusable
	}

	row, err := s.people.db.Queries.ConfirmEmailWaiting(ctx, pgUUID(personID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return httpapi.Anonymous(), &httpapi.RefusedError{Reason: ErrNoAddressWaiting.Error()}
		}
		return httpapi.Anonymous(), fmt.Errorf("changing the address failed: %w", err)
	}

	return actorOf(personFrom(row)), nil
}

// WhereSignedIn is one place somebody is signed in, as they see it. It is
// not Session, which is the row: a person reading this wants to know where
// and when, not what the token hash was.
type WhereSignedIn struct {
	ID       uuid.UUID
	LastSeen time.Time
	Started  time.Time
	// ThisOne says it is the session reading the answer.
	ThisOne bool
}

// SessionsFor is everywhere somebody is signed in, newest use first.
func (s *Service) SessionsFor(ctx context.Context, who httpapi.Actor, thisToken string) ([]WhereSignedIn, error) {
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return nil, httpapi.ErrNotAllowed
	}

	rows, err := s.people.db.Queries.SessionsForPerson(ctx, pgUUID(personID))
	if err != nil {
		return nil, fmt.Errorf("reading your sessions failed: %w", err)
	}

	here := hashToken(thisToken)
	out := make([]WhereSignedIn, 0, len(rows))
	for _, row := range rows {
		out = append(out, WhereSignedIn{
			ID:       uuid.UUID(row.ID.Bytes),
			LastSeen: row.LastSeenAt.Time,
			Started:  row.CreatedAt.Time,
			ThisOne:  bytesEqual(row.TokenHash, here),
		})
	}
	return out, nil
}

// EndSession ends one of somebody's own sessions, and nobody else's.
func (s *Service) EndSession(ctx context.Context, who httpapi.Actor, sessionID string) error {
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return httpapi.ErrNotAllowed
	}
	session, err := uuid.Parse(sessionID)
	if err != nil {
		return httpapi.ErrMissing
	}

	// The person is part of the question, so one account cannot end
	// another's by guessing an identifier.
	return s.people.db.Queries.RevokeSessionForPerson(ctx, store.RevokeSessionForPersonParams{
		ID:       pgUUID(session),
		PersonID: pgUUID(personID),
	})
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) || len(a) == 0 {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// SetPicture, RemovePicture and ReadPicture are the service's side of a
// person's picture, so that the HTTP layer talks to one thing.
func (s *Service) SetPicture(ctx context.Context, who httpapi.Actor, content []byte) error {
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return httpapi.ErrNotAllowed
	}

	switch err := s.people.SetPicture(ctx, personID, content); {
	case errors.Is(err, ErrNotAPicture), errors.Is(err, ErrPictureTooLarge):
		return &httpapi.RefusedError{Reason: err.Error()}
	default:
		return err
	}
}

func (s *Service) RemovePicture(ctx context.Context, who httpapi.Actor) error {
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return httpapi.ErrNotAllowed
	}
	return s.people.RemovePicture(ctx, personID)
}

func (s *Service) ReadPicture(ctx context.Context, personID string) ([]byte, string, bool, error) {
	id, err := uuid.Parse(personID)
	if err != nil {
		return nil, "", false, httpapi.ErrMissing
	}

	picture, has, err := s.people.ReadPicture(ctx, id)
	if err != nil {
		return nil, "", false, err
	}
	return picture.Content, picture.Kind, has, nil
}

// SessionViews is where somebody is signed in, in the shape the HTTP layer
// speaks. Sessions() is already the store of them.
func (s *Service) SessionViews(ctx context.Context, who httpapi.Actor, thisToken string) ([]httpapi.SessionView, error) {
	places, err := s.SessionsFor(ctx, who, thisToken)
	if err != nil {
		return nil, err
	}

	out := make([]httpapi.SessionView, 0, len(places))
	for _, place := range places {
		out = append(out, httpapi.SessionView{
			ID:       place.ID.String(),
			LastSeen: place.LastSeen,
			Started:  place.Started,
			ThisOne:  place.ThisOne,
		})
	}
	return out, nil
}
