package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms/internal/redact"
	"github.com/quiqr/chatcms/internal/store"
)

// Job is one piece of work waiting to be done.
type Job struct {
	ID       int64
	Kind     string
	Subject  string
	Payload  []byte
	Attempts int
	RunAt    time.Time
}

// Handler does the work of one job kind.
type Handler func(ctx context.Context, job Job) error

// Queue holds work that must not block a request.
type Queue struct {
	db    *store.DB
	limit int
}

// Policy decides how often a failed job is tried again and how many times.
type Policy struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

// DefaultPolicy retries often enough to ride out a remote that is briefly
// unavailable, and gives up before a permanently broken job is retried
// forever.
func DefaultPolicy() Policy {
	return Policy{MaxAttempts: 6, BaseDelay: 5 * time.Second, MaxDelay: 10 * time.Minute}
}

// Delay is how long to wait before attempt number n, counting from one.
func (p Policy) Delay(attempt int) time.Duration {
	delay := p.BaseDelay
	for range attempt - 1 {
		delay *= 2
		if delay >= p.MaxDelay {
			return p.MaxDelay
		}
	}
	if delay > p.MaxDelay {
		return p.MaxDelay
	}
	return delay
}

// ErrNothingToDo says the queue held no job that is ready to run.
var ErrNothingToDo = errors.New("the queue has nothing ready to run")

func NewQueue(db *store.DB) *Queue {
	return &Queue{db: db}
}

// Enqueue adds work and returns at once. Work of the same kind for the same
// subject that is already waiting is not queued again; work whose run has
// already started does not stop a new one being queued, so a request made
// during a run is not lost.
func (q *Queue) Enqueue(ctx context.Context, kind, subject string, payload any) (int64, error) {
	encoded := []byte("{}")
	if payload != nil {
		body, err := json.Marshal(payload)
		if err != nil {
			return 0, fmt.Errorf("the job payload could not be written: %w", err)
		}
		encoded = body
	}

	job, err := q.db.Queries.EnqueueJob(ctx, store.EnqueueJobParams{
		Kind:    kind,
		Subject: subject,
		Payload: encoded,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		// The same work is already waiting, which is the answer the caller
		// wanted: it will happen.
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("queueing %s failed: %w", kind, err)
	}
	return job.ID, nil
}

// EnqueueAt adds work to run no earlier than a given time.
func (q *Queue) EnqueueAt(ctx context.Context, kind, subject string, payload any, when time.Time) (int64, error) {
	encoded := []byte("{}")
	if payload != nil {
		body, err := json.Marshal(payload)
		if err != nil {
			return 0, fmt.Errorf("the job payload could not be written: %w", err)
		}
		encoded = body
	}

	job, err := q.db.Queries.EnqueueJob(ctx, store.EnqueueJobParams{
		Kind:    kind,
		Subject: subject,
		Payload: encoded,
		RunAt:   pgtype.Timestamptz{Time: when, Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("queueing %s failed: %w", kind, err)
	}
	return job.ID, nil
}

// Claim takes the next job that is ready, exclusively. Two workers claiming at
// the same moment never get the same job.
func (q *Queue) Claim(ctx context.Context) (Job, error) {
	row, err := q.db.Queries.ClaimNextJob(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return Job{}, ErrNothingToDo
	}
	if err != nil {
		return Job{}, fmt.Errorf("claiming a job failed: %w", err)
	}

	return Job{
		ID:       row.ID,
		Kind:     row.Kind,
		Subject:  row.Subject,
		Payload:  row.Payload,
		Attempts: int(row.Attempts),
		RunAt:    row.RunAt.Time,
	}, nil
}

// Complete records a job as done.
func (q *Queue) Complete(ctx context.Context, id int64) error {
	return q.db.Queries.CompleteJob(ctx, id)
}

// Retry puts a failed job back with a growing delay, or records it as failed
// once it has used up its attempts.
func (q *Queue) Retry(ctx context.Context, job Job, policy Policy, reason error) error {
	message := redact.Secrets(reason.Error())

	if job.Attempts >= policy.MaxAttempts {
		return q.db.Queries.FailJob(ctx, store.FailJobParams{ID: job.ID, LastError: &message})
	}

	when := time.Now().Add(policy.Delay(job.Attempts))
	return q.db.Queries.RetryJob(ctx, store.RetryJobParams{
		ID:        job.ID,
		RunAt:     pgtype.Timestamptz{Time: when, Valid: true},
		LastError: &message,
	})
}

// Release returns a job nobody finished to the queue, so it is picked up
// again rather than counted as done.
func (q *Queue) Release(ctx context.Context, id int64) error {
	return q.db.Queries.ReleaseJob(ctx, id)
}

// CountByState is how many jobs are in one state, which is what the interface
// and the tests ask.
func (q *Queue) CountByState(ctx context.Context, state string) (int, error) {
	count, err := q.db.Queries.CountJobsByState(ctx, state)
	return int(count), err
}

// CountForSubject is how many jobs of one kind and subject are in a state.
func (q *Queue) CountForSubject(ctx context.Context, kind, subject, state string) (int, error) {
	count, err := q.db.Queries.CountJobsForSubject(ctx, store.CountJobsForSubjectParams{
		Kind: kind, Subject: subject, State: state,
	})
	return int(count), err
}
