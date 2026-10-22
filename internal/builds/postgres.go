package builds

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms/internal/store"
)

type Records struct {
	queries *store.Queries
}

func NewRecords(queries *store.Queries) *Records {
	return &Records{queries: queries}
}

var ErrNoSuchBuild = errors.New("builds: no such build")

func (r *Records) Start(ctx context.Context, projectID uuid.UUID, reason string) (Build, error) {
	row, err := r.queries.StartBuild(ctx, store.StartBuildParams{
		ProjectID: pgUUID(projectID),
		Reason:    reason,
	})
	if err != nil {
		return Build{}, fmt.Errorf("builds: recording the start: %w", err)
	}
	return fromRow(row), nil
}

func (r *Records) Finish(ctx context.Context, id uuid.UUID, state State, exitCode *int32, log string) (Build, error) {
	row, err := r.queries.FinishBuild(ctx, store.FinishBuildParams{
		ID:       pgUUID(id),
		State:    string(state),
		ExitCode: exitCode,
		Log:      TrimToLimit(log),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Build{}, ErrNoSuchBuild
	}
	if err != nil {
		return Build{}, fmt.Errorf("builds: recording the finish: %w", err)
	}
	return fromRow(row), nil
}

func (r *Records) Get(ctx context.Context, id uuid.UUID) (Build, error) {
	row, err := r.queries.GetBuild(ctx, pgUUID(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Build{}, ErrNoSuchBuild
	}
	if err != nil {
		return Build{}, fmt.Errorf("builds: reading a build: %w", err)
	}
	return fromRow(row), nil
}

func (r *Records) List(ctx context.Context, projectID uuid.UUID, limit int32) ([]Build, error) {
	rows, err := r.queries.ListBuildsForProject(ctx, store.ListBuildsForProjectParams{
		ProjectID: pgUUID(projectID),
		Limit:     limit,
	})
	if err != nil {
		return nil, fmt.Errorf("builds: listing builds: %w", err)
	}
	out := make([]Build, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromRow(row))
	}
	return out, nil
}

func (r *Records) InterruptLeftovers(ctx context.Context) (int, error) {
	ids, err := r.queries.InterruptRunningBuilds(ctx, interruptedNote)
	if err != nil {
		return 0, err
	}
	return len(ids), nil
}

var _ Recorder = (*Records)(nil)

func fromRow(row store.Build) Build {
	build := Build{
		ID:        uuid.UUID(row.ID.Bytes),
		ProjectID: uuid.UUID(row.ProjectID.Bytes),
		Reason:    row.Reason,
		State:     State(row.State),
		ExitCode:  row.ExitCode,
		Log:       row.Log,
		StartedAt: row.StartedAt.Time,
	}
	if row.FinishedAt.Valid {
		finished := row.FinishedAt.Time
		build.FinishedAt = &finished
	}
	return build
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}
