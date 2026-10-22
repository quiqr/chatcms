-- +goose Up

CREATE TABLE publishes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    person_id   UUID REFERENCES people (id) ON DELETE SET NULL,
    note        TEXT NOT NULL DEFAULT '',
    state       TEXT NOT NULL DEFAULT 'running',
    change_count INTEGER NOT NULL DEFAULT 0,
    summary     JSONB NOT NULL DEFAULT '[]'::jsonb,
    steps       JSONB NOT NULL DEFAULT '[]'::jsonb,
    trouble     TEXT NOT NULL DEFAULT '',
    detail      TEXT NOT NULL DEFAULT '',
    started_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ,
    CONSTRAINT publishes_state_known
        CHECK (state IN ('running', 'succeeded', 'failed', 'interrupted')),
    CONSTRAINT publishes_finished_when_not_running
        CHECK ((state = 'running') = (finished_at IS NULL))
);

CREATE INDEX publishes_by_project ON publishes (project_id, started_at DESC);
CREATE UNIQUE INDEX publishes_one_at_a_time
    ON publishes (project_id) WHERE state = 'running';

-- +goose Down
DROP TABLE publishes;
