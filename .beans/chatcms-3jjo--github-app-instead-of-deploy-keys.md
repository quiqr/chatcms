---
# chatcms-3jjo
title: GitHub App instead of deploy keys
status: draft
type: epic
priority: deferred
created_at: 2026-10-21T15:32:56Z
updated_at: 2026-10-21T15:32:56Z
parent: chatcms-2udf
---

## Intent

Replace per-project SSH deploy keys with a GitHub App installation, so
implementers do not have to paste a public key into the git host.

## Room left for it

Credentials are held behind `internal/secrets` and the repository layer takes
its authentication from configuration, not from a hard-coded key path.
