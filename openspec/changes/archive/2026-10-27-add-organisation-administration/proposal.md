## Why

The administration screen lists every organisation and offers nothing to do
with one. Renaming exists and is an administrator's, but the only way in is
to know the address of that organisation's own settings screen. Putting
somebody into an organisation exists over the API from
`chatcms-rbdb` and has no screen. Removing one is a rule the specs state and
nothing offers.

Epic `chatcms-d4f3`.

Bean: `.beans/chatcms-d4f3--administration-open-change-and-remove-an-organisat.md`.

## What Changes

- **Each organisation opens** from the administration screen: its name, its
  address, who is in it, and its projects.
- **Renaming, and giving it an owner**, from there. Both work already and
  are unreachable.
- **Removing an organisation**, which the policy allows and nothing offers.
  The confirmation names what goes and says that every site in git is
  untouched.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `tenancy/administration`: removing an organisation, and reading one in
  full.

## Impact

- **New**: an endpoint that reads one organisation whole, an endpoint that
  removes one, and the screen.
- **Changed**: `internal/admin` gains both; the administration screen's
  organisation cards open.

## Assumptions

- **Removing an organisation does not touch any git remote.** What is in git
  is the customer's, and ChatCMS forgetting an organisation must never be a
  way to lose a site. The confirmation says so in as many words, and a test
  proves the remote still holds what it held.

- **It is refused while it still holds projects.** Removing an organisation
  with sites in it would make several customers' work unreachable through
  one confirmation. The screen says which projects are in the way, so that
  removing is a deliberate sequence rather than one button with a long
  reach.

- **This is not a second way to do what an owner does.** An organisation's
  owner manages their own people, which `chatcms-9lu6` shipped. The
  administrator's screen is for an organisation with nobody left to run it,
  or one being set up, and offers the same acts for those cases rather than
  a parallel set of rules.
