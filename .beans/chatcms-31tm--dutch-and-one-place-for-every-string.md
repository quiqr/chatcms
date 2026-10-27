---
# chatcms-31tm
title: Dutch, and one place for every string
status: in-progress
type: epic
priority: normal
created_at: 2026-10-27T14:38:39Z
updated_at: 2026-10-27T23:00:00Z
parent: chatcms-8vem
---

## Purpose

`chatcms-6h9u` shipped the preferences a person can choose and left the
language out. The interface is English only, though
`docs/design/DESIGN.md` section 7 says to plan for Dutch next, and the
preferences screen says in as many words that Dutch is being worked on.

This is both halves of that: one place for every string, and the Dutch.

## Why it was not done with the rest

The mechanical half is a change to 62 files and some 350 distinct pieces of
text. The other half is the product's voice in a language this repository
has a style guide about, with rules the English copy already follows: no git
words, sentence case, buttons that say exactly what happens, errors that go
what happened, reassurance, one concrete fix.

Generating 350 Dutch strings wholesale would produce text that reads like a
machine wrote it, in a product whose whole argument is that it speaks to
people who are not technical. The extraction and the copy belong together,
done by somebody who can judge the result.

## Scope

- One place every string goes through, and no string written where it is
  shown. 62 files.
- A check that fails when a new string is added outside it, the way
  `check-interface` already fails on a hand written colour.
- Dutch, written or reviewed by somebody who speaks it, following the copy
  rules in both languages.
- The chosen language on the account beside the appearance, falling back to
  what the browser asks for.
- Dates follow the chosen language: `web/src/lib/when.ts` already writes
  them in one place and passes `undefined` as the locale, which means the
  browser's.

## Done when

- No string is written where it is shown, enforced by a check.
- The interface is fully readable in Dutch, with no English left behind.
- Switching takes effect without signing out, and follows to another
  machine.
- The acceptance suite runs once in Dutch, so that a missing string fails
  for us rather than for a customer.

## References

- `docs/design/DESIGN.md` section 7, the copy rules; section 11, i18n.
- `.claude` prose rules: the Dutch writing guide and its amendments.
- `web/src/lib/when.ts`, which already writes every date.
