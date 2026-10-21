---
# chatcms-o9q7
title: Hugo in WASM for browser preview
status: draft
type: epic
priority: deferred
created_at: 2026-10-21T15:32:56Z
updated_at: 2026-10-21T15:32:56Z
parent: chatcms-2udf
---

## Intent

Run Hugo in the browser so preview needs no server-side session, removing the
per-session preview server and its idle shutdown.

## Room left for it

Preview is proxied through the app behind `internal/preview`, so the
server-side implementation can be swapped for a client-side one.
