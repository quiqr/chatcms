-- name: StartBuild :one
INSERT INTO builds (project_id, reason)
VALUES ($1, $2)
RETURNING *;

-- name: FinishBuild :one
UPDATE builds
SET state = $2,
    exit_code = $3,
    log = $4,
    finished_at = now()
WHERE id = $1 AND state = 'running'
RETURNING *;

-- name: GetBuild :one
SELECT * FROM builds WHERE id = $1;

-- name: ListBuildsForProject :many
SELECT * FROM builds
WHERE project_id = $1
ORDER BY started_at DESC, id DESC
LIMIT $2;

-- name: InterruptRunningBuilds :many
UPDATE builds
SET state = 'interrupted',
    finished_at = now(),
    log = CASE WHEN log = '' THEN $1::text ELSE log || E'\n' || $1::text END
WHERE state = 'running'
RETURNING id;
