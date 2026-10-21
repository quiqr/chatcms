-- name: EnqueueJob :one
INSERT INTO jobs (kind, subject, payload, run_at)
VALUES ($1, $2, $3, COALESCE(sqlc.narg('run_at')::timestamptz, now()))
ON CONFLICT (kind, subject) WHERE state = 'queued' DO NOTHING
RETURNING *;

-- name: GetJob :one
SELECT * FROM jobs WHERE id = $1;

-- name: ClaimNextJob :one
UPDATE jobs
SET state = 'running',
    attempts = attempts + 1,
    updated_at = now()
WHERE id = (
    SELECT id FROM jobs
    WHERE state = 'queued' AND run_at <= now()
    ORDER BY run_at, id
    FOR UPDATE SKIP LOCKED
    LIMIT 1
)
RETURNING *;

-- name: CompleteJob :exec
UPDATE jobs
SET state = 'done',
    last_error = NULL,
    updated_at = now()
WHERE id = $1;

-- name: RetryJob :exec
UPDATE jobs
SET state = 'queued',
    run_at = $2,
    last_error = $3,
    updated_at = now()
WHERE id = $1;

-- name: FailJob :exec
UPDATE jobs
SET state = 'failed',
    last_error = $2,
    updated_at = now()
WHERE id = $1;

-- name: ReleaseJob :exec
UPDATE jobs
SET state = 'queued',
    run_at = now(),
    updated_at = now()
WHERE id = $1;

-- name: CountJobsByState :one
SELECT count(*) FROM jobs WHERE state = $1;

-- name: CountJobsForSubject :one
SELECT count(*) FROM jobs WHERE kind = $1 AND subject = $2 AND state = $3;
