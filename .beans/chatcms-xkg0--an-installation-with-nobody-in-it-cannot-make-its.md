---
# chatcms-xkg0
title: An installation with nobody in it cannot make its first account
status: todo
type: bug
priority: high
created_at: 2026-11-05T16:49:49Z
updated_at: 2026-11-05T16:49:49Z
---

Found while deploying ChatCMS to durer for the first time (quiqr/chatcms.eu
and the mipnix change of 5 November 2026).

An installation with nobody in it cannot make anybody. Accounts are created
from the administration screens, and those need somebody who administers the
installation. `services.chatcms.firstAdministrator` promotes an account that
already exists, and says so plainly in its own documentation, but nothing
creates that account:

- `cmd/seed` makes people, and also an organisation, two projects and a demo
  site. It is a development tool and has no business on a customer's server.
- There is no sign-up, by design.
- `docs/DEPLOYING.md` says "The account must already exist" without saying
  how it comes to exist.

The first deployment got round it with SQL: a row in `people` with no
password, a row in `credential_tokens` of kind `invitation`, and then the
`/set-password` link followed by hand. That works, and it is not something a
deployment should need.

What to decide: whether `chatcmsd` grows a subcommand that makes one account
and prints the link, or whether the module creates the person named in
`firstAdministrator` when the people table is empty. The second is less to
explain and harder to misuse, because it only fires on an installation that
has nobody at all.

Either way `docs/DEPLOYING.md` should stop ending that sentence where it does.
