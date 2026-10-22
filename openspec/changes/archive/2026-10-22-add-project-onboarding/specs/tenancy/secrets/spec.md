## Purpose

Defines how a secret is kept, so that a copy of the database is worth nothing
and a secret never reaches a log, an answer or the Nix store.

## ADDED Requirements

### Requirement: A secret is encrypted before it is stored

Every secret SHALL be encrypted before being written, with a key the
application holds and the database does not.

#### Scenario: What is stored is not the secret

- **WHEN** a secret is stored and the stored rows are read directly
- **THEN** the secret does not appear in them

#### Scenario: The same secret stored twice differs

- **WHEN** the same secret is stored twice
- **THEN** the two stored values differ

#### Scenario: A stored secret comes back

- **WHEN** a stored secret is read back
- **THEN** it is exactly what was stored

#### Scenario: A stored secret that was tampered with

- **WHEN** a stored value is altered and then read back
- **THEN** reading fails rather than returning something

### Requirement: The master key comes from a file, not from configuration

The key that encrypts secrets SHALL be read from a file named by
configuration, so that it is never an environment variable, never in the
store that holds the program, and never in a log.

#### Scenario: The key is read from its file

- **WHEN** the application starts with a file holding a key
- **THEN** secrets can be stored and read

#### Scenario: No key configured

- **WHEN** the application starts with no key file
- **THEN** it refuses to start, naming what is missing

#### Scenario: A key that is too short

- **WHEN** the file holds fewer bytes than are needed
- **THEN** the application refuses to start and says so

#### Scenario: The key is not in what is logged

- **WHEN** anything is logged while secrets are used
- **THEN** the key does not appear in it

### Requirement: A private key is never given out

A stored private key SHALL NOT be returned by any endpoint, included in any
answer, or written to any log.

#### Scenario: Asking for a project

- **WHEN** a project is read through the API
- **THEN** its public key may be present and its private key is not

#### Scenario: Anything that fails while using the key

- **WHEN** an operation using the key fails and is reported
- **THEN** the report contains no part of the key
