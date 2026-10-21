---
# chatcms-0wba
title: Per-document permissions
status: draft
type: epic
priority: deferred
created_at: 2026-10-21T15:32:56Z
updated_at: 2026-10-21T15:32:56Z
parent: chatcms-2udf
---

## Intent

Permissions finer than a project role: who may edit which collection, which
page, which field.

## Room left for it

`internal/policy.Can(actor, action, resource)` already takes the resource
argument that milestone 01 mostly ignores. No handler compares roles itself.
