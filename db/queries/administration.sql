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

-- name: ListOrganisationMembers :many
SELECT m.person_id, m.role, p.email, p.display_name, p.password_hash
FROM organisation_members m
JOIN people p ON p.id = m.person_id
WHERE m.organisation_id = $1
ORDER BY lower(p.email);

-- name: UpdateProject :one
UPDATE projects SET name = $2, domain = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: RemoveProject :exec
DELETE FROM projects WHERE id = $1;

-- name: ProjectsInOrganisation :many
SELECT id, slug, name, domain FROM projects
WHERE organisation_id = $1
ORDER BY lower(name);

-- name: ListEveryPerson :many
-- Where somebody belongs is read with the list, so that fifty people cost
-- one query rather than a hundred and fifty.
SELECT p.id,
       p.email,
       p.display_name,
       p.administers,
       p.disabled_at,
       (p.password_hash IS NOT NULL)::boolean AS has_password,
       (SELECT max(s.last_seen_at)
        FROM sessions s
        WHERE s.person_id = p.id AND s.revoked_at IS NULL)::timestamptz AS last_seen,
       COALESCE((SELECT json_agg(json_build_object('slug', o.slug, 'name', o.name, 'role', m.role)
                                 ORDER BY lower(o.name))
                 FROM organisation_members m
                 JOIN organisations o ON o.id = m.organisation_id
                 WHERE m.person_id = p.id), '[]')::json AS organisations,
       COALESCE((SELECT json_agg(json_build_object('slug', pr.slug, 'name', pr.name,
                                                   'organisation', po.slug, 'role', pm.role)
                                 ORDER BY lower(pr.name))
                 FROM project_members pm
                 JOIN projects pr ON pr.id = pm.project_id
                 JOIN organisations po ON po.id = pr.organisation_id
                 WHERE pm.person_id = p.id), '[]')::json AS projects,
       count(*) OVER () AS altogether
FROM people p
WHERE sqlc.arg('search')::text = ''
   OR p.email ILIKE '%' || sqlc.arg('search')::text || '%'
   OR p.display_name ILIKE '%' || sqlc.arg('search')::text || '%'
ORDER BY lower(p.email)
LIMIT sqlc.arg('most')::int;

-- name: PersonWithReach :one
SELECT p.id,
       p.email,
       p.display_name,
       p.administers,
       p.disabled_at,
       (p.password_hash IS NOT NULL)::boolean AS has_password,
       (SELECT max(s.last_seen_at)
        FROM sessions s
        WHERE s.person_id = p.id AND s.revoked_at IS NULL)::timestamptz AS last_seen,
       COALESCE((SELECT json_agg(json_build_object('slug', o.slug, 'name', o.name, 'role', m.role)
                                 ORDER BY lower(o.name))
                 FROM organisation_members m
                 JOIN organisations o ON o.id = m.organisation_id
                 WHERE m.person_id = p.id), '[]')::json AS organisations,
       COALESCE((SELECT json_agg(json_build_object('slug', pr.slug, 'name', pr.name,
                                                   'organisation', po.slug, 'role', pm.role)
                                 ORDER BY lower(pr.name))
                 FROM project_members pm
                 JOIN projects pr ON pr.id = pm.project_id
                 JOIN organisations po ON po.id = pr.organisation_id
                 WHERE pm.person_id = p.id), '[]')::json AS projects
FROM people p
WHERE p.id = $1;
