-- +goose Up

-- Addresses are compared without regard to case, so that one person cannot
-- become two by capitalising differently.
CREATE TABLE people (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT        NOT NULL,
    display_name  TEXT        NOT NULL DEFAULT '',
    password_hash TEXT,
    disabled_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT people_email_not_empty CHECK (email <> '')
);

CREATE UNIQUE INDEX people_email_unique ON people (lower(email));

-- The cookie carries a random value; this holds only its hash, so a copy of
-- the database yields no usable session.
CREATE TABLE sessions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash   BYTEA       NOT NULL,
    person_id    UUID        NOT NULL REFERENCES people (id) ON DELETE CASCADE,
    remembered   BOOLEAN     NOT NULL DEFAULT false,
    expires_at   TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX sessions_token_unique ON sessions (token_hash);
CREATE INDEX sessions_by_person ON sessions (person_id) WHERE revoked_at IS NULL;

-- An invitation and a password reset are the same mechanism with different
-- lifetimes and different wording.
CREATE TABLE credential_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kind       TEXT        NOT NULL,
    token_hash BYTEA       NOT NULL,
    person_id  UUID        NOT NULL REFERENCES people (id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    spent_at   TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT credential_tokens_kind_known CHECK (kind IN ('invitation', 'password-reset'))
);

CREATE UNIQUE INDEX credential_tokens_token_unique ON credential_tokens (token_hash);
CREATE INDEX credential_tokens_by_person ON credential_tokens (person_id) WHERE spent_at IS NULL;

-- +goose Down
DROP TABLE credential_tokens;
DROP TABLE sessions;
DROP TABLE people;
