-- +goose Up

-- What an organisation is allowed: how many members, and how many sites.
-- Null is no limit, which is what a new organisation has and what most of
-- them will keep. A number beside a flag would be two things to read and two
-- ways to be wrong.
--
-- A limit set below what an organisation already holds removes nothing. It
-- stops what is new until usage is under it again, so that writing down what
-- was agreed is never a negotiation about what to throw away first.
ALTER TABLE organisations
    ADD COLUMN max_members  INTEGER,
    ADD COLUMN max_projects INTEGER;

ALTER TABLE organisations
    ADD CONSTRAINT organisations_max_members_positive
        CHECK (max_members IS NULL OR max_members > 0),
    ADD CONSTRAINT organisations_max_projects_positive
        CHECK (max_projects IS NULL OR max_projects > 0);

-- An archived site does not hold one of the sites an organisation is allowed.
-- Archiving itself is chatcms-hq8m; the column is here because the count that
-- the limit reads has to ask the question from the start, rather than being
-- rewritten the day it becomes possible to answer.
ALTER TABLE projects ADD COLUMN archived_at TIMESTAMPTZ;

CREATE INDEX projects_live_by_organisation
    ON projects (organisation_id) WHERE archived_at IS NULL;

-- +goose Down

DROP INDEX projects_live_by_organisation;
ALTER TABLE projects DROP COLUMN archived_at;

ALTER TABLE organisations
    DROP CONSTRAINT organisations_max_members_positive,
    DROP CONSTRAINT organisations_max_projects_positive;

ALTER TABLE organisations
    DROP COLUMN max_members,
    DROP COLUMN max_projects;
