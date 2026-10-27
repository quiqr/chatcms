-- +goose Up
-- Acting as somebody else keeps the administrator's own session and says, on
-- it, who they are acting as. Issuing a session as the customer would lose
-- who is really there, which is the whole trail.
ALTER TABLE sessions
    ADD COLUMN acting_as UUID REFERENCES people (id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE sessions
    DROP COLUMN acting_as;
