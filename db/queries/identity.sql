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
