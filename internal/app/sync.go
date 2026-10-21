package app

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/jobs"
)

// registerSyncJob teaches the worker how to bring one site level with its
// remote. The handler is here rather than in the queue, which knows nothing
// about sites, or in the engine, which knows nothing about queues.
func registerSyncJob(worker *jobs.Worker, sites *gitsync.Sites, logger *slog.Logger) {
	worker.Handle(gitsync.SyncJobKind, func(ctx context.Context, job jobs.Job) error {
		repository, err := sites.Open(ctx, job.Subject)
		if err != nil {
			return err
		}
		defer repository.Close()

		result, err := repository.Sync(ctx)
		if err != nil {
			return err
		}

		switch {
		case result.Conflict != nil:
			logger.Info("this site was edited by two people",
				"site", job.Subject, "pages", len(result.Conflict.Paths))
		case len(result.Arrived) > 0:
			logger.Info("changes from elsewhere arrived",
				"site", job.Subject, "versions", len(result.Arrived), "replayed", result.Replayed)
		}
		return nil
	})
}

// scheduleSyncs queues a sync for every site on an interval, so changes made
// elsewhere arrive without anyone asking.
func scheduleSyncs(ctx context.Context, queue *jobs.Queue, sites *gitsync.Sites, every time.Duration, logger *slog.Logger) {
	if every <= 0 {
		logger.Warn("no sync interval is configured, so changes made elsewhere will only arrive when asked for")
		return
	}

	ticker := time.NewTicker(every)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			queueSyncs(ctx, queue, sites, logger)
		}
	}
}

func queueSyncs(ctx context.Context, queue *jobs.Queue, sites *gitsync.Sites, logger *slog.Logger) {
	ids, err := sites.List()
	if err != nil {
		logger.Error("listing sites failed", "error", err)
		return
	}

	for _, id := range ids {
		if _, err := queue.Enqueue(ctx, gitsync.SyncJobKind, id, nil); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			logger.Error("queueing a sync failed", "site", id, "error", err)
		}
	}
}
