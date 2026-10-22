-- +goose Up

CREATE TABLE organisations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug       TEXT        NOT NULL,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT organisations_slug_shape CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$')
);

CREATE UNIQUE INDEX organisations_slug_unique ON organisations (slug);

CREATE TABLE projects (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organisation_id UUID        NOT NULL REFERENCES organisations (id) ON DELETE CASCADE,
    slug            TEXT        NOT NULL,
    name            TEXT        NOT NULL,
    domain          TEXT        NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT projects_slug_shape CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$')
);

-- An address only has to be unique inside its organisation.
CREATE UNIQUE INDEX projects_slug_unique ON projects (organisation_id, slug);

CREATE TABLE organisation_members (
    organisation_id UUID        NOT NULL REFERENCES organisations (id) ON DELETE CASCADE,
    person_id       UUID        NOT NULL REFERENCES people (id) ON DELETE CASCADE,
    role            TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (organisation_id, person_id),
    CONSTRAINT organisation_members_role_known CHECK (role IN ('owner', 'member'))
);

CREATE TABLE project_members (
    project_id UUID        NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    person_id  UUID        NOT NULL REFERENCES people (id) ON DELETE CASCADE,
    role       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (project_id, person_id),
    CONSTRAINT project_members_role_known CHECK (role IN ('owner', 'implementer', 'editor', 'viewer'))
);

CREATE INDEX organisation_members_by_person ON organisation_members (person_id);
CREATE INDEX project_members_by_person ON project_members (person_id);

-- +goose Down
DROP TABLE project_members;
DROP TABLE organisation_members;
DROP TABLE projects;
DROP TABLE organisations;
