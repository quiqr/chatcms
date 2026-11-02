---
# chatcms-7r0u
title: 02 Admin levels, limits and project moves
status: in-progress
type: milestone
priority: high
created_at: 2026-11-02T20:04:17Z
updated_at: 2026-11-02T20:04:17Z
---

## Goal

The three levels of administration the design hand-off of 2 November adds, and
the rules that come with them: a platform admin over the whole installation, an
organisation admin over one organisation, and project roles under that. Plus
optional limits per organisation, and moving a project from one organisation to
another.

Milestone 01 shipped a flatter model: an installation administrator, an
organisation owner, and four project roles including an owner. The new design
replaces it rather than extends it, so the work is a migration as much as an
addition.

## Definition of done

The screens in `docs/design/DESIGN.md` sections 5.11 to 5.16 are built at
1440 px and at 390 px, the permission table in section 6 is what
`internal/policy` answers, and every limit-guarded action is refused
server-side with the wording in section "Limits (rules)". The acceptance suite
covers each refusal.

## Source

`docs/design/` as of the hand-off dated 2 November 2026 (19 artboards, 30
rendered states). `docs/design/reference/html/hierarchy.html` is the picture of
the model; it is a reference board, not a screen to build.

## What stays as it is

Nothing in milestone 01's content, editing, publishing, history, media or sync
work changes. This milestone is about who may do what, and where projects live.
