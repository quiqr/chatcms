## Why

An administrator sees every organisation and nobody at all. There is no list
of the people on an installation, so "who is this address, and what can they
reach" can only be answered with a database client. Turning an account off
exists and is reachable only by typing an address into a box. Only
administrators can be listed, which answers a different question.

Epic `chatcms-294f`.

Bean: `.beans/chatcms-294f--administration-every-person-and-what-they-can-reac.md`.

## What Changes

- **Every person on the installation**, listed with their name, their
  address, whether their account is on, when they were last here, and where
  they belong.
- **One person opened**: the same facts, the organisations they are in with
  the role they hold, the sites they have a role on, and turning their
  account off or on again from there.
- **Finding somebody** by name or address, because the list grows with every
  customer.
- **Making a person**, which is inviting somebody without having to pick an
  organisation or a site first. It sends them the same way in and sets no
  password.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `tenancy/administration`: reading every person and what each can reach,
  and making one.

## Impact

- **New**: an endpoint that lists every person with where they belong, an
  endpoint that reads one, an endpoint that makes one, and the screens.
- **Changed**: `internal/admin` gains them; the administration screen gains
  a list of people beside its organisations.

## Assumptions

- **No administrator sets a password.** A password somebody else typed is
  one nobody can prove was never written down. Making a person sends the
  same invitation the rest of the product sends, and the person chooses
  their own. `chatcms-r7im` covers the case this would otherwise be used
  for: seeing what a customer sees.

- **The list answers in one request.** Where each person belongs is read
  with the list itself, not with a request per person, because fifty people
  must not become a hundred and fifty requests. The database answers it in
  one query over the two membership tables.

- **Searching is done by the database, not in the browser.** It takes a
  thousand customers to make this matter and a day's work to change later,
  so the request takes the search text from the start.

- **This lists people, not what they did.** The history of administration is
  its own screen and stays there.
