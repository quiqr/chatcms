---
# chatcms-0fml
title: 04 The old behaviour, proven against the application
status: in-progress
type: milestone
priority: high
created_at: 2026-11-03T20:16:31Z
updated_at: 2026-11-03T20:16:31Z
---

## Goal

Prove, against the application itself, that ChatCMS does to a value what the
old Quiqr does to it. Milestone 01 made the promise and built the first half
of the evidence: `scripts/generate-goldens.sh` runs the two serialiser
libraries the old app uses, so how bytes are written is settled by
observation. The other half, what a form does to a value when somebody edits
it, is settled by our reading of the old app's source, which is a belief.

## Why it exists now

The first real customer's site comes out of the old Quiqr. Every field type
their model uses is a place where a belief could be wrong, and a wrong one
shows up as a diff in their repository rather than as a failing test.
`internal/model` decodes all 28 field types the old application draws;
`web/src/fields/Fields.tsx` draws 17 and falls back to a text box for the
rest. Which of those fallbacks are harmless and which lose data is not
currently knowable.

## Definition of done

An answer about a field's behaviour can be produced by running the old
application rather than by reading its source, and the answers that matter
are recorded in the repository so the gate keeps them. Where ChatCMS and the
reference disagree, the disagreement is written down and deliberate.

## What stays as it is

The serialisation goldens and the way they are made. This milestone is the
layer above them: the same discipline, applied to what a field does rather
than to how bytes are written.
