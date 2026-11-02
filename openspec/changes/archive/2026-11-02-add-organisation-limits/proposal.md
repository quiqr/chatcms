## Why

An organisation may be sold a number of seats and a number of sites. Nothing
in the product knows that today, so the platform admin screens the design
draws have numbers to show and nothing behind them, and nothing stops an
organisation growing past what it pays for.

Epic `chatcms-qllv`.

Bean: `.beans/chatcms-qllv--limits-per-organisation-refused-before-anything-ru.md`.

## What Changes

- **Two optional limits per organisation**: how many members, and how many
  sites. A new organisation has neither.
- **Refused before it runs**, by the server, for every act that would take an
  organisation over: putting somebody in, inviting somebody, and making a
  site. Duplicating and moving a project ask the same question when those
  land.
- **The refusal names the limit and the numbers**, and says the ways out, in
  whatever language whoever is reading chose.
- **A pending invitation holds a seat**, because somebody who has been
  invited and has not accepted is somebody the organisation has taken on.
- **Usage is one function**, so the organisation's own screens and the
  platform admin's cannot disagree about how full it is.
- **A screen that would be refused says so before it is used**: the button is
  disabled and the reason stands beside it.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `tenancy/organisations`: the limits, what they bind, and how full an
  organisation is.

## Impact

- **New**: two columns, a usage reading, a guard in front of three acts, and
  an endpoint for the platform admin to set them.
- **Changed**: making a project and putting somebody into an organisation
  can now be refused for a reason that is nobody's fault.

## Assumptions

- **A limit is a number or nothing.** Not zero, and not a flag beside a
  number: a column that is null means no limit, which makes "no limit" the
  state a new organisation is in without anybody choosing it.

- **Lowering a limit below what is in use is allowed, and removes nothing.**
  The alternative is refusing to lower it until somebody has cut back, which
  would make a platform admin negotiate with the organisation before they can
  write down what was agreed. What is already there stays; nothing new is
  added until usage is under the limit again.

- **A seat is a membership, held or waiting.** Somebody invited and not yet
  accepted holds one, because the organisation has taken them on and the
  invitation can be withdrawn. Otherwise the limit could be walked past by
  inviting everybody and waiting.

- **An archived project does not hold a site.** Archiving is not built yet
  (`chatcms-hq8m`), so the count asks for projects that are not archived,
  against a column that is false for every row until that epic adds the act.

- **The guard is in front of the act, not inside the table.** A database
  constraint cannot say "that makes six members and the limit is five", and
  it would turn a refusal somebody can read into an error nobody can. The
  count and the act are in one transaction, so two people inviting at once
  cannot both get in.
