-- +goose Up
CREATE TABLE jobs (
    id          BIGSERIAL PRIMARY KEY,
    kind        TEXT        NOT NULL,
    payload     JSONB       NOT NULL DEFAULT '{}'::jsonb,
    state       TEXT        NOT NULL DEFAULT 'queued',
    attempts    INTEGER     NOT NULL DEFAULT 0,
    last_error  TEXT,
    run_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT jobs_state_known CHECK (state IN ('queued', 'running', 'done', 'failed'))
);

CREATE INDEX jobs_runnable ON jobs (run_at) WHERE state = 'queued';

-- +goose Down
DROP TABLE jobs;
