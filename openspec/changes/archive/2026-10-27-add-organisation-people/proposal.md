## Why

`chatcms-c8em` shipped the account half of its bean and left the
organisation half. An administrator can put somebody into an organisation
and take them out, which `chatcms-rbdb` added, but an organisation's own
owner cannot: they must ask whoever runs the service every time somebody
joins. An invitation that was never accepted can be neither sent again nor
withdrawn, and the people screen shows one level where the design draws two.

Epic `chatcms-9lu6`, closing the rest of `chatcms-c8em`.

Bean: `.beans/chatcms-9lu6--organisation-people-the-owner-invites-changes-and.md`.

## What Changes

- **An organisation's owner invites somebody to it**, changes what they hold
  there, and takes them out. The same inviter a project uses, so an address
  with no account still gets a way in.
- **An invitation can be sent again, or withdrawn.** Withdrawing takes that
  person out of whatever it was for.
- **The people screen shows both levels**, with the switch the design draws.
- **When somebody was last here**, which the server can already answer and no
  screen asked for.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `tenancy/organisations`: inviting to an organisation, changing a role
  there, taking somebody out, and what an unaccepted invitation can do.

## Impact

- **New**: endpoints for an organisation's people, resending and
  withdrawing, and the switch on the people screen.
- **Changed**: `Inviter` learns to invite to an organisation as well as to a
  project; the people screen shows both levels and when each person was last
  here.

## Assumptions

- **One inviter, two places.** Inviting to an organisation and inviting to a
  project differ only in what the person is given; everything else, making an
  account that does not exist and sending the way in, is the same. Two copies
  of that would be two places for it to go wrong.

- **Withdrawing is removing, said in the words of somebody who has not
  accepted.** There is nothing else to undo: the account exists only because
  it was invited, and the role is the whole of what it was given. The screen
  says withdraw because that is what it means to the person reading it.

- **Sending again does not make a second invitation.** The old link stops
  working, because two live links to one account is one more than anybody
  needs.

- **The last owner cannot be taken out**, as the organisations spec already
  says, and that rule is now reachable from a screen rather than only from
  an administrator's.
