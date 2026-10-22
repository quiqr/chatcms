-- name: StartPublish :one
INSERT INTO publishes (project_id, person_id, note, change_count, summary)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: FinishPublish :one
UPDATE publishes
SET state = $2,
    steps = $3,
    trouble = $4,
    detail = $5,
    finished_at = now()
WHERE id = $1 AND state = 'running'
RETURNING *;

-- name: GetPublish :one
SELECT * FROM publishes WHERE id = $1;

-- name: ListPublishesForProject :many
SELECT * FROM publishes
WHERE project_id = $1
ORDER BY started_at DESC, id DESC
LIMIT $2;

-- name: LastSucceededPublish :one
SELECT * FROM publishes
WHERE project_id = $1 AND state = 'succeeded'
ORDER BY finished_at DESC
LIMIT 1;

-- name: InterruptRunningPublishes :many
UPDATE publishes
SET state = 'interrupted',
    finished_at = now(),
    trouble = $1::text
WHERE state = 'running'
RETURNING id;
