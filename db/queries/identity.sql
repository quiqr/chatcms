-- name: CreatePerson :one
INSERT INTO people (email, display_name, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetPersonByEmail :one
SELECT * FROM people WHERE lower(email) = lower($1);

-- name: GetPerson :one
SELECT * FROM people WHERE id = $1;

-- name: SetPersonPassword :exec
UPDATE people
SET password_hash = $2,
    updated_at = now()
WHERE id = $1;

-- name: CreateSession :one
INSERT INTO sessions (token_hash, person_id, remembered, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetSessionByTokenHash :one
SELECT sqlc.embed(sessions), sqlc.embed(people)
FROM sessions
JOIN people ON people.id = sessions.person_id
WHERE sessions.token_hash = $1;

-- name: TouchSession :exec
UPDATE sessions SET last_seen_at = now() WHERE id = $1;

-- name: RevokeSession :exec
UPDATE sessions SET revoked_at = now() WHERE id = $1 AND revoked_at IS NULL;

-- name: RevokeSessionByTokenHash :exec
UPDATE sessions SET revoked_at = now() WHERE token_hash = $1 AND revoked_at IS NULL;

-- name: RevokeSessionsForPerson :exec
UPDATE sessions SET revoked_at = now() WHERE person_id = $1 AND revoked_at IS NULL;

-- name: CountLiveSessionsForPerson :one
SELECT count(*) FROM sessions
WHERE person_id = $1 AND revoked_at IS NULL AND expires_at > now();

-- name: CreateCredentialToken :one
INSERT INTO credential_tokens (kind, token_hash, person_id, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetCredentialTokenByHash :one
SELECT * FROM credential_tokens WHERE token_hash = $1;

-- name: SpendCredentialToken :exec
UPDATE credential_tokens SET spent_at = now() WHERE id = $1 AND spent_at IS NULL;

-- name: TurnAccountOff :one
UPDATE people SET disabled_at = now(), updated_at = now()
WHERE id = $1 AND disabled_at IS NULL
RETURNING *;

-- name: TurnAccountOn :one
UPDATE people SET disabled_at = NULL, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SetDisplayName :one
UPDATE people SET display_name = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: LastSeenFor :one
SELECT max(last_seen_at)::timestamptz AS last_seen
FROM sessions
WHERE person_id = $1 AND revoked_at IS NULL;

-- name: SpendEveryTokenFor :exec
-- Sending an invitation again stops the previous way in: two live links to
-- one account is one more than anybody needs.
UPDATE credential_tokens SET spent_at = now()
WHERE person_id = $1 AND kind = $2 AND spent_at IS NULL;

-- name: SetPicture :exec
UPDATE people SET picture = $2, picture_kind = $3, updated_at = now()
WHERE id = $1;

-- name: ReadPicture :one
SELECT picture, picture_kind FROM people WHERE id = $1;

-- name: SetEmailWaiting :exec
UPDATE people SET email_waiting = $2, updated_at = now()
WHERE id = $1;

-- name: ConfirmEmailWaiting :one
UPDATE people
SET email = email_waiting, email_waiting = NULL, updated_at = now()
WHERE id = $1 AND email_waiting IS NOT NULL
RETURNING *;

-- name: SessionsForPerson :many
SELECT id, last_seen_at, created_at, remembered, token_hash
FROM sessions
WHERE person_id = $1 AND revoked_at IS NULL AND expires_at > now()
ORDER BY last_seen_at DESC;

-- name: RevokeSessionForPerson :exec
UPDATE sessions SET revoked_at = now()
WHERE id = $1 AND person_id = $2 AND revoked_at IS NULL;
