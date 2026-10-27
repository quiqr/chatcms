-- +goose Up

-- A person's picture lives here rather than in a site's repository or in the
-- state directory. Not the repository: it belongs to the account, crosses
-- organisations, and a customer's published site must never carry a picture
-- of somebody from another tenant. Not the state directory: that is
-- per-instance and not in the backup story, so a picture there would be lost
-- on the first restore.
--
-- It is one small square, made on the way in, so the column holds kilobytes
-- rather than what somebody's phone produced.
ALTER TABLE people ADD COLUMN picture BYTEA;
ALTER TABLE people ADD COLUMN picture_kind TEXT NOT NULL DEFAULT '';

-- An address waiting to be proven. Signing in uses the one in email until
-- the link sent to this one is followed, so that a typo cannot lock somebody
-- out of their own account.
ALTER TABLE people ADD COLUMN email_waiting TEXT;

CREATE UNIQUE INDEX people_email_waiting_unique
  ON people (lower(email_waiting)) WHERE email_waiting IS NOT NULL;

-- A link that proves a new email address is a third kind of link, and the
-- table names the kinds it will take.
ALTER TABLE credential_tokens DROP CONSTRAINT credential_tokens_kind_known;
ALTER TABLE credential_tokens ADD CONSTRAINT credential_tokens_kind_known
  CHECK (kind IN ('invitation', 'password-reset', 'new-address'));

-- +goose Down

ALTER TABLE credential_tokens DROP CONSTRAINT credential_tokens_kind_known;
ALTER TABLE credential_tokens ADD CONSTRAINT credential_tokens_kind_known
  CHECK (kind IN ('invitation', 'password-reset'));

DROP INDEX IF EXISTS people_email_waiting_unique;
ALTER TABLE people DROP COLUMN email_waiting;
ALTER TABLE people DROP COLUMN picture_kind;
ALTER TABLE people DROP COLUMN picture;
