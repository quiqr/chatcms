## Why

Somebody in two organisations cannot get from one to the other. The switcher
in the header lists them both and does nothing when either is chosen: the
menu items were drawn as text and never given an address.

The only way across today is to type the address, which is exactly what the
switcher exists to spare people.

Bean: `.beans/chatcms-6xft--switching-organization-does-not-work.md`.

## What Changes

- **Choosing an organisation in the switcher goes there**, to that
  organisation's projects.
- **The one you are in is marked**, so that the menu says where you are as
  well as where you can go.
- **A test that would have caught it**: the acceptance suite now chooses the
  other organisation and checks what the screen shows afterwards, rather
  than only that both names appear in the menu.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `tenancy/organisations`: switching between the organisations somebody
  belongs to.

## Impact

- **Changed**: the switcher in `web/src/shells/OrganisationShell.tsx`.

## Assumptions

- **Switching lands on that organisation's projects.** It is the one screen
  every organisation has, and the one whoever is switching is on their way
  to. Staying on the screen somebody is looking at would mean carrying a
  page across a boundary where it may not exist.

- **The test asserts what the screen shows, not what was clicked.** The old
  test passed against a menu that did nothing, because it only checked that
  the names were listed. The new one follows the switch and reads the
  projects that come back.
