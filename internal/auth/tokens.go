package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms/internal/store"
)

// TokenKind is what a single-use link is for. An invitation and a reset are
// the same mechanism with different lifetimes and different wording.
type TokenKind string

const (
	KindInvitation    TokenKind = "invitation"
	KindPasswordReset TokenKind = "password-reset"
)

// How long each kind of link stays usable.
const (
	InvitationLifetime    = 7 * 24 * time.Hour
	PasswordResetLifetime = time.Hour
)

// ErrBadToken says a link cannot be used: it was never issued, it has already
// been used, or it has expired. The three are one error so that nothing can be
// learned from which.
var ErrBadToken = errors.New("this link cannot be used any more")

// Credentials issues and spends the links behind invitations and resets.
type Credentials struct {
	db  *store.DB
	now func() time.Time
}

func NewCredentials(db *store.DB) *Credentials {
	return &Credentials{db: db, now: time.Now}
}

// Issue creates a single-use link and returns its value once. The row holds
// only a hash of it.
func (c *Credentials) Issue(ctx context.Context, kind TokenKind, personID uuid.UUID) (string, time.Time, error) {
	token, err := newToken()
	if err != nil {
		return "", time.Time{}, err
	}

	lifetime := InvitationLifetime
	if kind == KindPasswordReset {
		lifetime = PasswordResetLifetime
	}
	expiresAt := c.now().Add(lifetime)

	if _, err := c.db.Queries.CreateCredentialToken(ctx, store.CreateCredentialTokenParams{
		Kind:      string(kind),
		TokenHash: hashToken(token),
		PersonID:  pgUUID(personID),
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	}); err != nil {
		return "", time.Time{}, fmt.Errorf("issuing the link failed: %w", err)
	}

	return token, expiresAt, nil
}

// Spend uses a link to set a password. An invitation sets the first one; a
// reset replaces it and ends every session the person had.
func (c *Credentials) Spend(ctx context.Context, token, password string, people *People, sessions *Sessions) (Person, error) {
	if token == "" {
		return Person{}, ErrBadToken
	}

	row, err := c.db.Queries.GetCredentialTokenByHash(ctx, hashToken(token))
	if errors.Is(err, pgx.ErrNoRows) {
		return Person{}, ErrBadToken
	}
	if err != nil {
		return Person{}, fmt.Errorf("reading the link failed: %w", err)
	}

	if row.SpentAt.Valid || !row.ExpiresAt.Time.After(c.now()) {
		return Person{}, ErrBadToken
	}

	if err := CheckPasswordRules(password); err != nil {
		return Person{}, err
	}

	personID := uuid.UUID(row.PersonID.Bytes)
	if err := people.SetPassword(ctx, personID, password); err != nil {
		return Person{}, err
	}
	if err := c.db.Queries.SpendCredentialToken(ctx, row.ID); err != nil {
		return Person{}, fmt.Errorf("marking the link as used failed: %w", err)
	}

	// Someone resetting a password is often someone who thinks their account
	// is not theirs any more, so everything else they had is ended.
	if TokenKind(row.Kind) == KindPasswordReset {
		if err := sessions.RevokeAll(ctx, personID); err != nil {
			return Person{}, err
		}
	}

	personRow, err := c.db.Queries.GetPerson(ctx, row.PersonID)
	if err != nil {
		return Person{}, fmt.Errorf("reading the person failed: %w", err)
	}
	return personFrom(personRow), nil
}
