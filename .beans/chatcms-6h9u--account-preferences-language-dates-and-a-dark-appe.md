---
# chatcms-6h9u
title: 'Account preferences: language, dates, and a dark appearance'
status: completed
type: epic
priority: normal
created_at: 2026-10-27T13:06:37Z
openspec-link: openspec/changes/archive/2026-10-27-add-preferences
updated_at: 2026-10-27T14:25:42Z
parent: chatcms-8vem
---

## Purpose

Nothing about how the application behaves can be chosen. Dates are shown in
whatever the browser's locale happens to be, in four places with three
different formats; the interface is English only, though
`docs/design/DESIGN.md` section 11 says to plan for Dutch next; and there is
one appearance, with no dark one and no way to ask for either.

These are preferences rather than account settings: they say how somebody
likes things, not who they are. They are kept apart from `chatcms-clq4` so
that a picture and a date format do not end up in one screen arguing about
what it is for.

## Scope

- **Language.** English and Dutch, chosen by the person, falling back to what
  the browser asks for. Every string in the interface goes through one place;
  today they are written where they are shown. The copy rules in
  `docs/design/DESIGN.md` section 7 apply to both languages, and the Dutch
  must not acquire git words the English carefully avoids.
- **How a date and a time are written**, which is the same decision and is
  made four times today with `toLocaleString(undefined, ...)`. One helper,
  reading the chosen language.
- **A dark appearance**, and following what the machine asks for. The tokens
  in `web/src/index.css` are already one layer; this gives them a second set
  rather than scattering colours.
- **Where a preference is kept.** On the account, so it follows somebody to
  another machine, rather than in that browser's storage where it is lost the
  first time they use another.
- **A default that is not a choice.** Somebody who never opens this screen
  gets what their machine asks for, and nothing about the product should
  depend on them having chosen.

## Done when

- The interface is fully readable in Dutch, with no English left behind, and
  switching takes effect without signing out.
- No screen formats a date by itself.
- The dark appearance covers every screen, with no hand written colour, which
  `check-interface` already enforces.
- A preference chosen on one machine is in force on another.
- The acceptance suite runs once in Dutch, so that a missing string fails
  rather than being noticed by a customer.

## References

- `docs/design/DESIGN.md` section 7, the copy rules; section 11, i18n.
- `web/src/index.css`, the tokens.
- `web/scripts/check-interface.mjs`, which already refuses a hand written
  colour.

## Summary of Changes

Shipped as `openspec/changes/archive/2026-10-27-add-preferences`.

### Done

- **A dark appearance.** A second set of the same token names, which is what
  naming every colour bought: no screen changed, because no screen knows a
  colour. `check-interface` still refuses a hand written one in either.
- **Following the device is the default and not a choice.** Somebody who
  never opens the screen gets what their machine asks for. Choosing light
  wins over a dark machine, which is why the attribute is tested as well as
  the query.
- **The choice is kept with the account**, so it is in force wherever
  somebody signs in. An acceptance test chooses dark and then signs in from
  a browser that has never been used.
- **Applied before the first paint**, from what the browser remembered, so
  nobody sees the wrong appearance flash. The account is still where it is
  kept; that copy only saves the flash.
- **One way to write a date.** Four screens each decided for themselves and
  produced three formats. There is one function now, in the style the design
  asks for, and `check-interface` fails when a screen writes its own.

### Not done: Dutch

The bean also asks for the interface in Dutch. That is `chatcms-31tm`.

The mechanical half is 62 files and some 350 distinct strings. The other
half is the product's voice in a language this repository has a style guide
about, following copy rules the English already follows. Generating 350
Dutch strings wholesale would produce text that reads like a machine wrote
it, in a product whose argument is that it speaks plainly to people who are
not technical. The extraction and the copy belong together, done by somebody
who can judge the result.

The preferences screen says "English, for now. Dutch is being worked on",
rather than offering a switch that half works.
