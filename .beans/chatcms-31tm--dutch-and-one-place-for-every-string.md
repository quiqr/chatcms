---
# chatcms-31tm
title: Dutch, and one place for every string
status: completed
type: epic
priority: normal
created_at: 2026-10-27T14:38:39Z
updated_at: 2026-10-28T02:00:00Z
parent: chatcms-8vem
openspec-link: openspec/changes/archive/2026-10-27-add-dutch-and-one-place-for-strings
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

## Summary of Changes

The interface is readable in Dutch, and no string is written where it is
shown any more.

- `web/src/strings/en.ts` is the catalogue and the type; `nl.ts` is typed
  against it, so a missing or misspelt key fails the type check rather than
  reaching a screen. 466 strings, and `web/src/lib/say.ts` is how a screen
  asks for one.
- The move itself was done with a codemod over the TypeScript syntax tree
  rather than by hand, and the sentences it cut in half were put back
  together with placeholders: "Showing {shown} of {altogether}" is one
  string, because Dutch does not put the words in that order.
- `npm run check:strings` fails when a string is written in a screen, the
  way `check-interface` fails on a colour written by hand, and reads the
  catalogue in both languages so that a version control word cannot come
  back in through a translation. It is in `nix flake check`.
- A language is chosen on the preferences screen, kept with the account
  (migration `00012_language.sql`), in force at once, and the same on
  another machine. Somebody who has chosen nothing is read to in the
  language their browser asks for, and English is the fallback.
- Dates and times go through `lib/when.ts`, which now writes them in the
  language that is in force.
- Every refusal the server makes carries a key beside its English reason,
  and `whatWentWrong` says the catalogue's version of it. A refusal nobody
  has translated appears in the server's own words rather than not at all.
- Found on the way: asking a screen to load again blanked it, which lost
  whatever somebody was typing. `useAsync` now keeps what it has while it
  refetches.
- Proved by 19 frontend tests of the catalogue, the fallback and the dates,
  two integration tests of the preference, three handler tests, and
  `test/e2e/21-in-dutch.spec.mjs` at 1440 px and at 390 px, which walks the
  screens a content creator uses in Dutch and fails on an English word left
  behind.
