-- +goose Up
ALTER TABLE jobs ADD COLUMN subject TEXT NOT NULL DEFAULT '';

-- At most one queued job per kind and subject, so asking twice for the same
-- work while it is still waiting does not queue it twice. A job that is
-- already running does not occupy the index, so a request made during a run
-- is not lost.
CREATE UNIQUE INDEX jobs_one_queued_per_subject
    ON jobs (kind, subject)
    WHERE state = 'queued';

-- +goose Down
DROP INDEX jobs_one_queued_per_subject;
ALTER TABLE jobs DROP COLUMN subject;
