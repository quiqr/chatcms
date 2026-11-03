-- +goose Up

-- Who made an organisation. The administration record already says who did
-- it, but a record is a list of acts and the screen wants a fact about a
-- thing: "created 21 October by Pim Snel", beside its name.
--
-- Null for the organisations made before this, which is honest: nobody wrote
-- it down at the time, and inventing it from the record would be a guess
-- dressed as a fact.
ALTER TABLE organisations
    ADD COLUMN created_by UUID REFERENCES people (id) ON DELETE SET NULL;

-- +goose Down

ALTER TABLE organisations DROP COLUMN created_by;
