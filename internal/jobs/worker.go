package jobs

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
)

// Worker takes jobs off the queue and runs them. It stops claiming when it is
// asked to stop, and returns work it could not finish rather than losing it.
type Worker struct {
	queue    *Queue
	logger   *slog.Logger
	policy   Policy
	idleWait time.Duration

	mu       sync.RWMutex
	handlers map[string]Handler

	// running is the job in flight, so shutdown can return it to the queue if
	// it does not finish in time.
	runningMu sync.Mutex
	running   *Job
}

type WorkerOptions struct {
	Logger   *slog.Logger
	Policy   Policy
	IdleWait time.Duration
}

func NewWorker(queue *Queue, opts WorkerOptions) *Worker {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}
	policy := opts.Policy
	if policy.MaxAttempts == 0 {
		policy = DefaultPolicy()
	}
	idle := opts.IdleWait
	if idle == 0 {
		idle = time.Second
	}

	return &Worker{
		queue:    queue,
		logger:   logger,
		policy:   policy,
		idleWait: idle,
		handlers: map[string]Handler{},
	}
}

// Handle registers what to do for one kind of job.
func (w *Worker) Handle(kind string, handler Handler) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.handlers[kind] = handler
}

func (w *Worker) handlerFor(kind string) (Handler, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	handler, ok := w.handlers[kind]
	return handler, ok
}

// Run claims and runs jobs until the context is done. It returns once the job
// it is running has finished or the context's deadline has passed, whichever
// is first; an unfinished job is returned to the queue by Drain.
func (w *Worker) Run(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}

		ran, err := w.runOne(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			w.logger.Error("claiming work failed", "error", err)
		}
		if ran {
			continue
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(w.idleWait):
		}
	}
}

// RunOnce claims and runs a single job, reporting whether there was one.
func (w *Worker) RunOnce(ctx context.Context) (bool, error) {
	return w.runOne(ctx)
}

func (w *Worker) runOne(ctx context.Context) (bool, error) {
	job, err := w.queue.Claim(ctx)
	if errors.Is(err, ErrNothingToDo) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	w.setRunning(&job)
	defer w.setRunning(nil)

	handler, ok := w.handlerFor(job.Kind)
	if !ok {
		w.logger.Error("no handler for this kind of work", "kind", job.Kind, "job", job.ID)
		return true, w.queue.Retry(context.WithoutCancel(ctx), job, w.policy,
			errors.New("there is no handler for "+job.Kind))
	}

	// A job runs with a context that is not cancelled by shutdown, so that
	// stopping the server does not tear a job in half. Drain decides how long
	// to wait for it.
	runCtx := context.WithoutCancel(ctx)

	if err := handler(runCtx, job); err != nil {
		w.logger.Warn("work failed",
			"kind", job.Kind, "job", job.ID, "attempt", job.Attempts, "error", err)
		return true, w.queue.Retry(runCtx, job, w.policy, err)
	}

	return true, w.queue.Complete(runCtx, job.ID)
}

func (w *Worker) setRunning(job *Job) {
	w.runningMu.Lock()
	defer w.runningMu.Unlock()
	w.running = job
}

func (w *Worker) current() *Job {
	w.runningMu.Lock()
	defer w.runningMu.Unlock()
	return w.running
}

// Drain waits for the job in flight to finish, and returns it to the queue if
// it does not finish within the budget.
func (w *Worker) Drain(ctx context.Context, budget time.Duration) error {
	deadline := time.Now().Add(budget)

	for time.Now().Before(deadline) {
		job := w.current()
		if job == nil {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}

	job := w.current()
	if job == nil {
		return nil
	}

	w.logger.Warn("returning unfinished work to the queue", "kind", job.Kind, "job", job.ID)
	return w.queue.Release(context.WithoutCancel(ctx), job.ID)
}
