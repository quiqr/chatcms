## Purpose

Defines how a password is stored and checked, and how an invitation or a
password reset is issued and spent, so that neither becomes a way in.

## ADDED Requirements

### Requirement: A password is stored as a slow hash, never recoverably

A password SHALL be stored only as an argon2id hash with a per-password salt.
Nothing SHALL be stored from which a password can be recovered.

#### Scenario: Setting a password

- **WHEN** a password is set
- **THEN** what is stored is a hash that names its algorithm and parameters
- **AND** two people with the same password have different stored values

#### Scenario: Checking a password

- **WHEN** a password is checked
- **THEN** the comparison does not reveal, by how long it takes, how much of
  the password was right

#### Scenario: A stored hash from different parameters

- **WHEN** a stored hash was made with different parameters from today's
- **THEN** it still verifies correctly

### Requirement: A password has a minimum length and no composition rules

Setting a password SHALL require a minimum length and SHALL refuse the most
common passwords. It SHALL NOT require particular kinds of character.

#### Scenario: Too short

- **WHEN** a password shorter than the minimum is set
- **THEN** it is refused, and the message says what the minimum is

#### Scenario: A very common password

- **WHEN** a password from the most common list is set
- **THEN** it is refused, and the message says it is too easy to guess

#### Scenario: A long passphrase of only lower-case letters

- **WHEN** a long passphrase with no digits or symbols is set
- **THEN** it is accepted

### Requirement: An invitation is a single-use link that expires

An invitation SHALL issue a token that sets the person's first password. It
SHALL be usable once, SHALL expire, and SHALL be stored as a hash.

#### Scenario: Accepting an invitation

- **WHEN** an invitation token is presented with a new password
- **THEN** the password is set and the token is spent
- **AND** the person can sign in

#### Scenario: Using it twice

- **WHEN** a spent invitation token is presented again
- **THEN** it is refused in plain language that says to ask for a new
  invitation

#### Scenario: Using it too late

- **WHEN** an expired invitation token is presented
- **THEN** it is refused in plain language that says it has expired

#### Scenario: A token nobody issued

- **WHEN** a token that was never issued is presented
- **THEN** it is refused the same way an expired one is

### Requirement: A password reset works the same way and says nothing

A reset SHALL issue the same kind of single-use expiring token. Asking for one
SHALL NOT reveal whether the address is known.

#### Scenario: Asking for a reset

- **WHEN** a reset is asked for, for a known or an unknown address
- **THEN** the answer is the same either way
- **AND** nothing in it would let a caller tell which

#### Scenario: Spending a reset

- **WHEN** a reset token is presented with a new password
- **THEN** the password is replaced and the token is spent

#### Scenario: Resetting ends other sessions

- **WHEN** a password is replaced through a reset
- **THEN** every session that person had is revoked

### Requirement: Guessing is rate limited

Sign-in and reset SHALL be limited both per address and per caller, so that
guessing is slow regardless of which is varied.

#### Scenario: Many attempts on one address

- **WHEN** sign-in is attempted repeatedly for one address
- **THEN** further attempts are refused for a time, with a message that does
  not say whether the address exists

#### Scenario: Many attempts from one caller

- **WHEN** sign-in is attempted repeatedly from one caller across many
  addresses
- **THEN** further attempts are refused for a time

#### Scenario: A correct sign-in is not punished

- **WHEN** someone signs in correctly after a few failures
- **THEN** it succeeds, and the count against them is cleared

### Requirement: A link is recorded where an implementer can read it

Until there is a way to send mail, an issued invitation or reset SHALL record
its link in the server log and SHALL NOT return it to the caller.

#### Scenario: The link is logged

- **WHEN** an invitation is issued
- **THEN** the link appears in the server log

#### Scenario: The link is not in the answer

- **WHEN** an invitation is issued
- **THEN** the response body carries no token and no link
