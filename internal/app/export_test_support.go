package app

import (
	"log/slog"

	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/jobs"
)

// RegisterSyncJobForTest exposes the sync handler so an integration test can
// run it without starting a whole server.
func RegisterSyncJobForTest(worker *jobs.Worker, sites *gitsync.Sites, logger *slog.Logger) {
	registerSyncJob(worker, sites, logger)
}
