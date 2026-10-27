---
# chatcms-294f
title: 'Administration: every person, and what they can reach'
status: in-progress
type: epic
priority: high
created_at: 2026-10-27T15:17:33Z
updated_at: 2026-10-27T19:45:00Z
parent: chatcms-8vem
---

## Purpose

An administrator can see every organisation and nobody at all. There is no
list of the people on an installation, so the question "who is this address,
and what can they reach" can only be answered with a database client. Only
administrators can be listed, which is a different question.

## Scope

- Every person on the installation: their name, their address, whether their
  account is on, and when they were last here. `People.LastSeen` already
  answers the last of those.
- What each person can reach: the organisations they belong to and what they
  hold there, and the projects they have a role on. One answer, because
  three requests per person is how a list of fifty becomes a hundred and
  fifty.
- Finding somebody by name or address, because the list grows with every
  customer.
- Opening one person: the same facts, plus turning their account off, which
  exists and is only reachable by typing an address into a box.
- Making a person. An address with no account can already be invited to an
  organisation or a project; this is the same act without having to pick one
  first, and it sends the same way in.

## Done when

- Every person is listed with where they belong and when they were last
  here, at 1440 px and at 390 px.
- Opening one shows everything they can reach, and nothing about anybody
  else.
- Somebody who does not administer the installation is refused the list and
  each person on it, proven by an integration test.
- Making a person sends them the way in, and does not set a password.
- The list answers in one request per page rather than one per person.

## References

- `openspec/changes/archive/2026-10-26-add-application-administrator`
- `internal/auth/identity.go`, `People.LastSeen`.
