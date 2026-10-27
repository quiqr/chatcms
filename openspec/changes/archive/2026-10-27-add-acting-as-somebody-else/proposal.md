## Why

Support means answering "I cannot see the publish button", and the only way
to answer it today is to ask the customer to describe their screen. Whoever
runs the installation needs to see, and do, what that person sees and does.

This is the one deliberate exception to the rule the administrator epic
established, that administering is not editing. It is allowed to be an
exception because it announces itself: on the screen while it happens, in
the record afterwards, and in the customer's own history forever.

Epic `chatcms-r7im`.

Bean: `.beans/chatcms-r7im--acting-as-somebody-else-visibly.md`.

## What Changes

- **Acting as somebody**, begun from the person opened on the
  administration screen and ended from a band that is on every screen while
  it is in force.
- **The session stays the administrator's**, carrying who they are acting
  as. No session is issued as the customer, because that would lose who is
  really there.
- **Permissions are the person's being acted as.** `policy.Can` answers for
  them, so "can do everything they can" is true without a second set of
  rules to keep in step.
- **Attribution keeps both.** A version saved while acting reads "Pim Snel
  as Marieke de Vries", in ChatCMS and in git.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `tenancy/administration`: acting as somebody else, and what that does to
  permission, attribution and the record.

## Impact

- **New**: a column on a session saying who is being acted as, two
  endpoints, the band, and the attribution.
- **Changed**: `auth.Service.Lookup` answers as the person being acted as
  while saying who is really there; `internal/app` builds the author of a
  version from both.

## Assumptions

- **The acted-as person is who the actor is, everywhere but the record.**
  `httpapi.Actor` carries their address and name, so every screen, every
  permission check and every listing is theirs without a single call site
  learning about this feature. Who is really there rides alongside, and only
  three places read it: the band, the record, and the author of a version.

- **Git carries the name of both and the address of the administrator.** A
  version has one author. The name reads "Pim Snel as Marieke de Vries" so
  that the customer's own repository says what happened, and the address is
  the administrator's, because that is who was really at the keyboard. The
  alternative, the customer's address, would make the trail depend on
  reading the name.

- **The author is found through the request rather than through a new
  argument.** Who is acting is already in the request's context, which the
  place that builds an author already receives. Threading an extra argument
  through every caller would put this feature in the signature of code that
  has nothing else to do with it. A version written by a background job,
  which has no request, is attributed as it was before.

- **Administrators cannot be acted as.** Acting as one would be a way of
  becoming one. An account that is off cannot be acted as either: it cannot
  be used by the person it belongs to, so it must not be usable by anybody
  else.

- **Ending needs no permission.** Beginning asks `policy.Can`; ending is
  offered to whoever is acting, because a way out that can be refused is a
  trap. Signing out ends it, and so does the session expiring, because the
  acting lives on the session.

- **While acting, administration is not offered.** `policy.Can` answers as
  the person being acted as, who does not administer the installation, so
  the administration screen and everything on it is refused until the acting
  ends. The band is the way back.

- **A publish keeps the site's own record, and the trail is elsewhere.** The
  publish record names the person whose site it is, which is what it has
  always been for, and that name is shown to nobody. Who was really there is
  in the version the publish carries, whose author names both, and between
  the two record entries that say when the acting began and ended. A column
  nobody reads would be a second trail to keep in step with the first.
