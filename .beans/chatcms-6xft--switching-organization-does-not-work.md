---
# chatcms-6xft
title: switching organization does not work
status: completed
type: bug
priority: high
created_at: 2026-10-27T20:46:06Z
updated_at: 2026-10-28T03:00:00Z
parent: chatcms-8vem
openspec-link: openspec/changes/archive/2026-10-27-fix-organisation-switching
---

as admin or regular user, when there are more then one orgaztions switching does not work.

## Summary of Changes

The switcher in the header drew each organisation as plain text, with no
address behind it. It listed everywhere somebody could go and took them
nowhere; typing the address was the only way across.

- Each organisation in the switcher is now a way to that organisation's
  projects, and the one being looked at carries a mark and `aria-current`.
- Both screens that draw the header say which organisation it is, so that
  the switcher knows which of its entries is the current one.
- A frontend test reads the address behind every entry, which is what was
  missing: the old acceptance test checked that both names were listed, and
  passed against a menu that did nothing.
- The acceptance test now follows the switch and reads what comes back, from
  the projects screen and from a screen that is not the projects, at 1440 px
  and at 390 px.
