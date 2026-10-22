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

-- name: CountOrganisationOwners :one
SELECT count(*) FROM organisation_members WHERE organisation_id = $1 AND role = 'owner';

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
