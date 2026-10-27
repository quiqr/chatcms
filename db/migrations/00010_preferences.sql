-- +goose Up

-- How somebody likes the application to behave, kept with the account so
-- that it follows them to another machine rather than being lost the first
-- time they use one.
--
-- Empty means they have chosen nothing, and the application follows what
-- their machine asks for. A product that depends on people having chosen is
-- a product that looks broken until they do.
ALTER TABLE people ADD COLUMN appearance TEXT NOT NULL DEFAULT '';

ALTER TABLE people ADD CONSTRAINT people_appearance_known
  CHECK (appearance IN ('', 'light', 'dark'));

-- +goose Down

ALTER TABLE people DROP CONSTRAINT people_appearance_known;
ALTER TABLE people DROP COLUMN appearance;
