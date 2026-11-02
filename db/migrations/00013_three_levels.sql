-- +goose Up

-- Three levels of role, from the design hand-off of 2 November: a platform
-- admin over the installation, an organisation admin over one organisation,
-- and project roles under that.
--
-- The project role "owner" goes. Everybody who held it becomes an
-- implementer, which is what a project owner did day to day: the repository,
-- publishing and the technical detail. What they lose is billing and deleting
-- the project, which belongs to whoever administers the organisation now.
-- Making every project owner an organisation admin instead would hand
-- somebody who was given one project the keys to every project beside it.
ALTER TABLE project_members DROP CONSTRAINT project_members_role_known;
UPDATE project_members SET role = 'implementer' WHERE role = 'owner';
ALTER TABLE project_members ADD CONSTRAINT project_members_role_known
  CHECK (role IN ('implementer', 'editor', 'viewer'));

-- An organisation's owner is its admin, stored as it is spoken. A database
-- that says owner while every screen says admin is a difference that is cheap
-- today and expensive the first time somebody reads a query and believes it.
-- The old word has to stop being required before the rows can stop using it,
-- and the new one has to be allowed before the rows can use it: drop, update,
-- add, in that order.
ALTER TABLE organisation_members DROP CONSTRAINT organisation_members_role_known;
UPDATE organisation_members SET role = 'admin' WHERE role = 'owner';
ALTER TABLE organisation_members ADD CONSTRAINT organisation_members_role_known
  CHECK (role IN ('admin', 'member'));

-- +goose Down

ALTER TABLE organisation_members DROP CONSTRAINT organisation_members_role_known;
UPDATE organisation_members SET role = 'owner' WHERE role = 'admin';
ALTER TABLE organisation_members ADD CONSTRAINT organisation_members_role_known
  CHECK (role IN ('owner', 'member'));

ALTER TABLE project_members DROP CONSTRAINT project_members_role_known;
ALTER TABLE project_members ADD CONSTRAINT project_members_role_known
  CHECK (role IN ('owner', 'implementer', 'editor', 'viewer'));
