-- name: EnqueueJob :one
INSERT INTO jobs (kind, payload, run_at)
VALUES ($1, $2, COALESCE($3, now()))
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
    ORDER BY run_at
    FOR UPDATE SKIP LOCKED
    LIMIT 1
)
RETURNING *;

-- name: FinishJob :exec
UPDATE jobs
SET state = $2,
    last_error = $3,
    updated_at = now()
WHERE id = $1;

-- name: CountJobsByState :one
SELECT count(*) FROM jobs WHERE state = $1;
