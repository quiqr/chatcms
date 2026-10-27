---
# chatcms-r7im
title: Acting as somebody else, visibly
status: in-progress
type: epic
priority: normal
created_at: 2026-10-27T16:32:01Z
updated_at: 2026-10-27T21:30:00Z
parent: chatcms-8vem
blocked_by:
    - chatcms-294f
---

## Purpose

Support means answering "I cannot see the publish button", and the only way
to answer it today is to ask the customer to describe their screen. An
administrator needs to be able to see, and do, what that person sees and
does.

This is the one deliberate exception to the rule the administrator epic
established, that administering is not editing. It is allowed to be an
exception because it announces itself: on the screen while it happens, in
the record afterwards, and in the customer's own history forever.

## Decided

Settled rather than left open, because each of these shapes the work:

1. **An impersonator can do everything** the person they are acting as can
   do. Not read-only.
2. **Every version is attributed "Pim as Marieke".** History never says the
   customer did something an administrator did.
3. **The customer can see it happened**, in the history of their own site,
   not only in a record that administrators read.
4. **An administrator cannot be impersonated**, because that is a way of
   becoming one.

## Scope

- Beginning and ending it, from the person opened on the administration
  screen, and an exit from wherever it is in force.
- **The session stays the administrator's**, carrying who they are acting
  as, rather than issuing a session as the customer. Issuing one would lose
  who is really there, which is the whole trail.
- **Permissions are the impersonated person's.** `policy.Can` answers for
  them, which is what makes "can do everything" true without a second set of
  rules to keep in step.
- **What is recorded is the administrator's.** Every version, publish and
  record entry keeps both: who was really there and who they were acting as.
- Attribution reaches git: the author of a version reads as both, so that
  the trail survives ChatCMS and is in the customer's repository.
- A band across every screen while it is in force, saying who is being acted
  as and offering the way out, at 1440 px and at 390 px.
- Ending it when the administrator signs out, and when the session expires.
- Refusing to begin it on an administrator, and on an account that is off.

## Done when

- An administrator acts as a customer, edits and publishes, and the history
  of that site says "Pim as Marieke" where it would otherwise say Marieke,
  on the customer's own screen.
- The git repository says the same, so the trail is not only in ChatCMS.
- The band is on every screen while it is in force, and the way out works
  from each.
- Beginning and ending are both in the administration record, with who and
  when.
- Impersonating an administrator is refused, proven by an integration test.
- An editor who tries to begin it is refused by `policy.Can`, not by a
  hidden button.
- Signing out ends it.

## References

- `openspec/changes/archive/2026-10-26-add-application-administrator`, the
  rule this is the exception to, and the record it writes into.
- `internal/policy/policy.go`, which decides, and the test that administering
  grants nothing on a project.
- `internal/gitsync/runner.go`, `Identity`, which is what a version carries.
