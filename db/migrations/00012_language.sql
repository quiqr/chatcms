-- +goose Up

-- Which language somebody reads the interface in, kept with the account the
-- way the appearance is, so that it follows them to another machine.
--
-- Empty means they have chosen nothing, and the interface follows what their
-- browser asks for. A Dutch customer should be read to in Dutch before they
-- have found a preferences screen, not after.
ALTER TABLE people ADD COLUMN language TEXT NOT NULL DEFAULT '';

ALTER TABLE people ADD CONSTRAINT people_language_known
  CHECK (language IN ('', 'en', 'nl'));

-- +goose Down

ALTER TABLE people DROP CONSTRAINT people_language_known;
ALTER TABLE people DROP COLUMN language;
