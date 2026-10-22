## ADDED Requirements

### Requirement: A request that changes something proves it was meant

A request that changes anything SHALL present a token that only pages served
by this site can read, so that another site cannot make a reader's browser act
on their behalf.

#### Scenario: A change without proof

- **WHEN** a state-changing request arrives with a session but no matching
  token
- **THEN** it is refused with 403 and the JSON error shape
- **AND** nothing is changed

#### Scenario: A change with proof

- **WHEN** the request presents the token matching the one it was given
- **THEN** it proceeds

#### Scenario: Reading needs no proof

- **WHEN** a request only reads
- **THEN** it needs no such token

#### Scenario: Signing in needs no session but still needs proof

- **WHEN** a sign-in request arrives
- **THEN** it needs the token, because it changes something
- **AND** it does not need an existing session
