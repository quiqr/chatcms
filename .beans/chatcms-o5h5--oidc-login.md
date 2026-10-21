---
# chatcms-o5h5
title: OIDC login
status: draft
type: epic
priority: deferred
created_at: 2026-10-21T15:32:56Z
updated_at: 2026-10-21T15:32:56Z
parent: chatcms-2udf
---

## Intent

Single sign-on alongside the built-in email and password login.

## Room left for it

The identity source in `internal/auth` is an interface with one
implementation, and sessions do not assume a password ever existed.
