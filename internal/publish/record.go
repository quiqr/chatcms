package publish

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms/internal/store"
	"github.com/quiqr/chatcms/internal/workflows"
)

type State string

const (
	Running     State = "running"
	Succeeded   State = "succeeded"
	Failed      State = "failed"
	Interrupted State = "interrupted"
)

const interruptedNote = "This publish was still going when the server stopped, so we do not know whether it finished."

// Record is one publish: who asked for it, what it carried, how it went.
type Record struct {
	ID          uuid.UUID               `json:"id"`
	ProjectID   uuid.UUID               `json:"-"`
	PersonID    *uuid.UUID              `json:"-"`
	Note        string                  `json:"note,omitempty"`
	State       State                   `json:"state"`
	ChangeCount int                     `json:"changeCount"`
	Summary     []Change                `json:"summary,omitempty"`
	Steps       []workflows.StepOutcome `json:"steps,omitempty"`
	Trouble     string                  `json:"trouble,omitempty"`
	Detail      string                  `json:"-"`
	StartedAt   time.Time               `json:"startedAt"`
	FinishedAt  *time.Time              `json:"finishedAt,omitempty"`
}

var (
	ErrNoSuchPublish     = errors.New("publish: no such publish")
	ErrAlreadyPublishing = errors.New("publish: this site is already being published")
)

type Records struct {
	queries *store.Queries
}

func NewRecords(queries *store.Queries) *Records {
	return &Records{queries: queries}
}

func (r *Records) Start(ctx context.Context, projectID uuid.UUID, personID *uuid.UUID, note string, summary Summary) (Record, error) {
	encoded, err := json.Marshal(summary.Changes)
	if err != nil {
		return Record{}, err
	}

	row, err := r.queries.StartPublish(ctx, store.StartPublishParams{
		ProjectID:   pgUUID(projectID),
		PersonID:    optionalUUID(personID),
		Note:        note,
		ChangeCount: int32(summary.Count()),
		Summary:     encoded,
	})
	if isUniqueViolation(err) {
		return Record{}, ErrAlreadyPublishing
	}
	if err != nil {
		return Record{}, fmt.Errorf("publish: recording the start: %w", err)
	}
	return fromRow(row), nil
}

func (r *Records) Finish(ctx context.Context, id uuid.UUID, state State, steps []workflows.StepOutcome, trouble, detail string) (Record, error) {
	encoded, err := json.Marshal(steps)
	if err != nil {
		return Record{}, err
	}

	row, err := r.queries.FinishPublish(ctx, store.FinishPublishParams{
		ID:      pgUUID(id),
		State:   string(state),
		Steps:   encoded,
		Trouble: trouble,
		Detail:  detail,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, ErrNoSuchPublish
	}
	if err != nil {
		return Record{}, fmt.Errorf("publish: recording the finish: %w", err)
	}
	return fromRow(row), nil
}

func (r *Records) Get(ctx context.Context, id uuid.UUID) (Record, error) {
	row, err := r.queries.GetPublish(ctx, pgUUID(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, ErrNoSuchPublish
	}
	if err != nil {
		return Record{}, fmt.Errorf("publish: reading a publish: %w", err)
	}
	return fromRow(row), nil
}

func (r *Records) List(ctx context.Context, projectID uuid.UUID, limit int32) ([]Record, error) {
	rows, err := r.queries.ListPublishesForProject(ctx, store.ListPublishesForProjectParams{
		ProjectID: pgUUID(projectID),
		Limit:     limit,
	})
	if err != nil {
		return nil, fmt.Errorf("publish: listing publishes: %w", err)
	}
	out := make([]Record, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromRow(row))
	}
	return out, nil
}

func (r *Records) LastSucceeded(ctx context.Context, projectID uuid.UUID) (Record, error) {
	row, err := r.queries.LastSucceededPublish(ctx, pgUUID(projectID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, ErrNoSuchPublish
	}
	if err != nil {
		return Record{}, fmt.Errorf("publish: reading the last publish: %w", err)
	}
	return fromRow(row), nil
}

func (r *Records) InterruptLeftovers(ctx context.Context) (int, error) {
	ids, err := r.queries.InterruptRunningPublishes(ctx, interruptedNote)
	if err != nil {
		return 0, err
	}
	return len(ids), nil
}

func fromRow(row store.Publish) Record {
	record := Record{
		ID:          uuid.UUID(row.ID.Bytes),
		ProjectID:   uuid.UUID(row.ProjectID.Bytes),
		Note:        row.Note,
		State:       State(row.State),
		ChangeCount: int(row.ChangeCount),
		Trouble:     row.Trouble,
		Detail:      row.Detail,
		StartedAt:   row.StartedAt.Time,
	}
	if row.PersonID.Valid {
		person := uuid.UUID(row.PersonID.Bytes)
		record.PersonID = &person
	}
	if row.FinishedAt.Valid {
		finished := row.FinishedAt.Time
		record.FinishedAt = &finished
	}
	json.Unmarshal(row.Summary, &record.Summary)
	json.Unmarshal(row.Steps, &record.Steps)
	return record
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func optionalUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgUUID(*id)
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pgErr interface{ SQLState() string }
	return errors.As(err, &pgErr) && pgErr.SQLState() == "23505"
}
