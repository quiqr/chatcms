-- +goose Up

ALTER TABLE projects
    ADD COLUMN remote            TEXT NOT NULL DEFAULT '',
    ADD COLUMN source_kind       TEXT NOT NULL DEFAULT 'none',
    ADD COLUMN source_address    TEXT NOT NULL DEFAULT '',
    ADD COLUMN deploy_public_key TEXT NOT NULL DEFAULT '',
    -- The private half is sealed before it is written; nothing readable is
    -- stored. See internal/secrets.
    ADD COLUMN deploy_private_key BYTEA,
    ADD COLUMN publish_to        TEXT NOT NULL DEFAULT 'main-line',
    ADD COLUMN publish_branch    TEXT NOT NULL DEFAULT '',
    ADD COLUMN publish_form      TEXT NOT NULL DEFAULT 'source',
    ADD CONSTRAINT projects_source_kind_known
        CHECK (source_kind IN ('none', 'repository', 'export', 'template')),
    ADD CONSTRAINT projects_publish_to_known
        CHECK (publish_to IN ('main-line', 'its-own-line')),
    ADD CONSTRAINT projects_publish_form_known
        CHECK (publish_form IN ('source', 'built'));

-- +goose Down
ALTER TABLE projects
    DROP CONSTRAINT projects_publish_form_known,
    DROP CONSTRAINT projects_publish_to_known,
    DROP CONSTRAINT projects_source_kind_known,
    DROP COLUMN publish_form,
    DROP COLUMN publish_branch,
    DROP COLUMN publish_to,
    DROP COLUMN deploy_private_key,
    DROP COLUMN deploy_public_key,
    DROP COLUMN source_address,
    DROP COLUMN source_kind,
    DROP COLUMN remote;
