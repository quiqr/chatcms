-- +goose Up

-- Administering the installation is a property of an account, not a role on
-- anything: it is held over the whole installation and over no part of it.
ALTER TABLE people ADD COLUMN administers BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX people_administers ON people (id) WHERE administers;

-- What somebody with power over every tenant did, so that it leaves a trail.
-- The subject is kept as text rather than as a reference, because the record
-- has to outlive the thing it talks about: removing an organisation is one of
-- the acts worth recording.
CREATE TABLE administration_records (
    id          BIGSERIAL PRIMARY KEY,
    person_id   UUID REFERENCES people (id) ON DELETE SET NULL,
    person_said TEXT NOT NULL DEFAULT '',
    did         TEXT NOT NULL,
    subject     TEXT NOT NULL DEFAULT '',
    subject_said TEXT NOT NULL DEFAULT '',
    detail      JSONB NOT NULL DEFAULT '{}'::jsonb,
    at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT administration_records_did_not_empty CHECK (did <> '')
);

CREATE INDEX administration_records_newest ON administration_records (at DESC, id DESC);

-- +goose Down

DROP TABLE administration_records;
DROP INDEX IF EXISTS people_administers;
ALTER TABLE people DROP COLUMN administers;
