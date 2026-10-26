package app

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/quiqr/chatcms/internal/jobs"
	"github.com/quiqr/chatcms/internal/mail"
)

// registerMailJob teaches the worker how to send one message. The handler is
// here rather than in the queue, which knows nothing about mail, or in mail,
// which knows nothing about queues.
func registerMailJob(worker *jobs.Worker, settings mail.Settings, logger *slog.Logger) {
	sender := mail.NewSMTP(settings)

	worker.Handle(mail.JobKind, func(ctx context.Context, job jobs.Job) error {
		var letter mail.Letter
		if err := json.Unmarshal(job.Payload, &letter); err != nil {
			// A message nobody can read will never become readable, so it is
			// said out loud and dropped rather than retried forever.
			logger.ErrorContext(ctx, "a queued message could not be read", "error", err)
			return nil
		}
		return mail.Deliver(ctx, sender, letter, logger)
	})
}
