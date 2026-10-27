---
# chatcms-6h9u
title: 'Account preferences: language, dates, and a dark appearance'
status: todo
type: epic
priority: normal
created_at: 2026-10-27T13:06:37Z
updated_at: 2026-10-27T13:06:37Z
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
