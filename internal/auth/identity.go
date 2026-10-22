package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

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
