-- name: MakeAdministrator :one
UPDATE people SET administers = TRUE, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: MakeAdministratorByEmail :one
UPDATE people SET administers = TRUE, updated_at = now()
WHERE lower(email) = lower(sqlc.arg('email')::text) AND NOT administers
RETURNING *;

-- name: StopAdministrating :one
-- The last administrator is kept here as well as above, because an
-- installation with none can only be repaired with SQL.
UPDATE people AS leaving SET administers = FALSE, updated_at = now()
WHERE leaving.id = $1
  AND leaving.administers
  AND (SELECT count(*) FROM people AS others WHERE others.administers) > 1
RETURNING leaving.*;

-- name: CountAdministrators :one
SELECT count(*) FROM people WHERE administers;

-- name: ListAdministrators :many
SELECT * FROM people WHERE administers ORDER BY lower(email);

-- name: PersonByEmail :one
SELECT * FROM people WHERE lower(email) = lower(sqlc.arg('email')::text);

-- name: RecordAdministration :one
INSERT INTO administration_records (person_id, person_said, did, subject, subject_said, detail)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListAdministrationRecords :many
SELECT * FROM administration_records
ORDER BY at DESC, id DESC
LIMIT $1;

-- name: ListEveryOrganisation :many
SELECT o.*, (SELECT count(*) FROM projects p WHERE p.organisation_id = o.id) AS project_count,
            (SELECT count(*) FROM organisation_members m WHERE m.organisation_id = o.id) AS people_count
FROM organisations o
ORDER BY lower(o.name);

-- name: RenameOrganisation :one
UPDATE organisations SET name = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: RemoveOrganisation :exec
DELETE FROM organisations WHERE id = $1;
