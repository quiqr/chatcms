# identity/authentication Specification

## Purpose
Defines how someone proves who they are, how the system remembers it, and how
it stops remembering, so that every other capability can ask who is asking.

## Requirements

### Requirement: Signing in exchanges a password for a session

Signing in with a correct address and password SHALL produce a session. The
session SHALL be addressed by a cookie the page's own scripts cannot read, and
SHALL be stored so that a copy of the database does not yield usable sessions.

#### Scenario: Correct credentials

- **WHEN** someone signs in with the right address and password
- **THEN** a session is created and returned as a cookie
- **AND** the response says who they are

#### Scenario: The cookie is protected

- **WHEN** a session cookie is set
- **THEN** it is unreadable by page scripts, restricted to this site, and
  marked as requiring a secure connection whenever the site is served over one

#### Scenario: The stored session is not the cookie

- **WHEN** the stored sessions are read directly
- **THEN** none of them contains the value the cookie carries

#### Scenario: Wrong credentials

- **WHEN** the address is unknown, or the password is wrong
- **THEN** no session is created
- **AND** the answer is the same in both cases

### Requirement: Signing in never reuses an existing session

A sign-in SHALL always create a new session, and SHALL invalidate any session
the request arrived with.

#### Scenario: A session was already present

- **WHEN** someone signs in while carrying an existing session
- **THEN** the session they end up with is a different one
- **AND** the one they arrived with no longer works

### Requirement: Staying signed in is the person's choice

A sign-in SHALL offer to be remembered on the device. Without that, the
session SHALL last a working day; with it, long enough not to ask again in
ordinary use.

#### Scenario: Not remembered

- **WHEN** someone signs in without asking to be remembered
- **THEN** the session expires within a working day

#### Scenario: Remembered

- **WHEN** someone signs in asking to be remembered
- **THEN** the session lasts considerably longer
- **AND** the difference is visible in the session's own expiry, not only in
  the cookie

### Requirement: A request carries the person who made it

Every request SHALL carry either the person its session identifies or nobody,
and handlers SHALL read it from one place rather than inspecting cookies.

#### Scenario: A request with a valid session

- **WHEN** a request arrives with a valid session
- **THEN** the person it identifies is available to the handler

#### Scenario: A request with no session

- **WHEN** a request arrives with no session, an unknown one, an expired one
  or a revoked one
- **THEN** the handler sees nobody
- **AND** an endpoint that requires a person answers 401 with the JSON error
  shape

#### Scenario: Using a session keeps it alive

- **WHEN** a session is used
- **THEN** the time it was last seen is recorded
- **AND** recording it does not extend its expiry

### Requirement: Signing out ends the session everywhere

Signing out SHALL revoke the session in the store as well as clearing the
cookie, so that a copy of the cookie is useless afterwards.

#### Scenario: Signing out

- **WHEN** someone signs out
- **THEN** the cookie is cleared
- **AND** presenting the same session value again is refused

#### Scenario: Signing out without a session

- **WHEN** a sign-out arrives with no session
- **THEN** it succeeds rather than failing

### Requirement: What proves identity sits behind one interface

Checking whether someone is who they claim SHALL go through a single
interface, and nothing that reads a session SHALL depend on a password having
existed.

#### Scenario: A person with no password

- **WHEN** a person exists with no password set
- **THEN** sessions, sign-out and reading the current person all behave
  normally
- **AND** signing in with a password is refused

#### Scenario: Adding another way to prove identity

- **WHEN** a second implementation of the interface is added
- **THEN** nothing that reads a session has to change

### Requirement: Somebody can sign out from anywhere in the application

Every screen SHALL offer a way to sign out and a way to reach your own
account.

#### Scenario: From a site's screens

- **WHEN** somebody is looking at any screen of a project
- **THEN** signing out SHALL be offered
- **AND** their own account SHALL be reachable

#### Scenario: On a phone

- **WHEN** the phone layout is shown
- **THEN** both SHALL be offered there too

#### Scenario: After signing out

- **WHEN** somebody signs out
- **THEN** the session SHALL be ended
- **AND** the old session value SHALL be refused from then on
- **AND** the screen SHALL say that they signed out, which a session that
  merely ran out SHALL NOT say

### Requirement: Somebody can sign out everywhere

Somebody SHALL be able to end every session their account has, not only the
one they are using.

#### Scenario: Ending them all

- **WHEN** somebody signs out everywhere
- **THEN** every session that account holds SHALL be ended
- **AND** the one that asked SHALL be ended with them

#### Scenario: Somebody else's sessions

- **WHEN** somebody signs out everywhere
- **THEN** no other account's sessions SHALL be touched

#### Scenario: Not signed in

- **WHEN** nobody is signed in and signing out everywhere is asked for
- **THEN** it SHALL be refused
