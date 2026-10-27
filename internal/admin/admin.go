// Package admin is what somebody who administers the installation may do:
// bring organisations into being, say who owns them, and appoint the next
// administrator. It decides nothing itself. Every entry point asks
// policy.Can first, and every act it allows is written down.
package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms/internal/store"
)

var (
	// ErrLastAdministrator says the installation would be left with nobody to
	// administer it, which only SQL on the database could then repair.
	ErrLastAdministrator = errors.New("this is the last administrator, so there would be nobody left to appoint another")

	// ErrNoSuchPerson says nobody holds that email address yet.
	ErrNoSuchPerson = errors.New("nobody with that email address has an account yet")
)

// What names an act worth recording. They are written as the thing that
// happened, because that is how somebody reads them back.
type What string

const (
	MadeAnOrganisation    What = "made an organisation"
	RenamedAnOrganisation What = "renamed an organisation"
	RemovedAnOrganisation What = "removed an organisation"
	PutSomebodyIn         What = "put somebody into an organisation"
	TookSomebodyOut       What = "took somebody out of an organisation"
	AppointedAnAdmin      What = "appointed an administrator"
	StoodDownAnAdmin      What = "removed an administrator"
	TurnedAnAccountOff    What = "turned an account off"
	TurnedAnAccountOn     What = "turned an account on again"
	MadeAPerson           What = "made a person"
)

// Record is one act, with who did it and what it was done to. The names are
// kept as they read at the time, because the record has to outlive what it
// talks about.
type Record struct {
	ID          int64
	PersonID    string
	PersonSaid  string
	Did         What
	Subject     string
	SubjectSaid string
	Detail      map[string]any
	At          time.Time
}

// Store is the database behind administration.
type Store struct {
	db *store.DB
}

func NewStore(db *store.DB) *Store { return &Store{db: db} }

// CountAdministrators is how many people administer the installation.
func (s *Store) CountAdministrators(ctx context.Context) (int64, error) {
	return s.db.Queries.CountAdministrators(ctx)
}

// MakeAdministrator appoints somebody who already has an account.
func (s *Store) MakeAdministrator(ctx context.Context, personID uuid.UUID) error {
	if _, err := s.db.Queries.MakeAdministrator(ctx, pgUUID(personID)); err != nil {
		return fmt.Errorf("appointing an administrator failed: %w", err)
	}
	return nil
}

// MakeAdministratorByEmail appoints the person with that address, and says
// whether it changed anything. It is how the first one is appointed from
// configuration, so it must be safe to run on every start.
func (s *Store) MakeAdministratorByEmail(ctx context.Context, email string) (changed bool, err error) {
	if _, err := s.db.Queries.MakeAdministratorByEmail(ctx, strings.TrimSpace(email)); err == nil {
		return true, nil
	} else if !isNoRows(err) {
		return false, fmt.Errorf("appointing the first administrator failed: %w", err)
	}

	// No row was changed: either there is no such person, or they already
	// administer the installation. The two read differently to whoever is
	// watching the log, so they are told apart here.
	if _, err := s.db.Queries.PersonByEmail(ctx, strings.TrimSpace(email)); err != nil {
		if isNoRows(err) {
			return false, ErrNoSuchPerson
		}
		return false, fmt.Errorf("looking for that person failed: %w", err)
	}
	return false, nil
}

// StopAdministrating takes it away, unless this is the last one.
func (s *Store) StopAdministrating(ctx context.Context, personID uuid.UUID) error {
	if _, err := s.db.Queries.StopAdministrating(ctx, pgUUID(personID)); err != nil {
		if isNoRows(err) {
			return ErrLastAdministrator
		}
		return fmt.Errorf("removing an administrator failed: %w", err)
	}
	return nil
}

// Write records an act.
func (s *Store) Write(ctx context.Context, personID uuid.UUID, personSaid string, did What, subject, subjectSaid string, detail map[string]any) error {
	body := []byte("{}")
	if len(detail) > 0 {
		encoded, err := json.Marshal(detail)
		if err != nil {
			return fmt.Errorf("writing the record failed: %w", err)
		}
		body = encoded
	}

	if _, err := s.db.Queries.RecordAdministration(ctx, store.RecordAdministrationParams{
		PersonID:    pgUUID(personID),
		PersonSaid:  personSaid,
		Did:         string(did),
		Subject:     subject,
		SubjectSaid: subjectSaid,
		Detail:      body,
	}); err != nil {
		return fmt.Errorf("writing the record failed: %w", err)
	}
	return nil
}

// Records reads the newest acts first.
func (s *Store) Records(ctx context.Context, most int32) ([]Record, error) {
	if most <= 0 || most > 500 {
		most = 100
	}
	rows, err := s.db.Queries.ListAdministrationRecords(ctx, most)
	if err != nil {
		return nil, fmt.Errorf("reading the record failed: %w", err)
	}

	out := make([]Record, 0, len(rows))
	for _, row := range rows {
		detail := map[string]any{}
		_ = json.Unmarshal(row.Detail, &detail)
		out = append(out, Record{
			ID:          row.ID,
			PersonID:    uuidText(row.PersonID),
			PersonSaid:  row.PersonSaid,
			Did:         What(row.Did),
			Subject:     row.Subject,
			SubjectSaid: row.SubjectSaid,
			Detail:      detail,
			At:          row.At.Time,
		})
	}
	return out, nil
}

func isNoRows(err error) bool { return errors.Is(err, pgx.ErrNoRows) }

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func uuidText(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	return uuid.UUID(id.Bytes).String()
}

// Belonging is one place somebody can reach, and what they hold there.
type Belonging struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
	// Organisation is the address of the organisation a site is in, and is
	// empty for an organisation itself.
	Organisation string `json:"organisation,omitempty"`
	Role         string `json:"role"`
}

// PersonWithReach is one person and everywhere they can reach.
type PersonWithReach struct {
	ID            string
	Email         string
	DisplayName   string
	Administers   bool
	AccountOn     bool
	StillToAccept bool
	LastSeen      time.Time
	EverSeen      bool
	Organisations []Belonging
	Projects      []Belonging
}

// People is every person on the installation, with where each belongs read
// in the same query. The search is matched against name and address; an
// empty one matches everybody.
func (s *Store) People(ctx context.Context, search string, most int32) ([]PersonWithReach, int64, error) {
	if most <= 0 || most > 500 {
		most = 100
	}

	rows, err := s.db.Queries.ListEveryPerson(ctx, store.ListEveryPersonParams{
		Search: strings.TrimSpace(search),
		Most:   most,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("reading the people failed: %w", err)
	}

	out := make([]PersonWithReach, 0, len(rows))
	var altogether int64
	for _, row := range rows {
		altogether = row.Altogether
		out = append(out, PersonWithReach{
			ID:            uuidText(row.ID),
			Email:         row.Email,
			DisplayName:   row.DisplayName,
			Administers:   row.Administers,
			AccountOn:     !row.DisabledAt.Valid,
			StillToAccept: !row.HasPassword,
			LastSeen:      row.LastSeen.Time,
			EverSeen:      row.LastSeen.Valid,
			Organisations: belongings(row.Organisations),
			Projects:      belongings(row.Projects),
		})
	}
	return out, altogether, nil
}

// Person is one of them, by the address that identifies them.
func (s *Store) Person(ctx context.Context, personID uuid.UUID) (PersonWithReach, error) {
	row, err := s.db.Queries.PersonWithReach(ctx, pgUUID(personID))
	if err != nil {
		if isNoRows(err) {
			return PersonWithReach{}, ErrNoSuchPerson
		}
		return PersonWithReach{}, fmt.Errorf("reading that person failed: %w", err)
	}

	return PersonWithReach{
		ID:            uuidText(row.ID),
		Email:         row.Email,
		DisplayName:   row.DisplayName,
		Administers:   row.Administers,
		AccountOn:     !row.DisabledAt.Valid,
		StillToAccept: !row.HasPassword,
		LastSeen:      row.LastSeen.Time,
		EverSeen:      row.LastSeen.Valid,
		Organisations: belongings(row.Organisations),
		Projects:      belongings(row.Projects),
	}, nil
}

func belongings(encoded []byte) []Belonging {
	out := []Belonging{}
	if len(encoded) == 0 {
		return out
	}
	_ = json.Unmarshal(encoded, &out)
	return out
}
