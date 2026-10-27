package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms/internal/store"
)

// Identity answers whether someone is who they claim to be. It is an
// interface with one implementation so that single sign-on can be added
// without anything that reads a session changing. See
// docs/ARCHITECTURE.md.
type Identity interface {
	// Authenticate returns the person the claim identifies, or an error.
	Authenticate(ctx context.Context, claim Claim) (Person, error)
}

// Claim is what someone presents to prove who they are. A password is one
// kind of claim; an assertion from an identity provider would be another.
type Claim struct {
	Email    string
	Password string
}

// PasswordIdentity checks a claim against a stored password.
type PasswordIdentity struct {
	db *store.DB
}

func NewPasswordIdentity(db *store.DB) *PasswordIdentity {
	return &PasswordIdentity{db: db}
}

func (p *PasswordIdentity) Authenticate(ctx context.Context, claim Claim) (Person, error) {
	row, err := p.db.Queries.GetPersonByEmail(ctx, strings.TrimSpace(claim.Email))
	if errors.Is(err, pgx.ErrNoRows) {
		// The same work is done for an address nobody has, so that a wrong
		// address does not come back faster than a wrong password.
		SpendTimeLikeAFailedCheck(claim.Password)
		return Person{}, ErrNoSuchPerson
	}
	if err != nil {
		return Person{}, fmt.Errorf("reading the person failed: %w", err)
	}

	if row.DisabledAt.Valid {
		SpendTimeLikeAFailedCheck(claim.Password)
		return Person{}, ErrDisabled
	}

	if row.PasswordHash == nil || *row.PasswordHash == "" {
		SpendTimeLikeAFailedCheck(claim.Password)
		return Person{}, ErrNoPassword
	}

	if err := VerifyPassword(*row.PasswordHash, claim.Password); err != nil {
		return Person{}, ErrWrongPassword
	}

	return personFrom(row), nil
}

// People creates and reads the people the system knows.
type People struct {
	db *store.DB
}

func NewPeople(db *store.DB) *People {
	return &People{db: db}
}

// Create adds a person, optionally with a password. A person with no password
// cannot sign in with one, which is what an invited person looks like until
// they accept.
func (p *People) Create(ctx context.Context, email, displayName, password string) (Person, error) {
	var hash *string
	if password != "" {
		encoded, err := HashPassword(password)
		if err != nil {
			return Person{}, err
		}
		hash = &encoded
	}

	row, err := p.db.Queries.CreatePerson(ctx, store.CreatePersonParams{
		Email:        strings.TrimSpace(email),
		DisplayName:  displayName,
		PasswordHash: hash,
	})
	if err != nil {
		return Person{}, fmt.Errorf("creating the person failed: %w", err)
	}
	return personFrom(row), nil
}

// ByEmail finds a person by address, without regard to case.
func (p *People) ByEmail(ctx context.Context, email string) (Person, error) {
	row, err := p.db.Queries.GetPersonByEmail(ctx, strings.TrimSpace(email))
	if errors.Is(err, pgx.ErrNoRows) {
		return Person{}, ErrNoSuchPerson
	}
	if err != nil {
		return Person{}, fmt.Errorf("reading the person failed: %w", err)
	}
	return personFrom(row), nil
}

// ByID finds a person by their identifier.
func (p *People) ByID(ctx context.Context, personID uuid.UUID) (Person, error) {
	row, err := p.db.Queries.GetPerson(ctx, pgtype.UUID{Bytes: personID, Valid: true})
	if errors.Is(err, pgx.ErrNoRows) {
		return Person{}, ErrNoSuchPerson
	}
	if err != nil {
		return Person{}, fmt.Errorf("reading the person failed: %w", err)
	}
	return personFrom(row), nil
}

// SetPassword replaces a person's password.
func (p *People) SetPassword(ctx context.Context, personID uuid.UUID, password string) error {
	encoded, err := HashPassword(password)
	if err != nil {
		return err
	}
	return p.db.Queries.SetPersonPassword(ctx, store.SetPersonPasswordParams{
		ID:           pgUUID(personID),
		PasswordHash: &encoded,
	})
}

// TurnOff stops an account being used. What the person wrote stays theirs:
// history that renames its author is history nobody can trust.
func (p *People) TurnOff(ctx context.Context, personID uuid.UUID) error {
	if _, err := p.db.Queries.TurnAccountOff(ctx, pgUUID(personID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Already off, which is what was asked for.
			return nil
		}
		return fmt.Errorf("turning the account off failed: %w", err)
	}
	return nil
}

// TurnOn lets an account be used again, with the password it had.
func (p *People) TurnOn(ctx context.Context, personID uuid.UUID) error {
	if _, err := p.db.Queries.TurnAccountOn(ctx, pgUUID(personID)); err != nil {
		return fmt.Errorf("turning the account on failed: %w", err)
	}
	return nil
}

// Rename changes what somebody is called.
func (p *People) Rename(ctx context.Context, personID uuid.UUID, displayName string) (Person, error) {
	row, err := p.db.Queries.SetDisplayName(ctx, store.SetDisplayNameParams{
		ID:          pgUUID(personID),
		DisplayName: strings.TrimSpace(displayName),
	})
	if err != nil {
		return Person{}, fmt.Errorf("changing the name failed: %w", err)
	}
	return personFrom(row), nil
}

// LastSeen is when somebody was last here, and whether they ever were.
func (p *People) LastSeen(ctx context.Context, personID uuid.UUID) (time.Time, bool, error) {
	seen, err := p.db.Queries.LastSeenFor(ctx, pgUUID(personID))
	if err != nil {
		return time.Time{}, false, fmt.Errorf("reading when they were last here failed: %w", err)
	}
	if !seen.Valid {
		return time.Time{}, false, nil
	}
	return seen.Time, true, nil
}
