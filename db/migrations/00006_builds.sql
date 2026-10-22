-- +goose Up

CREATE TABLE builds (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    reason      TEXT NOT NULL,
    state       TEXT NOT NULL DEFAULT 'running',
    exit_code   INTEGER,
    log         TEXT NOT NULL DEFAULT '',
    started_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ,
    CONSTRAINT builds_state_known
        CHECK (state IN ('running', 'succeeded', 'failed', 'out-of-time',
                         'out-of-memory', 'interrupted')),
    CONSTRAINT builds_finished_when_not_running
        CHECK ((state = 'running') = (finished_at IS NULL))
);

CREATE INDEX builds_by_project ON builds (project_id, started_at DESC);
CREATE INDEX builds_running ON builds (state) WHERE state = 'running';

-- +goose Down
DROP TABLE builds;
