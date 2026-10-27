package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms/internal/store"
)

// How long a session lasts. A working day when it was not asked to be
// remembered, and long enough not to ask again in ordinary use when it was.
const (
	SessionLifetime           = 12 * time.Hour
	RememberedSessionLifetime = 60 * 24 * time.Hour
)

var (
	// ErrNoSession says the presented value identifies nothing usable: it was
	// never issued, it has expired, or it was revoked. The three are one
	// error on purpose.
	ErrNoSession = errors.New("this session is not valid")

	// ErrNoSuchPerson says there is nobody with that address.
	ErrNoSuchPerson = errors.New("there is nobody with that address")

	// ErrNoPassword says this person has no password, so a password cannot
	// sign them in.
	ErrNoPassword = errors.New("this person has no password set")

	// ErrDisabled says the person may not sign in.
	ErrDisabled = errors.New("this person may not sign in")
)

// Person is who a request is from.
type Person struct {
	ID          uuid.UUID
	Email       string
	DisplayName string
	HasPassword bool
	// Off says the account has been turned off, so it may not be signed in
	// to and may not be acted as.
	Off bool
}

// Session is a remembered sign-in.
type Session struct {
	ID         uuid.UUID
	PersonID   uuid.UUID
	Remembered bool
	ExpiresAt  time.Time
	// ActingAs is who the person signed in is acting as, and is nil the rest
	// of the time. The session stays theirs while it is set: issuing one as
	// the other person would lose who is really there.
	ActingAs *uuid.UUID
}

// Sessions issues, finds and revokes sessions.
type Sessions struct {
	db  *store.DB
	now func() time.Time
}

func NewSessions(db *store.DB) *Sessions {
	return &Sessions{db: db, now: time.Now}
}

// Issue creates a session and returns the value the cookie carries. The value
// is returned once and never stored; the row holds only its hash.
func (s *Sessions) Issue(ctx context.Context, personID uuid.UUID, remembered bool) (string, Session, error) {
	token, err := newToken()
	if err != nil {
		return "", Session{}, err
	}

	lifetime := SessionLifetime
	if remembered {
		lifetime = RememberedSessionLifetime
	}

	row, err := s.db.Queries.CreateSession(ctx, store.CreateSessionParams{
		TokenHash:  hashToken(token),
		PersonID:   pgUUID(personID),
		Remembered: remembered,
		ExpiresAt:  pgtype.Timestamptz{Time: s.now().Add(lifetime), Valid: true},
	})
	if err != nil {
		return "", Session{}, fmt.Errorf("creating the session failed: %w", err)
	}

	return token, sessionFrom(row), nil
}

// Lookup finds the person a presented session value identifies, and records
// that the session was used. Using a session does not extend it.
func (s *Sessions) Lookup(ctx context.Context, token string) (Person, Session, error) {
	if token == "" {
		return Person{}, Session{}, ErrNoSession
	}

	row, err := s.db.Queries.GetSessionByTokenHash(ctx, hashToken(token))
	if errors.Is(err, pgx.ErrNoRows) {
		return Person{}, Session{}, ErrNoSession
	}
	if err != nil {
		return Person{}, Session{}, fmt.Errorf("reading the session failed: %w", err)
	}

	if row.Session.RevokedAt.Valid || !row.Session.ExpiresAt.Time.After(s.now()) {
		return Person{}, Session{}, ErrNoSession
	}
	if row.Person.DisabledAt.Valid {
		return Person{}, Session{}, ErrNoSession
	}

	if err := s.db.Queries.TouchSession(ctx, row.Session.ID); err != nil {
		return Person{}, Session{}, fmt.Errorf("recording the session's use failed: %w", err)
	}

	return personFrom(row.Person), sessionFrom(row.Session), nil
}

// Revoke ends one session, named by the value its cookie carries.
func (s *Sessions) Revoke(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.db.Queries.RevokeSessionByTokenHash(ctx, hashToken(token))
}

// RevokeAll ends every session a person has, which is what a password reset
// does.
func (s *Sessions) RevokeAll(ctx context.Context, personID uuid.UUID) error {
	return s.db.Queries.RevokeSessionsForPerson(ctx, pgUUID(personID))
}

// LiveCount is how many usable sessions a person has, which the tests read.
func (s *Sessions) LiveCount(ctx context.Context, personID uuid.UUID) (int, error) {
	count, err := s.db.Queries.CountLiveSessionsForPerson(ctx, pgUUID(personID))
	return int(count), err
}

func newToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("no randomness available: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// hashToken is a plain hash, which is right here and wrong for passwords: the
// value has full entropy, so there is nothing to guess, and lookups have to
// be fast.
func hashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

func sessionFrom(row store.Session) Session {
	session := Session{
		ID:         uuid.UUID(row.ID.Bytes),
		PersonID:   uuid.UUID(row.PersonID.Bytes),
		Remembered: row.Remembered,
		ExpiresAt:  row.ExpiresAt.Time,
	}
	if row.ActingAs.Valid {
		acting := uuid.UUID(row.ActingAs.Bytes)
		session.ActingAs = &acting
	}
	return session
}

func personFrom(row store.Person) Person {
	return Person{
		ID:          uuid.UUID(row.ID.Bytes),
		Email:       row.Email,
		DisplayName: row.DisplayName,
		HasPassword: row.PasswordHash != nil && *row.PasswordHash != "",
		Off:         row.DisabledAt.Valid,
	}
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

// ActAs says that this session is being used to act as somebody else. The
// session is not reissued: it stays the session of whoever signed in.
func (s *Sessions) ActAs(ctx context.Context, token string, personID uuid.UUID) error {
	if token == "" {
		return ErrNoSession
	}
	return s.db.Queries.ActAs(ctx, store.ActAsParams{
		TokenHash: hashToken(token),
		ActingAs:  pgUUID(personID),
	})
}

// StopActingAs ends it, leaving the session as it was.
func (s *Sessions) StopActingAs(ctx context.Context, token string) error {
	if token == "" {
		return ErrNoSession
	}
	return s.db.Queries.StopActingAs(ctx, hashToken(token))
}
