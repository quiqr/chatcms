//go:build integration

package jobs_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/quiqr/chatcms/internal/jobs"
	"github.com/quiqr/chatcms/internal/store"
	"github.com/quiqr/chatcms/internal/testsupport"
)

const kind = "sync"

func newQueue(t *testing.T) (*jobs.Queue, *store.DB) {
	t.Helper()

	db, err := store.Open(t.Context(), testsupport.Postgres(t))
	if err != nil {
		t.Fatalf("store.Open() error = %v, want nil", err)
	}
	t.Cleanup(db.Close)

	if err := db.Migrate(t.Context()); err != nil {
		t.Fatalf("Migrate() error = %v, want nil", err)
	}
	return jobs.NewQueue(db), db
}

func quiet() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func TestQueueReturnsImmediately(t *testing.T) {
	queue, _ := newQueue(t)

	started := time.Now()
	id, err := queue.Enqueue(t.Context(), kind, "korenaar", map[string]string{"site": "korenaar"})
	if err != nil {
		t.Fatalf("Enqueue() error = %v, want nil", err)
	}
	if id == 0 {
		t.Fatal("Enqueue() returned no identifier")
	}
	if time.Since(started) > time.Second {
		t.Errorf("Enqueue took %v, want it to return at once", time.Since(started))
	}

	queued, _ := queue.CountByState(t.Context(), "queued")
	if queued != 1 {
		t.Errorf("queued = %d, want 1", queued)
	}
}

func TestQueuedJobSurvivesARestart(t *testing.T) {
	url := testsupport.Postgres(t)

	first, err := store.Open(t.Context(), url)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.Migrate(t.Context()); err != nil {
		t.Fatal(err)
	}
	if _, err := jobs.NewQueue(first).Enqueue(t.Context(), kind, "korenaar", nil); err != nil {
		t.Fatal(err)
	}
	first.Close()

	second, err := store.Open(t.Context(), url)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()

	job, err := jobs.NewQueue(second).Claim(t.Context())
	if err != nil {
		t.Fatalf("Claim() after a restart error = %v, want the job", err)
	}
	if job.Subject != "korenaar" {
		t.Errorf("subject = %q, want korenaar", job.Subject)
	}
}

func TestTwoWorkersOneJob(t *testing.T) {
	queue, _ := newQueue(t)
	if _, err := queue.Enqueue(t.Context(), kind, "korenaar", nil); err != nil {
		t.Fatal(err)
	}

	var wait sync.WaitGroup
	results := make([]error, 2)
	claimed := make([]jobs.Job, 2)

	for i := range results {
		wait.Add(1)
		go func() {
			defer wait.Done()
			claimed[i], results[i] = queue.Claim(context.Background())
		}()
	}
	wait.Wait()

	got, empty := 0, 0
	for i, err := range results {
		switch {
		case err == nil:
			got++
			if claimed[i].ID == 0 {
				t.Error("a claimed job has no identifier")
			}
		case errors.Is(err, jobs.ErrNothingToDo):
			empty++
		default:
			t.Fatalf("Claim() error = %v", err)
		}
	}
	if got != 1 || empty != 1 {
		t.Errorf("%d workers got the job and %d found nothing, want one each", got, empty)
	}
}

func TestClaimOnAnEmptyQueue(t *testing.T) {
	queue, _ := newQueue(t)

	_, err := queue.Claim(t.Context())
	if !errors.Is(err, jobs.ErrNothingToDo) {
		t.Errorf("Claim() error = %v, want ErrNothingToDo", err)
	}
}

func TestQueueingTwiceQueuesOnce(t *testing.T) {
	queue, _ := newQueue(t)

	for range 3 {
		if _, err := queue.Enqueue(t.Context(), kind, "korenaar", nil); err != nil {
			t.Fatalf("Enqueue() error = %v, want nil", err)
		}
	}

	queued, _ := queue.CountForSubject(t.Context(), kind, "korenaar", "queued")
	if queued != 1 {
		t.Errorf("queued = %d, want 1", queued)
	}
}

func TestQueueingForAnotherSubjectIsItsOwnJob(t *testing.T) {
	queue, _ := newQueue(t)

	if _, err := queue.Enqueue(t.Context(), kind, "korenaar", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := queue.Enqueue(t.Context(), kind, "studio-noord", nil); err != nil {
		t.Fatal(err)
	}

	queued, _ := queue.CountByState(t.Context(), "queued")
	if queued != 2 {
		t.Errorf("queued = %d, want one per site", queued)
	}
}

func TestQueueingDuringARunQueuesOneMore(t *testing.T) {
	queue, _ := newQueue(t)

	if _, err := queue.Enqueue(t.Context(), kind, "korenaar", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := queue.Claim(t.Context()); err != nil {
		t.Fatal(err)
	}

	if _, err := queue.Enqueue(t.Context(), kind, "korenaar", nil); err != nil {
		t.Fatalf("Enqueue() during a run error = %v, want the request to be kept", err)
	}

	queued, _ := queue.CountForSubject(t.Context(), kind, "korenaar", "queued")
	if queued != 1 {
		t.Errorf("queued = %d, want the request made during the run to be waiting", queued)
	}
}

func TestFailedJobIsRetriedLater(t *testing.T) {
	queue, _ := newQueue(t)
	policy := jobs.Policy{MaxAttempts: 3, BaseDelay: time.Hour, MaxDelay: time.Hour}

	if _, err := queue.Enqueue(t.Context(), kind, "korenaar", nil); err != nil {
		t.Fatal(err)
	}
	job, err := queue.Claim(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	if err := queue.Retry(t.Context(), job, policy, errors.New("the remote did not answer")); err != nil {
		t.Fatalf("Retry() error = %v, want nil", err)
	}

	if _, err := queue.Claim(t.Context()); !errors.Is(err, jobs.ErrNothingToDo) {
		t.Error("the job ran again immediately rather than after its delay")
	}

	queued, _ := queue.CountByState(t.Context(), "queued")
	if queued != 1 {
		t.Errorf("queued = %d, want the job waiting for its delay", queued)
	}
}

func TestRetriedJobRunsAgainOnceItsDelayHasPassed(t *testing.T) {
	queue, _ := newQueue(t)
	policy := jobs.Policy{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond}

	if _, err := queue.Enqueue(t.Context(), kind, "korenaar", nil); err != nil {
		t.Fatal(err)
	}
	job, _ := queue.Claim(t.Context())
	if err := queue.Retry(t.Context(), job, policy, errors.New("not this time")); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if again, err := queue.Claim(t.Context()); err == nil {
			if again.Attempts != 2 {
				t.Errorf("attempts = %d, want 2 on the second run", again.Attempts)
			}
			return
		}
	}
	t.Error("the job never became claimable again")
}

func TestRepeatedFailureIsRecorded(t *testing.T) {
	queue, db := newQueue(t)
	policy := jobs.Policy{MaxAttempts: 1, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond}

	id, err := queue.Enqueue(t.Context(), kind, "korenaar", nil)
	if err != nil {
		t.Fatal(err)
	}
	job, _ := queue.Claim(t.Context())

	if err := queue.Retry(t.Context(), job, policy, errors.New("the remote is gone for good")); err != nil {
		t.Fatal(err)
	}

	failed, _ := queue.CountByState(t.Context(), "failed")
	if failed != 1 {
		t.Fatalf("failed = %d, want 1", failed)
	}

	row, err := db.Queries.GetJob(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}
	if row.LastError == nil || !strings.Contains(*row.LastError, "gone for good") {
		t.Errorf("last error = %v, want the reason", row.LastError)
	}
}

func TestRecordedReasonCarriesNoSecret(t *testing.T) {
	queue, db := newQueue(t)
	policy := jobs.Policy{MaxAttempts: 1, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond}

	id, _ := queue.Enqueue(t.Context(), kind, "korenaar", nil)
	job, _ := queue.Claim(t.Context())

	reason := errors.New("fatal: could not read from https://bot:ghp_topsecret@github.com/a/b.git")
	if err := queue.Retry(t.Context(), job, policy, reason); err != nil {
		t.Fatal(err)
	}

	row, err := db.Queries.GetJob(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}
	if row.LastError == nil {
		t.Fatal("no reason was recorded")
	}
	if strings.Contains(*row.LastError, "ghp_topsecret") {
		t.Errorf("the recorded reason carries a secret: %q", *row.LastError)
	}
}

func TestWorkerRunsAJob(t *testing.T) {
	queue, _ := newQueue(t)
	worker := jobs.NewWorker(queue, jobs.WorkerOptions{Logger: quiet()})

	ran := make(chan jobs.Job, 1)
	worker.Handle(kind, func(ctx context.Context, job jobs.Job) error {
		ran <- job
		return nil
	})

	if _, err := queue.Enqueue(t.Context(), kind, "korenaar", map[string]string{"site": "korenaar"}); err != nil {
		t.Fatal(err)
	}

	did, err := worker.RunOnce(t.Context())
	if err != nil || !did {
		t.Fatalf("RunOnce() = %v, %v, want it to have run one", did, err)
	}

	job := <-ran
	if job.Subject != "korenaar" {
		t.Errorf("subject = %q, want korenaar", job.Subject)
	}
	if !strings.Contains(string(job.Payload), "korenaar") {
		t.Errorf("payload = %q, want what was queued", job.Payload)
	}

	done, _ := queue.CountByState(t.Context(), "done")
	if done != 1 {
		t.Errorf("done = %d, want 1", done)
	}
}

func TestWorkerRetriesAFailingJob(t *testing.T) {
	queue, _ := newQueue(t)
	worker := jobs.NewWorker(queue, jobs.WorkerOptions{
		Logger: quiet(),
		Policy: jobs.Policy{MaxAttempts: 2, BaseDelay: time.Hour, MaxDelay: time.Hour},
	})
	worker.Handle(kind, func(context.Context, jobs.Job) error {
		return errors.New("not today")
	})

	if _, err := queue.Enqueue(t.Context(), kind, "korenaar", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := worker.RunOnce(t.Context()); err != nil {
		t.Fatal(err)
	}

	queued, _ := queue.CountByState(t.Context(), "queued")
	done, _ := queue.CountByState(t.Context(), "done")
	if queued != 1 || done != 0 {
		t.Errorf("queued = %d, done = %d, want the job waiting to be retried", queued, done)
	}
}

func TestWorkerWithNoHandlerDoesNotLoseTheJob(t *testing.T) {
	queue, _ := newQueue(t)
	worker := jobs.NewWorker(queue, jobs.WorkerOptions{
		Logger: quiet(),
		Policy: jobs.Policy{MaxAttempts: 5, BaseDelay: time.Hour, MaxDelay: time.Hour},
	})

	if _, err := queue.Enqueue(t.Context(), "unknown-kind", "korenaar", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := worker.RunOnce(t.Context()); err != nil {
		t.Fatal(err)
	}

	queued, _ := queue.CountByState(t.Context(), "queued")
	if queued != 1 {
		t.Errorf("queued = %d, want the job kept for when a handler exists", queued)
	}
}

func TestWorkerDrainsOnShutdown(t *testing.T) {
	queue, _ := newQueue(t)
	worker := jobs.NewWorker(queue, jobs.WorkerOptions{Logger: quiet(), IdleWait: 10 * time.Millisecond})

	entered := make(chan struct{})
	release := make(chan struct{})
	worker.Handle(kind, func(context.Context, jobs.Job) error {
		close(entered)
		<-release
		return nil
	})

	if _, err := queue.Enqueue(t.Context(), kind, "korenaar", nil); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	stopped := make(chan struct{})
	go func() {
		worker.Run(ctx)
		close(stopped)
	}()

	<-entered
	cancel()

	drained := make(chan error, 1)
	go func() { drained <- worker.Drain(context.Background(), 5*time.Second) }()

	close(release)

	if err := <-drained; err != nil {
		t.Fatalf("Drain() error = %v, want nil", err)
	}
	<-stopped

	done, _ := queue.CountByState(t.Context(), "done")
	if done != 1 {
		t.Errorf("done = %d, want the in-flight job to have finished", done)
	}
}

func TestWorkerReturnsAnOverrunningJob(t *testing.T) {
	queue, _ := newQueue(t)
	worker := jobs.NewWorker(queue, jobs.WorkerOptions{Logger: quiet(), IdleWait: 10 * time.Millisecond})

	entered := make(chan struct{})
	release := make(chan struct{})
	defer close(release)

	worker.Handle(kind, func(context.Context, jobs.Job) error {
		close(entered)
		<-release
		return nil
	})

	if _, err := queue.Enqueue(t.Context(), kind, "korenaar", nil); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go worker.Run(ctx)

	<-entered
	cancel()

	if err := worker.Drain(context.Background(), 50*time.Millisecond); err != nil {
		t.Fatalf("Drain() error = %v, want nil", err)
	}

	queued, _ := queue.CountByState(t.Context(), "queued")
	done, _ := queue.CountByState(t.Context(), "done")
	if queued != 1 || done != 0 {
		t.Errorf("queued = %d, done = %d, want the unfinished job back in the queue", queued, done)
	}
}
