-- name: CreateOrganisation :one
INSERT INTO organisations (slug, name) VALUES ($1, $2) RETURNING *;

-- name: GetOrganisationBySlug :one
SELECT * FROM organisations WHERE slug = $1;

-- name: GetOrganisation :one
SELECT * FROM organisations WHERE id = $1;

-- name: CreateProject :one
INSERT INTO projects (organisation_id, slug, name, domain) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetProjectBySlug :one
SELECT projects.* FROM projects
JOIN organisations ON organisations.id = projects.organisation_id
WHERE organisations.slug = $1 AND projects.slug = $2;

-- name: GetProject :one
SELECT * FROM projects WHERE id = $1;

-- name: ListProjectsInOrganisation :many
SELECT * FROM projects WHERE organisation_id = $1 ORDER BY name;

-- name: SetOrganisationMember :exec
INSERT INTO organisation_members (organisation_id, person_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (organisation_id, person_id) DO UPDATE SET role = EXCLUDED.role;

-- name: RemoveOrganisationMember :exec
DELETE FROM organisation_members WHERE organisation_id = $1 AND person_id = $2;

-- name: CountOrganisationAdmins :one
SELECT count(*) FROM organisation_members WHERE organisation_id = $1 AND role = 'admin';

-- name: SetProjectMember :exec
INSERT INTO project_members (project_id, person_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (project_id, person_id) DO UPDATE SET role = EXCLUDED.role;

-- name: RemoveProjectMember :exec
DELETE FROM project_members WHERE project_id = $1 AND person_id = $2;

-- name: CountProjectAdministrators :one
SELECT count(*) FROM project_members
WHERE project_id = $1 AND role IN ('owner', 'implementer');

-- name: ListProjectMembers :many
SELECT project_members.role, people.id, people.email, people.display_name,
       (people.password_hash IS NOT NULL)::boolean AS has_password
FROM project_members
JOIN people ON people.id = project_members.person_id
WHERE project_members.project_id = $1
ORDER BY people.display_name, people.email;

-- name: ListOrganisationMembershipsForPerson :many
SELECT organisation_id, role FROM organisation_members WHERE person_id = $1;

-- name: ListProjectMembershipsForPerson :many
SELECT project_members.project_id, project_members.role, projects.organisation_id
FROM project_members
JOIN projects ON projects.id = project_members.project_id
WHERE project_members.person_id = $1;

-- name: ListOrganisationsForPerson :many
SELECT DISTINCT organisations.*
FROM organisations
LEFT JOIN organisation_members
       ON organisation_members.organisation_id = organisations.id
      AND organisation_members.person_id = $1
LEFT JOIN projects ON projects.organisation_id = organisations.id
LEFT JOIN project_members
       ON project_members.project_id = projects.id
      AND project_members.person_id = $1
WHERE organisation_members.person_id IS NOT NULL OR project_members.person_id IS NOT NULL
ORDER BY organisations.name;

-- name: SetProjectDeployKey :exec
UPDATE projects
SET deploy_public_key = $2,
    deploy_private_key = $3,
    updated_at = now()
WHERE id = $1;

-- name: SetProjectSource :exec
UPDATE projects
SET remote = $2,
    source_kind = $3,
    source_address = $4,
    updated_at = now()
WHERE id = $1;

-- name: SetProjectPublishing :exec
UPDATE projects
SET publish_to = $2,
    publish_branch = $3,
    publish_form = $4,
    updated_at = now()
WHERE id = $1;

-- name: OrganisationUsage :one
-- What an organisation holds and what it is allowed, in one answer, so that
-- no two screens can disagree about how full it is. A membership held or
-- waiting is a seat; an archived site is not a site.
SELECT o.max_members,
       o.max_projects,
       (SELECT count(*) FROM organisation_members m WHERE m.organisation_id = o.id) AS members,
       (SELECT count(*) FROM projects p
        WHERE p.organisation_id = o.id AND p.archived_at IS NULL) AS projects
FROM organisations o
WHERE o.id = $1;

-- name: SetOrganisationLimits :one
UPDATE organisations
SET max_members = $2, max_projects = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListOrganisationMembersWithTheirProjects :many
-- Everybody in an organisation, with the sites inside it that each one holds
-- a role on, read with the list so that twenty members cost one query rather
-- than twenty-one.
SELECT m.person_id,
       m.role,
       p.email,
       p.display_name,
       p.password_hash,
       (SELECT max(s.last_seen_at)
        FROM sessions s
        WHERE s.person_id = m.person_id AND s.revoked_at IS NULL)::timestamptz AS last_seen,
       COALESCE((SELECT json_agg(json_build_object('slug', pr.slug, 'name', pr.name, 'role', pm.role)
                                 ORDER BY lower(pr.name))
                 FROM project_members pm
                 JOIN projects pr ON pr.id = pm.project_id
                 WHERE pm.person_id = m.person_id
                   AND pr.organisation_id = m.organisation_id
                   AND pr.archived_at IS NULL), '[]')::json AS projects
FROM organisation_members m
JOIN people p ON p.id = m.person_id
WHERE m.organisation_id = $1
ORDER BY lower(p.email);

-- name: RemoveEveryProjectRoleInOrganisation :exec
-- Taking somebody out of an organisation takes their roles on its sites with
-- them: a role on a project in an organisation nobody is in any more is a
-- reach no screen shows.
DELETE FROM project_members
WHERE person_id = $2
  AND project_id IN (SELECT id FROM projects WHERE organisation_id = $1);

-- name: MoveProject :exec
UPDATE projects SET organisation_id = $2, updated_at = now() WHERE id = $1;

-- name: ArchiveProject :exec
UPDATE projects SET archived_at = now(), updated_at = now() WHERE id = $1;

-- name: BringProjectBack :exec
UPDATE projects SET archived_at = NULL, updated_at = now() WHERE id = $1;

-- name: ListProjectMemberIDs :many
SELECT person_id FROM project_members WHERE project_id = $1;
